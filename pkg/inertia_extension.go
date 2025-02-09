package pkg

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/romsar/gonertia"
	"html/template"
	"io/fs"
	"log/slog"
	"net/url"
	"os/exec"
	"path"
	"reflect"
	"strings"
)

// DefaultIgnoreList is the default list of paths that will be ignored by the dev proxy
var DefaultIgnoreList = []string{"api", "swagger"}

//go:embed inertia_root.gohtml
var DefaultInertiaRootTemplate []byte

// InertiaExtensionOption is a function that sets an option on the InertiaExtension
type InertiaExtensionOption func(i *InertiaExtension) error

// WithIgnoreList sets the ignore list on the InertiaExtension
func WithIgnoreList(ignoreList []string) InertiaExtensionOption {
	return func(i *InertiaExtension) error {
		i.ignoreList = ignoreList
		return nil
	}
}

// WithFrontendPath sets the frontend path on the InertiaExtension
func WithFrontendPath(frontendPath string) InertiaExtensionOption {
	return func(i *InertiaExtension) error {
		i.frontendPath = frontendPath
		return nil
	}
}

// WithJSRuntime sets the JS runtime on the InertiaExtension
func WithJSRuntime(jsRuntime string) InertiaExtensionOption {
	return func(i *InertiaExtension) error {
		i.jsRuntime = jsRuntime
		return nil
	}
}

// WithDevServerURL sets the dev server URL on the InertiaExtension
func WithDevServerURL(devServerURL string) InertiaExtensionOption {
	return func(i *InertiaExtension) error {
		i.devServerURL = devServerURL
		return nil
	}
}

// WithRootTemplate sets the root template on the InertiaExtension
func WithRootTemplate(rootTemplate []byte) InertiaExtensionOption {
	return func(i *InertiaExtension) error {
		i.rootTemplate = rootTemplate
		return nil
	}
}

// NewInertiaExtension creates a new InertiaExtension
func NewInertiaExtension(distDirFS fs.FS, manifest []byte, isDev bool, opts ...InertiaExtensionOption) (*InertiaExtension, error) {
	ext := &InertiaExtension{
		rootTemplate: DefaultInertiaRootTemplate,
		manifest:     manifest,
		isDev:        isDev,
		ignoreList:   DefaultIgnoreList,
		frontendPath: "./frontend",
		jsRuntime:    "bun",
		devServerURL: "http://localhost:5173/",
		distDirFS:    distDirFS,
	}

	for _, opt := range opts {
		err := opt(ext)
		if err != nil {
			return nil, err
		}
	}

	return ext, nil
}

// InertiaExtension is an extension that provides Inertia support to the application
type InertiaExtension struct {
	// Inertia is the Inertia instance that will be used by the InertiaExtension
	Inertia *gonertia.Inertia
	// distDirFS is the dist directory that will be used by the InertiaExtension to serve the assets
	distDirFS fs.FS
	// rootTemplate is the root template for inertia, this is set in the NewInertiaExtension function, it is used to create the Inertia instance.
	rootTemplate []byte
	// manifest is the manifest file that is used to get the vite assets, this is set in the NewInertiaExtension function
	manifest []byte
	// logger is a logger that will be used by the InertiaExtension to log messages, this is set in the Register function
	logger *Logger
	// flashExtension is the flash extension that will be used by the InertiaExtension to create a flash provider, this is set in the Register function
	flashExtension *FlashExtension
	// isDev is a boolean that is set to true if the environment is dev, this is set in the Register function
	isDev bool
	// ignoreList is a list of paths that will be ignored by the dev proxy.
	ignoreList []string
	// frontendPath is the path to the frontend directory in the dir from the root of the project, this is used to start the dev server
	frontendPath string
	// jsRuntime is the path to the JS runtime that will be used by the InertiaExtension to run JS code.
	jsRuntime string
	// devServerURL is the URL of the dev server that will be used by the InertiaExtension to proxy requests to the dev server
	devServerURL string
}

// createHash creates a hash from the root template
func (i *InertiaExtension) createHash() string {
	hash := sha256.New()
	hash.Write(i.rootTemplate)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Register registers the InertiaExtension
func (i *InertiaExtension) Register(app *Bat) error {
	app.Logger.Info("Instantiating Inertia")
	i.flashExtension = GetExtension[*FlashExtension](app)
	i.logger = &Logger{app.Logger.With(slog.String("module", "inertia"))}
	var err error
	i.Inertia, err = gonertia.NewFromBytes(
		i.rootTemplate,
		gonertia.WithVersion(i.createHash()),
		gonertia.WithLogger(i.logger),
		gonertia.WithFlashProvider(i.flashExtension))
	if err != nil {
		i.logger.Error("Failed to initialize Inertia", slog.Any("err", err))
		return err
	}
	err = i.Inertia.ShareTemplateFunc("vite", i.vite("build"))
	if err != nil {
		return err
	}
	err = i.Inertia.ShareTemplateFunc("viteHead", i.viteHead())
	if err != nil {
		return err
	}

	err = i.Inertia.ShareTemplateFunc("reactRefresh", i.reactRefresh())
	if err != nil {
		return err
	}

	app.Echo.Use(echo.WrapMiddleware(i.Inertia.Middleware))

	if i.isDev {
		i.logger.Info("Setting up dev proxy")
		err := i.setupDevProxy(app)
		if err != nil {
			i.logger.Error("Failed to setup dev proxy", err)
			return err
		}
		return nil
	}

	app.Echo.StaticFS("/build", i.distDirFS)
	return nil
}

// Requirements returns the requirements for the InertiaExtension
func (i *InertiaExtension) Requirements() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf(SessionExtension{}),
		reflect.TypeOf(FlashExtension{}),
	}
}

// vite is a helper function that returns a function that returns the vite asset path
func (i *InertiaExtension) vite(buildDir string) func(path string) (string, error) {
	viteAssets := make(map[string]*struct {
		File   string `json:"file"`
		Source string `json:"src"`
	})
	err := json.Unmarshal(i.manifest, &viteAssets)
	if err != nil {
		i.logger.Error("Failed to unmarshal vite manifest file to json", err)
	}

	return func(p string) (string, error) {
		// If in dev mode and the asset is in the viteAssets map, return the vite asset path
		if i.isDev {
			if _, ok := viteAssets[p]; ok {
				return path.Join("/", p), nil
			}
		}
		// If in prod mode and the asset is in the viteAssets map, return the dist asset path
		if val, ok := viteAssets[p]; ok {
			return path.Join("/", buildDir, val.File), nil
		}
		return "", fmt.Errorf("asset %q not found", p)
	}
}

// reactRefresh is a helper function that returns a function that returns the react refresh script
func (i *InertiaExtension) reactRefresh() func() template.HTML {
	return func() template.HTML {
		if i.isDev {
			return "<script type=\"module\">import RefreshRuntime from '/@react-refresh';" +
				"RefreshRuntime.injectIntoGlobalHook(window);" +
				"window.$RefreshReg$ = () => { };" +
				"window.$RefreshSig$ = () => (type) => type;" +
				"window.__vite_plugin_react_preamble_installed__ = true;</script>"
		} else {
			return ""
		}
	}
}

// viteHead is a helper function that returns a function that returns the vite head script
func (i *InertiaExtension) viteHead() func() template.HTML {
	return func() template.HTML {
		if i.isDev {
			return "<script type=\"module\" src=\"/@vite/client\"></script>"
		} else {
			return ""
		}
	}
}

// setupDevProxy sets up a proxy to the vite dev server
func (i *InertiaExtension) setupDevProxy(bat *Bat) error {
	cmd := exec.Command(i.jsRuntime, "run", "dev")
	cmd.Dir = i.frontendPath
	err := cmd.Start()
	if err != nil {
		i.logger.Error("Failed to start the dev server", err)
	}

	url, err := url.Parse(i.devServerURL)
	if err != nil {
		i.logger.Error("Failed to parse the URL for the dev server", err, url)
		return err
	}
	// Setup a proxy to the vite dev server on localhost:5173
	balancer := middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
		{
			URL: url,
		},
	})

	px := middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: balancer,
		Skipper: func(c echo.Context) bool {
			// Skip the proxy if the path is in the ignore list
			for _, ignorePath := range i.ignoreList {
				if strings.Contains(c.Path(), ignorePath) {
					return true
				}
			}

			return false
		},
	})

	bat.Group("/src").Use(px)
	bat.Group("/@*").Use(px)
	bat.Group("/node_modules").Use(px)

	return nil
}

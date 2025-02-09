package pkg

import (
	"github.com/labstack/echo/v4"
	slogecho "github.com/samber/slog-echo"
	"log/slog"
	"reflect"
)

var skipPaths = []string{"/src", "/@*", "/node_modules", "/build/", "/@vite", "/@react-refresh"}

type BatOption func(*Bat) error

type Bat struct {
	Logger *Logger
	*echo.Echo
	extensions map[reflect.Type]interface{}
	// SkipPaths is a list of paths that should not be logged
	SkipPaths []string
}

func (b *Bat) RegisterControllers(controllers ...Controller) error {
	for _, controller := range controllers {
		if err := controller.Register(b); err != nil {
			return err
		}
	}
	return nil
}
func NewBat(logger *Logger, extensions ...Extension) (*Bat, error) {
	bat := &Bat{
		Logger:     logger,
		Echo:       echo.New(),
		extensions: make(map[reflect.Type]interface{}),
		SkipPaths:  skipPaths,
	}

	err := bat.registerExtensions(extensions...)
	if err != nil {
		return nil, err
	}

	bat.HideBanner = true
	bat.HidePort = true
	bat.Use(slogecho.NewWithFilters(logger.With(slog.String("module", "echo")), slogecho.IgnorePathContains(
		bat.SkipPaths...)))

	return bat, nil
}

func (b *Bat) Start(addr string) error {
	b.Logger.Info("Starting server", slog.String("address", addr))
	err := b.Echo.Start(addr)
	if err != nil {
		b.Logger.Error("Failed to start server", slog.String("error", err.Error()))
		return err
	}
	return nil
}

package internal

import "github.com/JensvandeWiel/go-bat/internal/templates/inertia_react_extra"

type InertiaReactExtra struct {
}

func NewInertiaReactExtra() *InertiaReactExtra {
	return &InertiaReactExtra{}
}

func (i *InertiaReactExtra) Generate(project *Project) error {
	err := project.writeStringTemplateToFile("controllers/inertia_controller.go", inertia_react_extra.ControllersInertiaController, nil)
	if err != nil {
		return err
	}

	err = project.copyEmbeddedFiles(inertia_react_extra.Frontend, "frontend", "frontend", func(s string) string {
		if s == "frontend.go.tmpl" {
			return "frontend.go"
		} else if s == "gitignore.tmpl" {
			return ".gitignore"
		}
		return s
	})
	if err != nil {
		return err
	}

	return nil
}

func (i *InertiaReactExtra) ModEntries() []string {
	return []string{
		"github.com/romsar/gonertia/v2 v2.0.3",
		"github.com/valkey-io/valkey-go v1.0.54",
	}
}

func (i *InertiaReactExtra) GitIgnoreEntries() []string {
	return []string{}
}

func (i *InertiaReactExtra) GetExtraPersistentFlags() []string {
	return []string{
		"rootCmd.PersistentFlags().String(\"CACHE_HOST\", \"localhost\", \"the cache host\")",
		"rootCmd.PersistentFlags().Int(\"CACHE_PORT\", 6379, \"the cache port\")",
		"viper.BindPFlag(\"CACHE_HOST\", rootCmd.PersistentFlags().Lookup(\"CACHE_HOST\"))",
		"viper.BindPFlag(\"CACHE_PORT\", rootCmd.PersistentFlags().Lookup(\"CACHE_PORT\"))",
	}
}

func (i *InertiaReactExtra) ExtraType() ExtraType {
	return InertiaReact
}

func (i *InertiaReactExtra) DisallowedExtraTypes() []ExtraType {
	return []ExtraType{InertiaSvelte}
}

func (i *InertiaReactExtra) ComposerServices() []string {
	return []string{`  valkey:
    image: valkey/valkey:8
    ports:
      - "6379:6379"
    volumes:
      - valkey_data:/data
`}
}

func (i *InertiaReactExtra) ComposerVolumes() []string {
	return []string{`  valkey_data:`}
}

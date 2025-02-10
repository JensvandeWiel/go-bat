package internal

import (
	"github.com/JensvandeWiel/go-bat/internal/templates/inertia_svelte_extra"
)

type InertiaSvelteExtra struct {
}

func NewInertiaSvelteExtra() *InertiaSvelteExtra {
	return &InertiaSvelteExtra{}
}

func (i *InertiaSvelteExtra) Generate(project *Project) error {
	err := project.writeStringTemplateToFile("controllers/inertia_controller.go", inertia_svelte_extra.ControllersInertiaController, nil)
	if err != nil {
		return err
	}

	err = project.copyEmbeddedFiles(inertia_svelte_extra.Frontend, "frontend", "frontend", func(s string) string {
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

func (i *InertiaSvelteExtra) ModEntries() []string {
	return []string{
		"github.com/romsar/gonertia/v2 v2.0.3",
		"github.com/valkey-io/valkey-go v1.0.54",
	}
}

func (i *InertiaSvelteExtra) GitIgnoreEntries() []string {
	return []string{}
}

func (i *InertiaSvelteExtra) GetExtraPersistentFlags() []string {
	return []string{
		"rootCmd.PersistentFlags().String(\"CACHE_HOST\", \"localhost\", \"the cache host\")",
		"rootCmd.PersistentFlags().Int(\"CACHE_PORT\", 6379, \"the cache port\")",
		"viper.BindPFlag(\"CACHE_HOST\", rootCmd.PersistentFlags().Lookup(\"CACHE_HOST\"))",
		"viper.BindPFlag(\"CACHE_PORT\", rootCmd.PersistentFlags().Lookup(\"CACHE_PORT\"))",
	}
}

func (i *InertiaSvelteExtra) ExtraType() ExtraType {
	return InertiaSvelte
}

func (i *InertiaSvelteExtra) DisallowedExtraTypes() []ExtraType {
	return []ExtraType{InertiaReact}
}

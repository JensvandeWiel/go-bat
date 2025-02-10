package internal

import (
	"fmt"
	bat "github.com/JensvandeWiel/go-bat/pkg"
	"html/template"
	"os"
	"path"
)

type Generator interface {
	Generate(name string) error
}

func ParseGenerator(name string, logger *bat.Logger, dir string) (Generator, error) {
	logger.Debug("Parsing generator", "name", name)
	switch name {
	case "model":
		return &ModelGenerator{
			logger: logger,
			dir:    dir,
		}, nil
	default:
		logger.Error("Unknown generator", "name", name)
		return nil, fmt.Errorf("unknown generator %q", name)
	}
}

func WriteStringTemplateToFile(filePath string, tmplString string, data interface{}) error {
	dir := path.Dir(filePath)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	tmpl, err := template.New(path.Base(filePath)).Parse(tmplString)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path.Join(filePath), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	err = file.Truncate(0)
	if err != nil {
		return err
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}

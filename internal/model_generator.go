package internal

import (
	"fmt"
	"github.com/JensvandeWiel/go-bat/internal/templates/generators"
	bat "github.com/JensvandeWiel/go-bat/pkg"
	"github.com/iancoleman/strcase"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ModelGenerator struct {
	logger *bat.Logger
	dir    string
}

func (m *ModelGenerator) Generate(name string) error {
	modelDir := path.Join(m.dir, "database", "models")
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		m.logger.Error("model directory does not exist, database extra is possibly not in use", "modelDir", modelDir)
		return fmt.Errorf("model directory does not exist: %s", modelDir)
	}

	modelFile := filepath.Join(modelDir, strings.ToLower(name)+".go")
	if _, err := os.Stat(modelFile); !os.IsNotExist(err) {
		m.logger.Error("model file already exists", "modelFile", modelFile)
		return fmt.Errorf("model file already exists: %s", modelFile)
	}

	data := struct {
		Name string
	}{
		Name: strcase.ToCamel(name),
	}

	err := WriteStringTemplateToFile(modelFile, generators.ModelTemplate, data)
	if err != nil {
		m.logger.Error("failed to write model file", "modelFile", modelFile, "error", err)
		return err
	}

	m.logger.Info("Generated model", "modelFile", modelFile)
	return nil
}

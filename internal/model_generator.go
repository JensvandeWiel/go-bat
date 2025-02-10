package internal

import (
	"fmt"
	"github.com/JensvandeWiel/go-bat/internal/templates/generators"
	"github.com/iancoleman/strcase"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ModelGenerator struct {
	project *Project
	extra   bool
}

func (m *ModelGenerator) Generate(name string, extra bool) error {
	if !m.project.ExtraTypes.HasExtra(DatabasePgSQL) {
		m.project.logger.Error("database extra is not in use")
		return fmt.Errorf("database extra is not in use")
	}

	modelDir := path.Join("database", "models")
	if _, err := os.Stat(path.Join(m.project.tempDir, modelDir)); os.IsNotExist(err) {
		m.project.logger.Error("model directory does not exist, database extra is possibly not in use", "modelDir", modelDir)
		return fmt.Errorf("model directory does not exist: %s", modelDir)
	}

	modelFile := filepath.Join(modelDir, strings.ToLower(name)+".go")
	if _, err := os.Stat(filepath.Join(m.project.tempDir, modelFile)); !os.IsNotExist(err) {
		m.project.logger.Error("model file already exists", "modelFile", modelFile)
		return fmt.Errorf("model file already exists: %s", modelFile)
	}

	data := struct {
		Name string
	}{
		Name: strcase.ToCamel(name),
	}

	err := m.project.writeStringTemplateToFile(modelFile, generators.ModelTemplate, data)
	if err != nil {
		m.project.logger.Error("failed to write model file", "modelFile", modelFile, "error", err)
		return err
	}

	m.project.logger.Info("Generated model", "modelFile", modelFile)
	return nil
}

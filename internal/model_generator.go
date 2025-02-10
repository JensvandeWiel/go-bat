package internal

import (
	"fmt"
	"github.com/JensvandeWiel/go-bat/internal/templates/generators"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type ModelGenerator struct {
	project *Project
	extra   bool
}

var timestampFormat = "20060102150405"

func generateTimestamp() string {
	return time.Now().UTC().Format(timestampFormat)
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

	if extra {
		m.project.logger.Info("Generating extra files for model")

		data := map[string]interface{}{
			"pluralLowName": pluralize.NewClient().Plural(strings.ToLower(name)),
			"PackageName":   m.project.PackageName,
			"modelName":     strcase.ToCamel(name),
			"modelNameLow":  strings.ToLower(name),
		}
		fileName := generateTimestamp() + "_create_" + strings.ToLower(name) + ".sql"
		err := m.project.writeStringTemplateToFile(filepath.Join("database", "migrations", fileName), generators.ModelMigrationTemplate, data)
		if err != nil {
			return err
		}

		err = m.project.writeStringTemplateToFile(filepath.Join("database", "stores", strings.ToLower(name)+"_store.go"), generators.ModelStoreTemplate, data)
		if err != nil {
			return err
		}

		err = m.project.writeStringTemplateToFile(filepath.Join("database", "stores", strings.ToLower(name)+"_store_test.go"), generators.ModelStoreTestTemplate, data)
		if err != nil {
			return err
		}
		m.project.logger.Info("Generated extra files for model")
	}

	m.project.logger.Info("Generated model", "modelFile", modelFile)
	return nil
}

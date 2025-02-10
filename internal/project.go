package internal

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JensvandeWiel/go-bat/pkg"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

const ProjectFile = "project.json"

type Project struct {
	logger      *pkg.Logger
	WorkDir     string     `json:"-"`
	ProjectName string     `json:"project_name"`
	PackageName string     `json:"package_name"`
	Force       bool       `json:"-"`
	Extras      []Extra    `json:"-"`
	ExtraTypes  ExtraTypes `json:"extras"`
	funcMap     template.FuncMap
	tempDir     string
}

func NewProjectFromConfig(dir string, logger *pkg.Logger) (*Project, error) {
	logger.Debug("Loading project config", "dir", dir)
	var p Project
	file, err := os.ReadFile(path.Join(dir, ProjectFile))
	if err != nil {
		logger.Error("Failed to read project config", "error", err)
		return nil, err
	}

	err = json.Unmarshal(file, &p)
	if err != nil {
		logger.Error("Failed to unmarshal project config", "error", err)
		return nil, err
	}

	p.logger = logger
	p.WorkDir = dir
	p.tempDir = dir
	return &p, nil
}

func NewProject(projectName, packageName, workDir string, force bool, logger *pkg.Logger, extras ...Extra) (*Project, error) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "project-*")
	if err != nil {
		return nil, err
	}

	extraTypes := make([]ExtraType, 0, len(extras))
	for _, extra := range extras {
		extraTypes = append(extraTypes, extra.ExtraType())
	}

	p := &Project{
		ProjectName: projectName,
		PackageName: packageName,
		Extras:      extras,
		ExtraTypes:  extraTypes,
		WorkDir:     workDir,
		tempDir:     tempDir,
		logger:      logger,
		funcMap:     make(template.FuncMap),
		Force:       force,
	}

	funcMap := make(template.FuncMap)
	funcMap["getExtraModEntries"] = func() string {
		modEntries := ""
		for _, extra := range p.Extras {
			for _, entry := range extra.ModEntries() {
				modEntries += entry + "\n\t"
			}
		}

		return modEntries
	}

	p.funcMap = funcMap

	funcMap["getExtraGitIgnoreEntries"] = func() string {
		gitIgnoreEntries := ""
		for _, extra := range p.Extras {
			for _, entry := range extra.GitIgnoreEntries() {
				gitIgnoreEntries += entry + "\n"
			}
		}

		return gitIgnoreEntries
	}

	funcMap["getExtraPersistentFlags"] = func() string {
		flags := ""
		for _, extra := range p.Extras {
			for _, entry := range extra.GetExtraPersistentFlags() {
				flags += entry + "\n\t"
			}
		}
		return flags
	}

	funcMap["isExtraEnabled"] = func(extra string) bool {
		for _, e := range p.Extras {
			if e.ExtraType().String() == extra {
				return true
			}
		}
		return false
	}

	funcMap["hasComposeFile"] = func() bool {
		for _, extra := range p.Extras {
			if len(extra.ComposerServices()) > 0 || len(extra.ComposerVolumes()) > 0 {
				return true
			}
		}
		return false
	}
	return p, nil
}

func (p *Project) Create() error {
	err := p.checkExtraIncompatibilities()
	if err != nil {
		return err
	}

	err = p.GenerateBase()
	if err != nil {
		return err
	}

	for _, extra := range p.Extras {
		err = extra.Generate(p)
		if err != nil {
			return err
		}
	}

	err = p.generateComposerFile()
	if err != nil {
		return err
	}

	err = p.SaveConfig()
	if err != nil {
		return err
	}

	// Copy tempdir to project dir
	err = p.moveToProjectDir()
	if err != nil {
		return err
	}

	return nil
}

func (p *Project) SaveConfig() error {
	p.logger.Debug("Saving project config")
	marshalledJson, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		p.logger.Error("Failed to marshal project config", "error", err)
		return err
	}

	err = os.WriteFile(path.Join(p.tempDir, ProjectFile), marshalledJson, os.ModePerm)
	if err != nil {
		p.logger.Error("Failed to write project config", "error", err)
		return err
	}

	return nil
}

// createDirectories creates directories in the project
func (p *Project) createDirectories(dirs []string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(p.tempDir, dir), 0755); err != nil {
			return err
		}
	}
	return nil
}

// writeStringTemplateToFile writes a string template to a file, filePath should be relative to the project path
func (p *Project) writeStringTemplateToFile(filePath string, tmplString string, data interface{}) error {
	p.logger.Debug("Writing template to file", "path", filePath)
	dir := path.Dir(filePath)
	err := p.createDirectories([]string{dir})
	if err != nil {
		return err
	}

	tmpl, err := template.New(path.Base(filePath)).Funcs(p.funcMap).Parse(tmplString)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path.Join(p.tempDir, filePath), os.O_CREATE|os.O_WRONLY, os.ModePerm)
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

// moveToProjectDir moves the contents of the tempdir to the project directory
func (p *Project) moveToProjectDir() error {
	// Check if the project directory exists already
	if _, err := os.Stat(path.Join(p.WorkDir, p.ProjectName)); !os.IsNotExist(err) && !p.Force {
		p.logger.Debug("Project directory already exists, stopping generation")
		return errors.New("project directory already exists")
	} else {
		return p.copyDir(p.tempDir, path.Join(p.WorkDir, p.ProjectName))
	}
}

// copyDir copies the contents of the source directory to the destination directory
func (p *Project) copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return p.copyFile(path, dstPath)
	})
}

// copyFile copies a file from src to dst
func (p *Project) copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	stat, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return dstFile.Chmod(stat.Mode())
}

// getExtraModEntries returns the go.mod entries for the extras
func (p *Project) getExtraModEntries() []string {
	var entries []string
	for _, extra := range p.Extras {
		entries = append(entries, extra.ModEntries()...)
	}
	return entries
}

// getExtraGitIgnoreEntries returns the .gitignore entries for the extras
func (p *Project) getExtraGitIgnoreEntries() []string {
	var entries []string
	for _, extra := range p.Extras {
		entries = append(entries, extra.GitIgnoreEntries()...)
	}
	return entries
}

// copyEmbeddedFiles copies files from the embedded filesystem to the project directory
func (p *Project) copyEmbeddedFiles(efs embed.FS, srcDir, destDir string, renameFunc func(string) string) error {
	return fs.WalkDir(efs, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Apply the rename function to the base name of the file
		baseName := filepath.Base(relPath)
		newBaseName := renameFunc(baseName)
		destPath := filepath.Join(p.tempDir, destDir, filepath.Dir(relPath), newBaseName)

		if d.IsDir() {
			return os.MkdirAll(destPath, os.ModePerm)
		}

		data, err := efs.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, os.ModePerm)
	})
}

// checkExtraIncompatibilities checks if the extras are compatible with each other
func (p *Project) checkExtraIncompatibilities() error {
	for _, extra := range p.Extras {
		for _, disallowed := range extra.DisallowedExtraTypes() {
			for _, otherExtra := range p.Extras {
				if otherExtra.ExtraType() == disallowed {
					return fmt.Errorf("extra %s is incompatible with extra %s", extra.ExtraType(), otherExtra.ExtraType())
				}
			}
		}
	}
	return nil
}

func (p *Project) generateComposerFile() error {
	file := "services:\n"
	for _, extra := range p.Extras {
		for _, service := range extra.ComposerServices() {
			file += service + "\n"
		}
	}
	file += "volumes:\n"
	for _, extra := range p.Extras {
		for _, volume := range extra.ComposerVolumes() {
			file += volume + "\n"
		}
	}

	if file == "services:\nvolumes:\n" {
		return nil
	}

	err := os.WriteFile(path.Join(p.tempDir, "docker-compose.yml"), []byte(file), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

package internal

import (
	"errors"
	"github.com/JensvandeWiel/go-bat/pkg"
	"io"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

type Project struct {
	logger      *pkg.Logger
	WorkDir     string
	ProjectName string
	PackageName string
	Force       bool
	Extras      []Extra
	funcMap     template.FuncMap
	tempDir     string
}

func NewProject(projectName, packageName, workDir string, force bool, logger *pkg.Logger, extras ...Extra) (*Project, error) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "project-*")
	if err != nil {
		return nil, err
	}

	funcMap := make(template.FuncMap)
	funcMap["getExtraModEntries"] = func(p *Project) string {
		modEntries := ""
		for _, extra := range p.Extras {
			for _, entry := range extra.ModEntries() {
				modEntries += entry + "\n"
			}
		}

		return modEntries
	}

	funcMap["getExtraGitIgnoreEntries"] = func(p *Project) string {
		gitIgnoreEntries := ""
		for _, extra := range p.Extras {
			for _, entry := range extra.GitIgnoreEntries() {
				gitIgnoreEntries += entry + "\n"
			}
		}

		return gitIgnoreEntries
	}

	return &Project{
		ProjectName: projectName,
		PackageName: packageName,
		Extras:      extras,
		WorkDir:     workDir,
		tempDir:     tempDir,
		logger:      logger,
		funcMap:     make(template.FuncMap),
		Force:       force,
	}, nil
}

func (p *Project) Create() error {
	err := p.GenerateBase()
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

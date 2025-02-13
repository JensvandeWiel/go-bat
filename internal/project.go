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
	"os/exec"
	"path"
	"path/filepath"
	"text/template"
)

const ProjectFile = "project.json"

type Project struct {
	logger        *pkg.Logger
	WorkDir       string     `json:"-"`
	ProjectName   string     `json:"project_name"`
	PackageName   string     `json:"package_name"`
	Force         bool       `json:"-"`
	Extras        []Extra    `json:"-"`
	ExtraTypes    ExtraTypes `json:"extras"`
	funcMap       template.FuncMap
	tempDir       string
	noInstallDeps bool
	noGit         bool
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

func NewProject(projectName, packageName, workDir string, force bool, noInstallDeps, noGit bool, logger *pkg.Logger, extras ...Extra) (*Project, error) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "project-*")
	if err != nil {
		return nil, err
	}

	extraTypes := make([]ExtraType, 0, len(extras))
	for _, extra := range extras {
		extraTypes = append(extraTypes, extra.ExtraType())
	}

	p := &Project{
		ProjectName:   projectName,
		PackageName:   packageName,
		Extras:        extras,
		ExtraTypes:    extraTypes,
		WorkDir:       workDir,
		tempDir:       tempDir,
		logger:        logger,
		funcMap:       make(template.FuncMap),
		Force:         force,
		noInstallDeps: noInstallDeps,
		noGit:         noGit,
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

	if !p.noInstallDeps {
		err = p.installDeps()
		if err != nil {
			return err
		}
	}

	if !p.noGit {
		err = p.initGit()
		if err != nil {
			return err
		}
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
		p.logger.Debug("Creating directory", "path", dir)
		if err := os.MkdirAll(filepath.Join(p.tempDir, dir), 0755); err != nil {
			p.logger.Error("Failed to create directory", "path", dir, "error", err)
			return err
		}
		p.logger.Debug("Directory created", "path", dir)
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
	p.logger.Debug("Template written to file", "path", filePath)
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
	p.logger.Debug("Copying file", "src", src, "dst", dst)
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

	err = dstFile.Chmod(stat.Mode())
	if err != nil {
		return err
	}

	p.logger.Debug("File copied", "src", src, "dst", dst)
	return nil
}

// getExtraModEntries returns the go.mod entries for the extras
func (p *Project) getExtraModEntries() []string {
	p.logger.Debug("Getting all extra mod entries")
	var entries []string
	for _, extra := range p.Extras {
		p.logger.Debug("Getting extra mod entries", "extra", extra.ExtraType())
		entries = append(entries, extra.ModEntries()...)
		p.logger.Debug("Got extra mod entries", "extra", extra.ExtraType())
	}
	p.logger.Debug("Got all extra mod entries", "entries", entries)
	return entries
}

// getExtraGitIgnoreEntries returns the .gitignore entries for the extras
func (p *Project) getExtraGitIgnoreEntries() []string {
	p.logger.Debug("Getting all extra gitignore entries")
	var entries []string
	for _, extra := range p.Extras {
		p.logger.Debug("Getting extra gitignore entries", "extra", extra.ExtraType())
		entries = append(entries, extra.GitIgnoreEntries()...)
		p.logger.Debug("Got extra gitignore entries", "extra", extra.ExtraType())
	}
	p.logger.Debug("Got all extra gitignore entries", "entries", entries)
	return entries
}

// copyEmbeddedFiles copies files from the embedded filesystem to the project directory
func (p *Project) copyEmbeddedFiles(efs embed.FS, srcDir, destDir string, renameFunc func(string) string) error {
	p.logger.Debug("Starting to copy embedded files", "srcDir", srcDir, "destDir", destDir)
	err := fs.WalkDir(efs, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			p.logger.Error("Error walking directory", "path", path, "error", err)
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			p.logger.Error("Error getting relative path", "srcDir", srcDir, "path", path, "error", err)
			return err
		}

		// Apply the rename function to the base name of the file
		baseName := filepath.Base(relPath)
		newBaseName := renameFunc(baseName)
		destPath := filepath.Join(p.tempDir, destDir, filepath.Dir(relPath), newBaseName)

		if d.IsDir() {
			p.logger.Debug("Creating directory", "path", destPath)
			return os.MkdirAll(destPath, os.ModePerm)
		}

		p.logger.Debug("Copying file", "src", path, "dest", destPath)
		data, err := efs.ReadFile(path)
		if err != nil {
			p.logger.Error("Error reading file", "path", path, "error", err)
			return err
		}

		err = os.WriteFile(destPath, data, os.ModePerm)
		if err != nil {
			p.logger.Error("Error writing file", "path", destPath, "error", err)
			return err
		}

		p.logger.Debug("File copied", "src", path, "dest", destPath)
		return nil
	})

	if err != nil {
		p.logger.Error("Error copying embedded files", "error", err)
		return err
	}

	p.logger.Debug("Finished copying embedded files", "srcDir", srcDir, "destDir", destDir)
	return nil
}

// checkExtraIncompatibilities checks if the extras are compatible with each other
func (p *Project) checkExtraIncompatibilities() error {
	p.logger.Debug("Checking extra incompatibilities")
	for _, extra := range p.Extras {
		for _, disallowed := range extra.DisallowedExtraTypes() {
			for _, otherExtra := range p.Extras {
				if otherExtra.ExtraType() == disallowed {
					return fmt.Errorf("extra %s is incompatible with extra %s", extra.ExtraType(), otherExtra.ExtraType())
				}
			}
		}
	}

	// Check if the required extras are present
	for _, extra := range p.Extras {
		for _, required := range extra.RequiredExtraTypes() {
			found := false
			for _, otherExtra := range p.Extras {
				if otherExtra.ExtraType() == required {
					found = true
					break
				}
			}
			if !found {
				p.logger.Error("Required extra not found", "extra", extra.ExtraType(), "required", required)
				return fmt.Errorf("extra %s requires extra %s", extra.ExtraType(), required)
			}
		}
	}

	// Check if one of the required extras is present
	for _, extra := range p.Extras {
		found := false
		for _, oneOf := range extra.OneOfExtraTypes() {
			for _, otherExtra := range p.Extras {
				if otherExtra.ExtraType() == oneOf {
					found = true
					break
				}
			}
		}
		if !found && len(extra.OneOfExtraTypes()) > 0 {
			p.logger.Error("One of the required extras not found", "extra", extra.ExtraType(), "required", extra.OneOfExtraTypes())
			return fmt.Errorf("extra %s requires one of the extras %v", extra.ExtraType(), extra.OneOfExtraTypes())
		}
	}

	p.logger.Debug("Checked extra incompatibilities")
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

func (p *Project) installDeps() error {
	p.logger.Debug("Installing dependencies")

	// Check if task is installed
	_, err := exec.LookPath("task")
	if err != nil {
		p.logger.Debug("task not found, installing task")
		c := exec.Command("go", "install", "github.com/go-task/task/v3/cmd/task@latest")
		c.Dir = path.Join(p.WorkDir, p.ProjectName)
		err = c.Run()
		if err != nil {
			return fmt.Errorf("failed to install task: %w", err)
		}
	}

	// Run go mod tidy
	c := exec.Command("go", "mod", "tidy")
	c.Dir = path.Join(p.WorkDir, p.ProjectName)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		return err
	}

	// Run task installdeps
	c2 := exec.Command("task", "installdeps")
	c2.Dir = path.Join(p.WorkDir, p.ProjectName)
	c2.Stdout = os.Stdout
	c2.Stderr = os.Stderr
	err = c2.Run()
	if err != nil {
		return err
	}

	p.logger.Debug("Dependencies installed")
	return nil
}

func (p *Project) initGit() error {
	p.logger.Debug("Initializing git")
	c := exec.Command("git", "init")
	c.Dir = path.Join(p.WorkDir, p.ProjectName)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		return err
	}

	p.logger.Debug("Git initialized")
	return nil
}

package internal

import "github.com/JensvandeWiel/go-bat/internal/templates/base"

func (p *Project) GenerateBase() error {
	p.logger.Debug("Generating base project")
	err := p.writeStringTemplateToFile("cmd/root.go", base.CmdRootTmpl, p)
	if err != nil {
		return err
	}
	err = p.writeStringTemplateToFile("cmd/serve.go", base.CmdServeTmpl, p)
	if err != nil {
		return err
	}
	err = p.writeStringTemplateToFile("controllers/main_controller.go", base.ControllersMainControllerTmpl, p)
	if err != nil {
		return err
	}
	err = p.writeStringTemplateToFile("go.mod", base.GoModTmpl, p)
	if err != nil {
		return err
	}
	err = p.writeStringTemplateToFile("go.sum", base.GoSumTmpl, p)
	if err != nil {
		return err
	}
	err = p.writeStringTemplateToFile("main.go", base.MainTmpl, p)
	if err != nil {
		return err
	}
	err = p.writeStringTemplateToFile(".gitignore", base.GitIgnoreTmpl, p)
	if err != nil {
		return err
	}

	err = p.writeStringTemplateToFile(".air.toml", base.AirTmpl, p)
	if err != nil {
		return err
	}

	err = p.writeStringTemplateToFile("Taskfile.yml", base.TaskfileTmpl, p)
	if err != nil {
		return err
	}

	p.logger.Debug("Generated base project")

	return nil
}

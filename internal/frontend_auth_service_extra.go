package internal

import "github.com/JensvandeWiel/go-bat/internal/templates/frontend_auth_service"

type FrontendAuthServiceExtra struct {
}

func (f *FrontendAuthServiceExtra) Generate(project *Project) error {
	data := map[string]interface{}{
		"PackageName": project.PackageName,
	}

	project.logger.Debug("Generating frontend auth service extra")
	err := project.writeStringTemplateToFile("database/stores/user_store.go", frontend_auth_service.UserStoreTemplate, data)
	if err != nil {
		return err
	}

	err = project.writeStringTemplateToFile("database/stores/user_store_test.go", frontend_auth_service.UserStoreTestTemplate, data)
	if err != nil {
		return err
	}

	err = project.writeStringTemplateToFile("services/frontend_auth_service.go", frontend_auth_service.FrontendAuthServiceTemplate, data)
	if err != nil {
		return err
	}

	err = project.writeStringTemplateToFile("services/frontend_auth_service_test.go", frontend_auth_service.FrontendAuthServiceTestTemplate, data)
	if err != nil {
		return err
	}

	err = project.writeStringTemplateToFile("requests/login_request.go", frontend_auth_service.LoginRequestTemplate, data)
	if err != nil {
		return err
	}

	err = project.writeStringTemplateToFile("database/models/user.go", frontend_auth_service.UserTemplate, data)
	if err != nil {
		return err
	}

	version := generateTimestamp()

	err = project.writeStringTemplateToFile("database/migrations/"+version+"_create_users_table.sql", frontend_auth_service.CreateUserTableUpMigration, data)
	if err != nil {
		return err
	}

	err = project.writeStringTemplateToFile("middleware/frontend_auth_middleware.go", frontend_auth_service.FrontendAuthMiddlewareTemplate, data)
	if err != nil {
		return err
	}

	err = project.writeStringTemplateToFile("database/stores/mock_user_store.go", frontend_auth_service.MockUserStoreTemplate, data)
	if err != nil {
		return err
	}

	return nil
}

func (f *FrontendAuthServiceExtra) ModEntries() []string {
	return []string{}
}

func (f *FrontendAuthServiceExtra) GitIgnoreEntries() []string {
	return []string{}
}

func (f *FrontendAuthServiceExtra) GetExtraPersistentFlags() []string {
	return []string{}
}

func (f *FrontendAuthServiceExtra) ExtraType() ExtraType {
	return FrontendAuth
}

func (f *FrontendAuthServiceExtra) DisallowedExtraTypes() []ExtraType {
	return ExtraTypes{}
}

func (f *FrontendAuthServiceExtra) ComposerServices() []string {
	return []string{}
}

func (f *FrontendAuthServiceExtra) ComposerVolumes() []string {
	return []string{}
}

func (f *FrontendAuthServiceExtra) RequiredExtraTypes() ExtraTypes {
	return ExtraTypes{DatabasePgSQL}
}

func (f *FrontendAuthServiceExtra) OneOfExtraTypes() ExtraTypes {
	return ExtraTypes{InertiaReact, InertiaSvelte}
}

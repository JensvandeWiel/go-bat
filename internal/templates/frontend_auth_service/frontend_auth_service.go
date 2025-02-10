package frontend_auth_service

import _ "embed"

//go:embed stores/user_store.go.tmpl
var UserStoreTemplate string

//go:embed stores/user_store_test.go.tmpl
var UserStoreTestTemplate string

//go:embed stores/mock_user_store.go.tmpl
var MockUserStoreTemplate string

//go:embed services/frontend_auth_service.go.tmpl
var FrontendAuthServiceTemplate string

//go:embed services/frontend_auth_service_test.go.tmpl
var FrontendAuthServiceTestTemplate string

//go:embed requests/login_request.go.tmpl
var LoginRequestTemplate string

//go:embed models/user.go.tmpl
var UserTemplate string

//go:embed migrations/20250210151341_create_users_table.sql.tmpl
var CreateUserTableUpMigration string

//go:embed middleware/frontend_auth_middleware.go.tmpl
var FrontendAuthMiddlewareTemplate string

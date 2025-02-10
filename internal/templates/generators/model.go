package generators

import _ "embed"

//go:embed model/model.go.tmpl
var ModelTemplate string

//go:embed model/model_migration.sql.tmpl
var ModelMigrationTemplate string

//go:embed model/model_store.go.tmpl
var ModelStoreTemplate string

//go:embed model/model_store_test.go.tmpl
var ModelStoreTestTemplate string

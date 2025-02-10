package database_pgsql_extra

import _ "embed"

//go:embed database/connect.go.tmpl
var DatabaseConnectTemplate string

//go:embed database/migrations/20250210123935_placeholder.sql
var DatabaseMigrationsPlaceholderTemplate string

//go:embed database/migrations/migrations.go.tmpl
var DatabaseMigrationsTemplate string

//go:embed cmd/migrate.go.tmpl
var MigrateTemplate string

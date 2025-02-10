package database_pgsql_extra

import _ "embed"

//go:embed database/connect.go.tmpl
var DatabaseConnectTemplate string

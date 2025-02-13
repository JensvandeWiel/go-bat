package base

import _ "embed"

//go:embed cmd/root.go.tmpl
var CmdRootTmpl string

//go:embed cmd/serve.go.tmpl
var CmdServeTmpl string

//go:embed controllers/main_controller.go.tmpl
var ControllersMainControllerTmpl string

//go:embed go.mod.tmpl
var GoModTmpl string

//go:embed go.sum.tmpl
var GoSumTmpl string

//go:embed main.go.tmpl
var MainTmpl string

//go:embed .gitignore.tmpl
var GitIgnoreTmpl string

//go:embed .air.toml.tmpl
var AirTmpl string

//go:embed Taskfile.yml.tmpl
var TaskfileTmpl string

//go:embed test_helpers/bat_context.go.tmpl
var TestHelpersEchoContextTmpl string

package inertia_react_extra

import (
	"embed"
	_ "embed"
)

//go:embed controllers/inertia_controller.go.tmpl
var ControllersInertiaController string

//go:embed frontend
var Frontend embed.FS

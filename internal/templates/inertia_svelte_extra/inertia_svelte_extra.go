package inertia_svelte_extra

import (
	"embed"
)

//go:embed controllers/inertia_controller.go.tmpl
var ControllersInertiaController string

//go:embed frontend
var Frontend embed.FS

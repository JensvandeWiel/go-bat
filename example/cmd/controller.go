package cmd

import (
	bat "github.com/JensvandeWiel/go-bat/pkg"
	"github.com/labstack/echo/v4"
	"github.com/romsar/gonertia/v2"
)

type MainController struct {
	bat     *bat.Bat
	inertia *gonertia.Inertia
}

func NewMainController() *MainController {
	return &MainController{}
}

func (c *MainController) Register(app *bat.Bat) error {
	c.bat = app
	c.inertia = bat.GetExtension[*bat.InertiaExtension](app).Inertia
	app.GET("/", c.Index)
	return nil
}

func (c *MainController) GetControllerName() string {
	return "MainController"
}

func (c *MainController) Index(ctx echo.Context) error {
	return c.inertia.Render(ctx.Response(), ctx.Request(), "Index", nil)
}

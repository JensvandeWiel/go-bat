package go_bat

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type BatContext struct {
	echo.Context
}

func (b *Bat) NewContext(req *http.Request, res http.ResponseWriter) echo.Context {
	return &BatContext{
		Context: b.Echo.NewContext(req, res),
	}
}

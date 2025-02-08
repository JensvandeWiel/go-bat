package go_bat

import (
	"github.com/labstack/echo/v4"
)

type BatBase struct {
	Logger *Logger
	*echo.Echo
}

func (b *BatBase) RegisterControllers(controllers ...Controller) error {
	for _, controller := range controllers {
		if err := controller.Register(b); err != nil {
			return err
		}
	}
	return nil
}

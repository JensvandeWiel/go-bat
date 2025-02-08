package pkg

import (
	"github.com/labstack/echo/v4"
	"reflect"
)

type Bat struct {
	Logger *Logger
	*echo.Echo
	extensions map[reflect.Type]interface{}
}

func (b *Bat) RegisterControllers(controllers ...Controller) error {
	for _, controller := range controllers {
		if err := controller.Register(b); err != nil {
			return err
		}
	}
	return nil
}

func NewBat(logger *Logger) *Bat {
	return &Bat{
		Logger:     logger,
		Echo:       echo.New(),
		extensions: make(map[reflect.Type]interface{}),
	}
}

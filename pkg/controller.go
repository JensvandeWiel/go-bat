package pkg

type Controller interface {
	Register(app *Bat) error
	GetControllerName() string
}

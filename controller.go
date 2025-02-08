package go_bat

type Controller interface {
	Register(app *Bat) error
	GetControllerName() string
}

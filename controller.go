package go_bat

type Controller interface {
	Register(app *BatBase) error
	GetControllerName() string
}

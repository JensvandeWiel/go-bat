package pkg

import (
	"errors"
	"reflect"
)

type Extension interface {
	Register(app *Bat) error
}

var ExtensionNotPointerError = errors.New("given extension must be a pointer")

func (b *Bat) RegisterExtension(extension Extension) error {
	if reflect.TypeOf(extension).Kind() != reflect.Ptr {
		return ExtensionNotPointerError
	}
	b.extensions[reflect.TypeOf(extension)] = extension
	return extension.Register(b)
}

func (b *Bat) RegisterExtensions(extensions ...Extension) error {
	for _, ext := range extensions {
		err := b.RegisterExtension(ext)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetExtension[T Extension](b *Bat) T {
	ext := b.extensions[reflect.TypeOf((*T)(nil)).Elem()]
	return ext.(T)
}

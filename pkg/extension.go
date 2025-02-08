package pkg

import (
	"errors"
	"reflect"
)

var ExtensionNotPointerError = errors.New("given extension must be a pointer")

func (b *Bat) RegisterExtension(extension interface{}) error {
	if reflect.TypeOf(extension).Kind() != reflect.Ptr {
		return ExtensionNotPointerError
	}
	b.extensions[reflect.TypeOf(extension)] = extension
	return nil
}

func (b *Bat) RegisterExtensions(extension ...interface{}) error {
	for _, ext := range extension {
		err := b.RegisterExtension(ext)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetExtension[T any](b *Bat) *T {
	ext := b.extensions[reflect.TypeOf((*T)(nil))]
	return ext.(*T)
}

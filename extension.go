package go_bat

import (
	"reflect"
)

func (b *Bat) RegisterExtension(extension interface{}) {
	b.extensions[reflect.TypeOf(extension)] = extension
}

func (b *Bat) RegisterExtensions(extension ...interface{}) {
	for _, ext := range extension {
		b.RegisterExtension(ext)
	}
}

func GetExtension[T any](b *Bat) *T {
	ext := b.extensions[reflect.TypeOf((*T)(nil))]
	return ext.(*T)
}

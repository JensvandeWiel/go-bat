package pkg

import (
	"errors"
	"reflect"
)

type Extension interface {
	Register(app *Bat) error
	Requirements() []reflect.Type
}

var (
	ExtensionNotPointerError = errors.New("given extension must be a pointer")
	CyclicDependencyError    = errors.New("cyclic extension dependency detected")
)

func (b *Bat) RegisterExtensions(extensions ...Extension) error {
	order, err := b.resolveLoadOrder(extensions)
	if err != nil {
		return err
	}
	for _, ext := range order {
		err := ext.Register(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bat) resolveLoadOrder(extensions []Extension) ([]Extension, error) {
	graph := make(map[reflect.Type][]reflect.Type)
	inDegree := make(map[reflect.Type]int)
	for _, ext := range extensions {
		extType := reflect.TypeOf(ext)
		graph[extType] = ext.Requirements()
		for _, req := range ext.Requirements() {
			inDegree[req]++
		}
	}

	var order []Extension
	queue := make([]Extension, 0)
	for _, ext := range extensions {
		if inDegree[reflect.TypeOf(ext)] == 0 {
			queue = append(queue, ext)
		}
	}

	for len(queue) > 0 {
		ext := queue[0]
		queue = queue[1:]
		order = append(order, ext)
		for _, req := range graph[reflect.TypeOf(ext)] {
			inDegree[req]--
			if inDegree[req] == 0 {
				for _, e := range extensions {
					if reflect.TypeOf(e) == req {
						queue = append(queue, e)
						break
					}
				}
			}
		}
	}

	if len(order) != len(extensions) {
		return nil, CyclicDependencyError
	}

	return order, nil
}

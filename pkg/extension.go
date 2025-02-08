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
		b.Logger.Debug("Extension has requirements", "extension", extType, "requirements", ext.Requirements())
	}

	var order []Extension
	queue := make([]Extension, 0)
	for _, ext := range extensions {
		if inDegree[reflect.TypeOf(ext)] == 0 {
			queue = append(queue, ext)
		}
	}

	b.Logger.Debug("Initial queue", "queue", queue)

	for len(queue) > 0 {
		ext := queue[0]
		queue = queue[1:]
		order = append(order, ext)
		b.Logger.Debug("Processing extension", "extension", reflect.TypeOf(ext))
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
		b.Logger.Debug("Current load order", "order", order)
		b.Logger.Debug("Current queue", "queue", queue)
	}

	if len(order) != len(extensions) {
		b.Logger.Error("Cyclic dependency detected")
		return nil, CyclicDependencyError
	}

	b.Logger.Debug("Final load order", "order", order)
	return order, nil
}

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

func (b *Bat) registerExtensions(extensions ...Extension) error {
	order, err := b.resolveLoadOrder(extensions)
	if err != nil {
		return err
	}
	for _, ext := range order {
		err := ext.Register(b)
		if err != nil {
			return err
		}
		b.extensions[reflect.TypeOf(ext)] = ext
	}
	return nil
}

func (b *Bat) resolveLoadOrder(extensions []Extension) ([]Extension, error) {
	// Dependency graph: Maps each extension to the list of extensions that depend on it
	graph := make(map[reflect.Type][]reflect.Type)
	// In-degree: Keeps track of how many dependencies each extension has
	inDegree := make(map[reflect.Type]int)
	// Mapping extensions for quick lookup
	extMap := make(map[reflect.Type]Extension)

	// Step 1: Initialize structures for tracking dependencies
	for _, ext := range extensions {
		// Check if the extension is a pointer
		if reflect.TypeOf(ext).Kind() != reflect.Ptr {
			b.Logger.Error("Extension must be a pointer", "extension", reflect.TypeOf(ext))
			return nil, ExtensionNotPointerError
		}
		var extType reflect.Type
		if reflect.TypeOf(ext).Kind() != reflect.Ptr {
			extType = reflect.TypeOf(ext)
		} else {
			extType = reflect.TypeOf(ext).Elem()
		}
		extMap[extType] = ext             // Store reference to extension
		graph[extType] = []reflect.Type{} // Initialize dependency list
		inDegree[extType] = 0             // Default in-degree (no dependencies)
	}

	// Step 2: Build dependency graph by linking each extension to its dependents
	for _, ext := range extensions {
		extType := reflect.TypeOf(ext).Elem()
		for _, req := range ext.Requirements() {
			reqType := req                                   // The required extension type
			graph[reqType] = append(graph[reqType], extType) // req → extType (dependency link)
			inDegree[extType]++                              // Increase in-degree for the dependent extension
		}

		// Debug: Log the dependencies of each extension
		b.Logger.Debug("Extension has requirements", "extension", extType.Name(), "requirements", ext.Requirements())
	}

	// Step 3: Initialize queue with extensions that have no dependencies (in-degree == 0)
	var order []Extension
	queue := []Extension{}

	for extType, ext := range extMap {
		if inDegree[extType] == 0 { // Only extensions with no dependencies are added
			queue = append(queue, ext)
		}
	}

	// Debug: Log the initial queue state
	b.Logger.Debug("Initial queue", "queue", getExtensionNames(queue))

	// Step 4: Process queue using Kahn's algorithm for topological sorting
	for len(queue) > 0 { // Continue until all extensions are processed
		ext := queue[0]   // Take the first extension in the queue
		queue = queue[1:] // Remove it from the queue
		extType := reflect.TypeOf(ext).Elem()

		// Add the extension to the final resolved load order
		order = append(order, ext)
		b.Logger.Debug("Processing extension", "extension", extType.Name())

		// Reduce the in-degree of all extensions that depend on this one
		for _, dependent := range graph[extType] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 { // If an extension has no remaining dependencies, it's ready to be loaded
				queue = append(queue, extMap[dependent])
			}
		}

		// Debug: Log the state of the queue and the load order after each step
		b.Logger.Debug("Current load order", "order", getExtensionNames(order))
		b.Logger.Debug("Current queue", "queue", getExtensionNames(queue))
	}

	// Step 5: If not all extensions are processed, there must be a cyclic dependency, or a missing dependency
	if len(order) != len(extensions) {
		// Check if the dependency is missing
		for _, ext := range extensions {
			extType := reflect.TypeOf(ext).Elem()
			if inDegree[extType] > 0 {
				for _, req := range ext.Requirements() {
					reqType := req
					if _, ok := extMap[reqType]; !ok {
						b.Logger.Error("Extension not found", "extension", reqType.Name())
						return nil, errors.New("extension not found")
					}
				}
			}
		}
		// If not, there must be a cyclic dependency
		b.Logger.Error("Cyclic dependency detected")
		return nil, CyclicDependencyError
	}

	// Debug: Log the final resolved extension load order
	b.Logger.Debug("Final load order", "order", getExtensionNames(order))
	return order, nil
}

func getExtensionNames(extensions []Extension) []string {
	names := make([]string, len(extensions))
	for i, ext := range extensions {
		names[i] = reflect.TypeOf(ext).Elem().Name()
	}
	return names
}

func GetExtension[T Extension](b *Bat) T {
	ext := b.extensions[reflect.TypeOf((*T)(nil)).Elem()]
	return ext.(T)
}

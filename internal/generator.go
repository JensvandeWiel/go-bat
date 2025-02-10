package internal

import (
	"fmt"
)

type Generator interface {
	Generate(name string, extra bool) error
}

func ParseGenerator(name string, project *Project) (Generator, error) {
	project.logger.Debug("Parsing generator", "name", name)
	switch name {
	case "model":
		return &ModelGenerator{
			project: project,
		}, nil
	default:
		project.logger.Error("Unknown generator", "name", name)
		return nil, fmt.Errorf("unknown generator %q", name)
	}
}

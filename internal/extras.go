package internal

type Extra interface {
	// Generate generates the extra
	Generate(project *Project) error
	// ModEntries returns the entries that need to be added to the go.mod file
	ModEntries() []string
	// GitIgnoreEntries returns the entries that need to be added to the .gitignore file
	GitIgnoreEntries() []string
}

type ExtraType string

func (t ExtraType) String() string {
	return string(t)
}

const (
	InertiaReact ExtraType = "inertia-react"
)

func ParseExtra(extra string) Extra {
	switch extra {
	case InertiaReact.String():
		return &InertiaReactExtra{}
	default:
		return nil
	}
}

func ParseExtras(extras []string) []Extra {
	var parsedExtras []Extra
	for _, extra := range extras {
		parsedExtra := ParseExtra(extra)
		if parsedExtra != nil {
			parsedExtras = append(parsedExtras, parsedExtra)
		}
	}
	return parsedExtras
}

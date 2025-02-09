package internal

type Extra interface {
	// Generate generates the extra
	Generate(project *Project) error
	// ModEntries returns the entries that need to be added to the go.mod file
	ModEntries() []string
	// GitIgnoreEntries returns the entries that need to be added to the .gitignore file
	GitIgnoreEntries() []string
	// GetExtraPersistentFlags returns the flags that need to be added to the root command
	GetExtraPersistentFlags() []string
	// ExtraType returns the name of the extra
	ExtraType() ExtraType
}

type ExtraType string

func (t ExtraType) String() string {
	return string(t)
}

const (
	InertiaReact ExtraType = "inertia-react"
)

func ParseExtraType(extra string) ExtraType {
	switch extra {
	case "inertia-react":
		return InertiaReact
	default:
		return ""
	}
}

func ParseExtraTypes(extras []string) []ExtraType {
	var parsedExtras []ExtraType
	for _, extra := range extras {
		parsedExtra := ParseExtraType(extra)
		if parsedExtra != "" {
			parsedExtras = append(parsedExtras, parsedExtra)
		}
	}
	return parsedExtras
}

func ParseExtra(extra ExtraType) Extra {
	switch extra {
	case InertiaReact:
		return &InertiaReactExtra{}
	default:
		return nil
	}
}

func ParseExtras(extras []ExtraType) []Extra {
	var parsedExtras []Extra
	for _, extra := range extras {
		parsedExtra := ParseExtra(extra)
		if parsedExtra != nil {
			parsedExtras = append(parsedExtras, parsedExtra)
		}
	}
	return parsedExtras
}

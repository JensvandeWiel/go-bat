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

const (
	InertiaReact ExtraType = "inertia-react"
)

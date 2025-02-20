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
	// DisallowedExtraTypes returns the extra types that are not allowed with this extra
	DisallowedExtraTypes() []ExtraType
	// ComposerServices returns the services that need to be added to the composer file
	ComposerServices() []string
	// ComposerVolumes returns the volumes that need to be added to the composer file
	ComposerVolumes() []string
	// RequiredExtraTypes returns the extra types that are required with this extra
	RequiredExtraTypes() ExtraTypes
	// OneOfExtraTypes returns the extra type where one of them is required
	OneOfExtraTypes() ExtraTypes
}

type ExtraType string

type ExtraTypes []ExtraType

func (e ExtraTypes) HasExtra(extra ExtraType) bool {
	for _, t := range e {
		if t == extra {
			return true
		}
	}
	return false
}

func (t ExtraType) String() string {
	return string(t)
}

const (
	InertiaReact  ExtraType = "inertia-react"
	InertiaSvelte ExtraType = "inertia-svelte"
	DatabasePgSQL ExtraType = "database-pgsql"
	FrontendAuth  ExtraType = "frontend-auth"
)

func ParseExtraType(extra string) ExtraType {
	switch extra {
	case "inertia-react":
		return InertiaReact
	case "inertia-svelte":
		return InertiaSvelte
	case "database-pgsql":
		return DatabasePgSQL
	case "frontend-auth":
		return FrontendAuth
	default:
		return ""
	}
}

func ParseExtraTypes(extras []string) ExtraTypes {
	var parsedExtras ExtraTypes
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
	case InertiaSvelte:
		return &InertiaSvelteExtra{}
	case DatabasePgSQL:
		return &DatabasePgSQLExtra{}
	case FrontendAuth:
		return &FrontendAuthServiceExtra{}
	default:
		return nil
	}
}

func ParseExtras(extras ExtraTypes) []Extra {
	var parsedExtras []Extra
	for _, extra := range extras {
		parsedExtra := ParseExtra(extra)
		if parsedExtra != nil {
			parsedExtras = append(parsedExtras, parsedExtra)
		}
	}
	return parsedExtras
}

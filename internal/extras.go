package internal

type Extra string

const (
	ExtraInertiaReact Extra = "inertia-react"
)

func ParseExtra(extra string) Extra {
	switch extra {
	case "inertia-react":
		return ExtraInertiaReact
	default:
		return ""
	}
}

func ParseExtras(extras ...string) []Extra {
	var parsedExtras []Extra
	for _, extra := range extras {
		parsedExtras = append(parsedExtras, ParseExtra(extra))
	}
	return parsedExtras
}

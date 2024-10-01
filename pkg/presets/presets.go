package presets

import (
	"embed"
)

//go:embed configs/*.json configs/*.yaml
var Presets embed.FS

type MatchStringPreset struct {
	MatchString       string
	ParameterDefaults []string
}

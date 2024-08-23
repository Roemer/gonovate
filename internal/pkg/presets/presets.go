package presets

import "embed"

//go:embed *.json *.yaml
var Presets embed.FS

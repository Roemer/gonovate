package presets

import (
	"fmt"
	"regexp"
	"strings"
)

var matchStringPresetRegex = regexp.MustCompile(`preset:\s*(.*?)(?:\((.*)\))?\s*$`)
var versioningPresetRegex = regexp.MustCompile(`preset:\s*(.*?)\s*$`)

// Resolves a given match string with a preset (if any).
func ResolveMatchString(matchString string, matchStringPresets map[string]*MatchStringPreset) (string, error) {
	m := matchStringPresetRegex.FindStringSubmatch(matchString)
	if m != nil {
		// Get the name and check if it exists
		presetName := m[1]
		preset, ok := matchStringPresets[presetName]
		if !ok {
			return "", fmt.Errorf("matchString preset '%s' not found", presetName)
		}

		// Get the parameters passed from the matchString
		parametersFromString := []string{}
		if m[2] != "" {
			parametersFromString = strings.Split(m[2], ",")
		}
		// Get the max number of parameters from the string and the defaults
		maxParams := len(parametersFromString)
		if len(preset.ParameterDefaults) > maxParams {
			maxParams = len(preset.ParameterDefaults)
		}
		// Just return the string if there are no parameters at all
		if maxParams == 0 {
			return preset.MatchString, nil
		}
		// Build the list of parameters
		params := make([]interface{}, maxParams)
		// Set the defaults
		for i, v := range preset.ParameterDefaults {
			params[i] = v
		}
		// Overwrite with parameters from the matchString
		for i, v := range parametersFromString {
			if v != "" {
				params[i] = v
			}
		}
		// Return the formatted string
		return fmt.Sprintf(preset.MatchString, params...), nil
	}
	return matchString, nil
}

// Resolves a given versioning with a preset (if any).
func ResolveVersioning(versioning string, versioningPresets map[string]string) (string, error) {
	m := versioningPresetRegex.FindStringSubmatch(versioning)
	if m != nil {
		presetName := m[1]
		preset, ok := versioningPresets[presetName]
		if !ok {
			return "", fmt.Errorf("versioning preset '%s' not found", presetName)
		}
		return preset, nil
	}
	return versioning, nil
}

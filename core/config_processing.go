package core

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var configMatchStringPresetRegex = regexp.MustCompile(`preset:\s*(.*?)(?:\((.*)\))?\s*$`)

// Filters all rules, creating a combined settings object for the manager and a list of possible rules for packages.
func (config *Config) FilterForManager(managerConfig *Manager) (*ManagerSettings, []*Rule) {
	possiblePackageRules := []*Rule{}
	managerSettings := &ManagerSettings{}
	// Loop thru all the rules
	for _, rule := range config.Rules {
		// Check if there are conditions which exclude this manager
		if rule.Matches != nil {
			// ManagerId
			if len(rule.Matches.Managers) > 0 && !slices.Contains(rule.Matches.Managers, managerConfig.Id) {
				continue
			}
		}
		// Process and apply the settings for the manager
		managerSettings.MergeWith(rule.ManagerSettings)
		// The rule contains settings for packages, so add it to the list
		if rule.PackageSettings != nil {
			possiblePackageRules = append(possiblePackageRules, rule)
		}
	}
	return managerSettings, possiblePackageRules
}

// Resolves a given match string with a template (if any)
func (config *Config) ResolveMatchString(matchString string) string {
	m := configMatchStringPresetRegex.FindStringSubmatch(matchString)
	if m != nil {
		preset, ok := config.MatchStringPresets[m[1]]
		if ok {
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
				return preset.MatchString
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
			return fmt.Sprintf(preset.MatchString, params...)
		}
	}
	return matchString
}

func FilterHostConfigsForHost(host string, hostRules []*HostRule) *HostRule {
	for _, hostRule := range hostRules {
		if strings.Contains(host, hostRule.MatchHost) {
			return hostRule
		}
	}
	return nil
}

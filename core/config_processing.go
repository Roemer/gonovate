package core

import (
	"slices"
)

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

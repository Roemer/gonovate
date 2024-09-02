package config

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/roemer/gonovate/internal/pkg/shared"
	"github.com/roemer/gotaskr/goext"
	"github.com/samber/lo"
)

var configMatchStringPresetRegex = regexp.MustCompile(`preset:\s*(.*?)(?:\((.*)\))?\s*$`)
var configVersioningPresetRegex = regexp.MustCompile(`preset:\s*(.*?)\s*$`)

// This method processes the root config object. This should be called on any config object just after loading.
func (c *RootConfig) PostLoadProcess() {
	// Convert managerSettings/dependencySettings to rules and add them to keep the priority order
	for _, managerConfig := range c.Managers {
		if managerConfig.ManagerSettings != nil || managerConfig.DependencySettings != nil {
			newRule := &Rule{
				Matches: &RuleMatch{
					Managers: []string{managerConfig.Id},
				},
			}
			if managerConfig.ManagerSettings != nil {
				newRule.ManagerSettings = &ManagerSettings{}
				newRule.ManagerSettings.MergeWith(managerConfig.ManagerSettings)
				managerConfig.ManagerSettings = nil
			}
			if managerConfig.DependencySettings != nil {
				newRule.DependencySettings = &DependencySettings{}
				newRule.DependencySettings.MergeWith(managerConfig.DependencySettings)
				managerConfig.DependencySettings = nil
			}
			c.Rules = goext.Prepend(c.Rules, newRule)
		}
	}
}

func (config *RootConfig) GetMergedManagerSettings(managerConfig *ManagerConfig) *ManagerSettings {
	mergedManagerSettings := &ManagerSettings{}
	for _, rule := range config.Rules {
		if rule.Matches != nil {
			// ManagerId
			if len(rule.Matches.Managers) > 0 && !slices.Contains(rule.Matches.Managers, managerConfig.Id) {
				continue
			}
			// ManagerTypes
			if len(rule.Matches.ManagerTypes) > 0 && !slices.Contains(rule.Matches.ManagerTypes, managerConfig.Type) {
				continue
			}
		}
		mergedManagerSettings.MergeWith(rule.ManagerSettings)
	}
	return mergedManagerSettings
}

func (config *RootConfig) GetManagerConfigById(managerId string) *ManagerConfig {
	managerConfig, _ := lo.Find(config.Managers, func(managerConfig *ManagerConfig) bool { return managerConfig.Id == managerId })
	return managerConfig
}

func (config *RootConfig) EnrichDependencyFromRules(dependency *shared.Dependency) {
	// Get the config of the manager for this dependency
	managerConfig := config.GetManagerConfigById(dependency.ManagerId)

	// Prepare the merged settings
	mergedDependencySettings := &DependencySettings{}

	// Search for matching rules and merge them
	for _, rule := range config.Rules {
		if rule.Matches != nil {
			// Manager related matches
			if managerConfig != nil {
				// ManagerIds
				if len(rule.Matches.Managers) > 0 && slices.IndexFunc(rule.Matches.Managers, func(matchId string) bool {
					return matchStringMatches(managerConfig.Id, matchId)
				}) < 0 {
					continue
				}
				// ManagerTypes
				if len(rule.Matches.ManagerTypes) > 0 && !slices.Contains(rule.Matches.ManagerTypes, managerConfig.Type) {
					continue
				}
			}
			// Files
			ok, _ := shared.FilePathMatchesPattern(dependency.FilePath, rule.Matches.Files...)
			if len(rule.Matches.Files) > 0 && !ok {
				continue
			}
			// DependencyNames
			if len(rule.Matches.DependencyNames) > 0 && slices.IndexFunc(rule.Matches.DependencyNames, func(matchName string) bool {
				return matchStringMatches(dependency.Name, matchName)
			}) < 0 {
				continue
			}
			// Datasources
			if len(rule.Matches.Datasources) > 0 && slices.IndexFunc(rule.Matches.Datasources, func(ds shared.DatasourceType) bool { return ds == dependency.Datasource }) < 0 {
				continue
			}
		}
		mergedDependencySettings.MergeWith(rule.DependencySettings)
	}

	// Apply the rule settings where the dependency has no value yet
	if dependency.Name == "" {
		dependency.Name = mergedDependencySettings.DependencyName
	}
	if dependency.Datasource == "" {
		dependency.Datasource = mergedDependencySettings.Datasource
	}
	if dependency.Skip == nil {
		dependency.Skip = mergedDependencySettings.Skip
	}
	if dependency.SkipReason == "" {
		dependency.SkipReason = mergedDependencySettings.SkipReason
	}
	if dependency.MaxUpdateType == "" {
		dependency.MaxUpdateType = mergedDependencySettings.MaxUpdateType
	}
	if dependency.AllowUnstable == nil {
		dependency.AllowUnstable = mergedDependencySettings.AllowUnstable
	}
	dependency.RegistryUrls = lo.Union(dependency.RegistryUrls, mergedDependencySettings.RegistryUrls)
	if dependency.Versioning == "" {
		dependency.Versioning = mergedDependencySettings.Versioning
	}
	if dependency.ExtractVersion == "" {
		dependency.ExtractVersion = mergedDependencySettings.ExtractVersion
	}
	if dependency.IgnoreNonMatching == nil {
		dependency.IgnoreNonMatching = mergedDependencySettings.IgnoreNonMatching
	}
	dependency.PostUpgradeReplacements = lo.Union(dependency.PostUpgradeReplacements, mergedDependencySettings.PostUpgradeReplacements)
	if dependency.GroupName == "" {
		dependency.GroupName = mergedDependencySettings.GroupName
	}
}

func matchStringMatches(input string, matchString string) bool {
	if strings.HasPrefix(matchString, "re:") {
		re := regexp.MustCompile(matchString[3:])
		return re.MatchString(input)
	}
	return input == matchString
}

// Resolves a given match string with a preset (if any).
func (config *RootConfig) ResolveMatchString(matchString string) (string, error) {
	m := configMatchStringPresetRegex.FindStringSubmatch(matchString)
	if m != nil {
		// Get the name and check if it exists
		presetName := m[1]
		preset, ok := config.MatchStringPresets[presetName]
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
func (config *RootConfig) ResolveVersioning(versioning string) (string, error) {
	m := configVersioningPresetRegex.FindStringSubmatch(versioning)
	if m != nil {
		presetName := m[1]
		preset, ok := config.VersioningPresets[presetName]
		if !ok {
			return "", fmt.Errorf("versioning preset '%s' not found", presetName)
		}
		return preset, nil
	}
	return versioning, nil
}

// Filters the host rules by the given host and returns the first match.
func (config *RootConfig) FilterHostConfigsForHost(host string) *HostRule {
	for _, hostRule := range config.HostRules {
		if strings.Contains(host, hostRule.MatchHost) {
			return hostRule
		}
	}
	return nil
}

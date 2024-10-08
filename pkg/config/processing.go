package config

import (
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/roemer/gonovate/pkg/datasources"
	"github.com/roemer/gonovate/pkg/managers"
	"github.com/roemer/gonovate/pkg/presets"
	"github.com/roemer/gotaskr/goext"
	"github.com/samber/lo"
)

// This method processes the gonovate config object. This should be called on any config object just after loading.
func (c *GonovateConfig) PostLoadProcess() {
	// Convert managerConfigs/dependencyConfigs to rules and add them to keep the priority order
	for _, managerConfig := range c.Managers {
		if managerConfig.ManagerConfig != nil || managerConfig.DependencyConfig != nil {
			newRule := &Rule{
				Matches: &RuleMatch{
					Managers: []string{managerConfig.Id},
				},
			}
			if managerConfig.ManagerConfig != nil {
				newRule.ManagerConfig = &ManagerConfig{}
				newRule.ManagerConfig.MergeWith(managerConfig.ManagerConfig)
				managerConfig.ManagerConfig = nil
			}
			if managerConfig.DependencyConfig != nil {
				newRule.DependencyConfig = &DependencyConfig{}
				newRule.DependencyConfig.MergeWith(managerConfig.DependencyConfig)
				managerConfig.DependencyConfig = nil
			}
			c.Rules = goext.Prepend(c.Rules, newRule)
		}
	}
}

func (config *GonovateConfig) GetMergedManagerConfig(managerId string, managerType common.ManagerType) *ManagerConfig {
	mergedManagerConfig := &ManagerConfig{}
	for _, rule := range config.Rules {
		if rule.Matches != nil {
			// ManagerId
			if len(rule.Matches.Managers) > 0 && !slices.Contains(rule.Matches.Managers, managerId) {
				continue
			}
			// ManagerTypes
			if len(rule.Matches.ManagerTypes) > 0 && !slices.Contains(rule.Matches.ManagerTypes, managerType) {
				continue
			}
		}
		mergedManagerConfig.MergeWith(rule.ManagerConfig)
	}
	return mergedManagerConfig
}

func (config *GonovateConfig) GetManagerConfigById(managerId string) *Manager {
	manager, _ := lo.Find(config.Managers, func(manager *Manager) bool { return manager.Id == managerId })
	return manager
}

// Creates a manager out of the config with the given id and manager type.
func (config *GonovateConfig) GetManager(managerId string, managerType common.ManagerType, logger *slog.Logger) (common.IManager, error) {
	mergedManagerConfig := config.GetMergedManagerConfig(managerId, managerType)

	managerSettings := &common.ManagerSettings{
		Logger:       logger,
		Disabled:     mergedManagerConfig.Disabled,
		FilePatterns: mergedManagerConfig.FilePatterns,
		RegexManagerSettings: &common.RegexManagerSettings{
			MatchStringPresets: config.MatchStringPresetsToPresets(),
			MatchStrings:       mergedManagerConfig.MatchStrings,
		},
		DevcontainerManagerSettings: mergedManagerConfig.ToCommonDevcontainerManagerSettings(),
	}

	return managers.GetManager(managerId, managerType, managerSettings)
}

// Creates a datasource out of the config of the given datasource type.
func (config *GonovateConfig) GetDatasource(datasourceType common.DatasourceType, logger *slog.Logger) (common.IDatasource, error) {
	datasourceSettings := &common.DatasourceSettings{
		Logger:    logger,
		HostRules: config.HostRules,
	}
	return datasources.GetDatasource(datasourceType, datasourceSettings)
}

// Applies rules and presets to the dependency
func (config *GonovateConfig) ApplyToDependency(dependency *common.Dependency) error {
	// Apply the rules to the dependency
	config.applyRulesToDependency(dependency)

	// Resolve the versioning
	if resolvedVersioning, err := presets.ResolveVersioning(dependency.Versioning, config.VersioningPresets); err != nil {
		return err
	} else {
		dependency.Versioning = resolvedVersioning
	}

	return nil
}

func (config *GonovateConfig) applyRulesToDependency(dependency *common.Dependency) {
	// Get the config of the manager for this dependency
	var managerConfig *Manager
	if dependency.ManagerInfo != nil && dependency.ManagerInfo.ManagerId != "" {
		managerConfig = config.GetManagerConfigById(dependency.ManagerInfo.ManagerId)
	}

	// Prepare the merged settings
	mergedDependencyConfig := &DependencyConfig{}

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
			ok, _ := common.FilePathMatchesPattern(dependency.FilePath, rule.Matches.Files...)
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
			if len(rule.Matches.Datasources) > 0 && slices.IndexFunc(rule.Matches.Datasources, func(ds common.DatasourceType) bool { return ds == dependency.Datasource }) < 0 {
				continue
			}
		}
		mergedDependencyConfig.MergeWith(rule.DependencyConfig)
	}

	// Apply the rule settings where the dependency has no value yet
	if dependency.Name == "" {
		dependency.Name = mergedDependencyConfig.DependencyName
	}
	if dependency.Datasource == "" {
		dependency.Datasource = mergedDependencyConfig.Datasource
	}
	if dependency.Skip == nil {
		dependency.Skip = mergedDependencyConfig.Skip
	}
	if dependency.SkipReason == "" {
		dependency.SkipReason = mergedDependencyConfig.SkipReason
	}
	if dependency.MaxUpdateType == "" {
		dependency.MaxUpdateType = mergedDependencyConfig.MaxUpdateType
	}
	if dependency.AllowUnstable == nil {
		dependency.AllowUnstable = mergedDependencyConfig.AllowUnstable
	}
	dependency.RegistryUrls = lo.Union(dependency.RegistryUrls, mergedDependencyConfig.RegistryUrls)
	if dependency.Versioning == "" {
		dependency.Versioning = mergedDependencyConfig.Versioning
	}
	if dependency.ExtractVersion == "" {
		dependency.ExtractVersion = mergedDependencyConfig.ExtractVersion
	}
	if dependency.IgnoreNonMatching == nil {
		dependency.IgnoreNonMatching = mergedDependencyConfig.IgnoreNonMatching
	}
	dependency.PostUpgradeReplacements = lo.Union(dependency.PostUpgradeReplacements, mergedDependencyConfig.PostUpgradeReplacements)
	if dependency.GroupName == "" {
		dependency.GroupName = mergedDependencyConfig.GroupName
	}
}

func matchStringMatches(input string, matchString string) bool {
	if strings.HasPrefix(matchString, "re:") {
		re := regexp.MustCompile(matchString[3:])
		return re.MatchString(input)
	}
	return input == matchString
}

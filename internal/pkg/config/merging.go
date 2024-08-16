package config

import (
	"maps"
	"slices"

	"github.com/samber/lo"
)

func (configA *RootConfig) MergeWithAsCopy(configB *RootConfig) *RootConfig {
	merged := &RootConfig{}
	merged.MergeWith(configA)
	merged.MergeWith(configB)
	return merged
}

func (configA *RootConfig) MergeWith(configB *RootConfig) {
	if configB == nil {
		return
	}
	// Platform
	if configB.Platform != "" {
		configA.Platform = configB.Platform
	}
	// PlatformSettings
	if configA.PlatformSettings == nil {
		configA.PlatformSettings = &PlatformSettings{}
	}
	configA.PlatformSettings.MergeWith(configB.PlatformSettings)
	// MatchStringPresets
	if configA.MatchStringPresets == nil {
		configA.MatchStringPresets = map[string]*MatchStringPreset{}
	}
	for key, value := range configB.MatchStringPresets {
		configA.MatchStringPresets[key] = &MatchStringPreset{
			MatchString:       value.MatchString,
			ParameterDefaults: make([]string, len(value.ParameterDefaults)),
		}
		copy(configA.MatchStringPresets[key].ParameterDefaults, value.ParameterDefaults)
	}
	// VersioningPresets
	if configA.VersioningPresets == nil {
		configA.VersioningPresets = map[string]string{}
	}
	maps.Copy(configA.VersioningPresets, configB.VersioningPresets)
	// Extends
	configA.Extends = lo.Union(configA.Extends, configB.Extends)
	// IgnorePatterns
	configA.IgnorePatterns = lo.Union(configA.IgnorePatterns, configB.IgnorePatterns)
	// Managers
	if configA.Managers == nil {
		configA.Managers = []*ManagerConfig{}
	}
	for _, managerB := range configB.Managers {
		// Search for an existing manager with the same id
		managerAIndex := slices.IndexFunc(configA.Managers, func(m *ManagerConfig) bool { return m.Id == managerB.Id })
		if managerAIndex >= 0 {
			// Found one so merge it
			configA.Managers[managerAIndex].MergeWith(managerB)
		} else {
			// Not found, so add it
			newManager := &ManagerConfig{}
			newManager.MergeWith(managerB)
			configA.Managers = append(configA.Managers, newManager)
		}
	}
	// Rules
	configA.Rules = append(configA.Rules, configB.Rules...)
	// Host Rules
	configA.HostRules = append(configA.HostRules, configB.HostRules...)
}

func (platformSettingsA *PlatformSettings) MergeWith(platformSettingsB *PlatformSettings) {
	if platformSettingsB == nil {
		return
	}
	// Token
	if platformSettingsB.Token != "" {
		platformSettingsA.Token = platformSettingsB.Token
	}
	// GitAuthor
	if platformSettingsB.GitAuthor != "" {
		platformSettingsA.GitAuthor = platformSettingsB.GitAuthor
	}
	// Endpoint
	if platformSettingsB.Endpoint != "" {
		platformSettingsA.Endpoint = platformSettingsB.Endpoint
	}
	// Direct
	if platformSettingsB.Inplace != nil {
		platformSettingsA.Inplace = platformSettingsB.Inplace
	}
	// Projects
	platformSettingsA.Projects = lo.Union(platformSettingsA.Projects, platformSettingsB.Projects)
	// BaseBranch
	if platformSettingsB.BaseBranch != "" {
		platformSettingsA.BaseBranch = platformSettingsB.BaseBranch
	}
	// BranchPrefix
	if platformSettingsB.BranchPrefix != "" {
		platformSettingsA.BranchPrefix = platformSettingsB.BranchPrefix
	}
}

func (managerA *ManagerConfig) MergeWith(managerB *ManagerConfig) {
	if managerB == nil {
		return
	}
	// Id
	managerA.Id = managerB.Id
	// Type
	managerA.Type = managerB.Type
	// Manager Settings
	if managerA.ManagerSettings == nil {
		managerA.ManagerSettings = &ManagerSettings{}
	}
	managerA.ManagerSettings.MergeWith(managerB.ManagerSettings)

	// Dependency Settings
	if managerA.DependencySettings == nil {
		managerA.DependencySettings = &DependencySettings{}
	}
	managerA.DependencySettings.MergeWith(managerB.DependencySettings)
}

func (managerSettingsA *ManagerSettings) MergeWith(managerSettingsB *ManagerSettings) {
	if managerSettingsB == nil {
		return
	}
	// Disabled
	if managerSettingsB.Disabled != nil {
		managerSettingsA.Disabled = managerSettingsB.Disabled
	}
	// FilePatterns
	managerSettingsA.FilePatterns = lo.Union(managerSettingsA.FilePatterns, managerSettingsB.FilePatterns)
	// MatchStrings
	managerSettingsA.MatchStrings = lo.Union(managerSettingsA.MatchStrings, managerSettingsB.MatchStrings)
}

func (dependencySettingsA *DependencySettings) MergeWith(dependencySettingsB *DependencySettings) {
	if dependencySettingsB == nil {
		return
	}
	// MaxUpdateType
	if dependencySettingsB.MaxUpdateType != "" {
		dependencySettingsA.MaxUpdateType = dependencySettingsB.MaxUpdateType
	}
	// AllowUnstable
	if dependencySettingsB.AllowUnstable != nil {
		dependencySettingsA.AllowUnstable = dependencySettingsB.AllowUnstable
	}
	// RegistryUrls
	dependencySettingsA.RegistryUrls = lo.Union(dependencySettingsA.RegistryUrls, dependencySettingsB.RegistryUrls)
	// Versioning
	if dependencySettingsB.Versioning != "" {
		dependencySettingsA.Versioning = dependencySettingsB.Versioning
	}
	// ExtractVersion
	if dependencySettingsB.ExtractVersion != "" {
		dependencySettingsA.ExtractVersion = dependencySettingsB.ExtractVersion
	}
	// IgnoreNonMatching
	if dependencySettingsB.IgnoreNonMatching != nil {
		dependencySettingsA.IgnoreNonMatching = dependencySettingsB.IgnoreNonMatching
	}
	// DependencyName
	if dependencySettingsB.DependencyName != "" {
		dependencySettingsA.DependencyName = dependencySettingsB.DependencyName
	}
	// Datasource
	if dependencySettingsB.Datasource != "" {
		dependencySettingsA.Datasource = dependencySettingsB.Datasource
	}
	// PostUpgradeReplacements
	dependencySettingsA.PostUpgradeReplacements = lo.Union(dependencySettingsA.PostUpgradeReplacements, dependencySettingsB.PostUpgradeReplacements)
	// GroupName
	if dependencySettingsB.GroupName != "" {
		dependencySettingsA.GroupName = dependencySettingsB.GroupName
	}
}

package config

import (
	"maps"
	"slices"

	"github.com/samber/lo"
)

func (configA *RootConfig) MergeWith(configB *RootConfig) *RootConfig {
	if configB == nil {
		return configA
	}
	// Platform
	if configB.Platform != "" {
		configA.Platform = configB.Platform
	}
	// PlatformSettings
	configA.PlatformSettings = configA.PlatformSettings.MergeWith(configB.PlatformSettings)
	// Extends
	configA.Extends = lo.Union(configA.Extends, configB.Extends)
	// IgnorePatterns
	configA.IgnorePatterns = lo.Union(configA.IgnorePatterns, configB.IgnorePatterns)
	// MatchStringPresets
	if configA.MatchStringPresets == nil {
		configA.MatchStringPresets = map[string]*MatchStringPreset{}
	}
	maps.Copy(configA.MatchStringPresets, configB.MatchStringPresets)
	// VersioningPresets
	if configA.VersioningPresets == nil {
		configA.VersioningPresets = map[string]string{}
	}
	maps.Copy(configA.VersioningPresets, configB.VersioningPresets)
	// Managers
	for _, manager := range configB.Managers {
		managerAIndex := slices.IndexFunc(configA.Managers, func(m *ManagerConfig) bool { return m.Id == manager.Id })
		if managerAIndex >= 0 {
			configA.Managers[managerAIndex].MergeWith(manager)
		} else {
			configA.Managers = append(configA.Managers, manager)
		}
		// Convert managerSettings/dependencySettings to rules and add them to keep the priority order
		if manager.managerSettings != nil {
			configA.Rules = append(configA.Rules, &Rule{
				Matches: &RuleMatch{
					Managers: []string{manager.Id},
				},
				ManagerSettings: (&ManagerSettings{}).MergeWith(manager.managerSettings),
			})
		}
		if manager.dependencySettings != nil {
			configA.Rules = append(configA.Rules, &Rule{
				Matches: &RuleMatch{
					Managers: []string{manager.Id},
				},
				DependencySettings: (&DependencySettings{}).MergeWith(manager.dependencySettings),
			})
		}
	}
	// Rules
	configA.Rules = append(configA.Rules, configB.Rules...)
	// Host Rules
	configA.HostRules = append(configA.HostRules, configB.HostRules...)

	return configA
}

func (platformSettingsA *PlatformSettings) MergeWith(platformSettingsB *PlatformSettings) *PlatformSettings {
	// Check if any of the objects is nil and if so, return the other (which might also be nil)
	if platformSettingsA == nil {
		return platformSettingsB
	}
	if platformSettingsB == nil {
		return platformSettingsA
	}
	// Both are set, so merge them
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
	if platformSettingsB.Direct != nil {
		platformSettingsA.Direct = platformSettingsB.Direct
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
	return platformSettingsA
}

func (managerA *ManagerConfig) MergeWith(managerB *ManagerConfig) {
	// Manager Settings
	managerA.managerSettings = managerA.managerSettings.MergeWith(managerB.managerSettings)

	// Dependency Settings
	managerA.dependencySettings = managerA.dependencySettings.MergeWith(managerB.dependencySettings)
}

func (managerSettingsA *ManagerSettings) MergeWith(managerSettingsB *ManagerSettings) *ManagerSettings {
	// Check if any of the objects is nil and if so, return the other (which might also be nil)
	if managerSettingsA == nil {
		return managerSettingsB
	}
	if managerSettingsB == nil {
		return managerSettingsA
	}
	// Both are set, so merge them
	// Disabled
	if managerSettingsB.Disabled != nil {
		managerSettingsA.Disabled = managerSettingsB.Disabled
	}
	// FilePatterns
	managerSettingsA.FilePatterns = lo.Union(managerSettingsA.FilePatterns, managerSettingsB.FilePatterns)
	// MatchStrings
	managerSettingsA.MatchStrings = lo.Union(managerSettingsA.MatchStrings, managerSettingsB.MatchStrings)
	return managerSettingsA
}

func (dependencySettingsA *DependencySettings) MergeWith(dependencySettingsB *DependencySettings) *DependencySettings {
	// Check if any of the objects is nil and if so, return the other (which might also be nil)
	if dependencySettingsA == nil {
		return dependencySettingsB
	}
	if dependencySettingsB == nil {
		return dependencySettingsA
	}
	// Both are set, so merge them
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
	return dependencySettingsA
}

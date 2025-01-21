package config

import (
	"maps"
	"slices"

	"github.com/samber/lo"
)

func (configA *GonovateConfig) MergeWithAsCopy(configB *GonovateConfig) *GonovateConfig {
	merged := &GonovateConfig{}
	merged.MergeWith(configA)
	merged.MergeWith(configB)
	return merged
}

func (configA *GonovateConfig) MergeWith(configB *GonovateConfig) {
	if configB == nil {
		return
	}
	// Platform
	if configA.Platform == nil {
		configA.Platform = &PlatformConfig{}
	}
	configA.Platform.MergeWith(configB.Platform)
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
		configA.Managers = []*Manager{}
	}
	for _, managerB := range configB.Managers {
		// Search for an existing manager with the same id
		managerAIndex := slices.IndexFunc(configA.Managers, func(m *Manager) bool { return m.Id == managerB.Id })
		if managerAIndex >= 0 {
			// Found one so merge it
			configA.Managers[managerAIndex].MergeWith(managerB)
		} else {
			// Not found, so add it
			newManager := &Manager{}
			newManager.MergeWith(managerB)
			configA.Managers = append(configA.Managers, newManager)
		}
	}
	// Rules
	configA.Rules = append(configA.Rules, configB.Rules...)
	// Host Rules
	configA.HostRules = append(configA.HostRules, configB.HostRules...)
}

func (platformConfigA *PlatformConfig) MergeWith(platformConfigB *PlatformConfig) {
	if platformConfigB == nil {
		return
	}
	// Platform
	if platformConfigB.Type != "" {
		platformConfigA.Type = platformConfigB.Type
	}
	// Token
	if platformConfigB.Token != "" {
		platformConfigA.Token = platformConfigB.Token
	}
	// GitAuthor
	if platformConfigB.GitAuthor != "" {
		platformConfigA.GitAuthor = platformConfigB.GitAuthor
	}
	// Endpoint
	if platformConfigB.Endpoint != "" {
		platformConfigA.Endpoint = platformConfigB.Endpoint
	}
	// Direct
	if platformConfigB.Inplace != nil {
		platformConfigA.Inplace = platformConfigB.Inplace
	}
	// Projects
	platformConfigA.Projects = lo.Union(platformConfigA.Projects, platformConfigB.Projects)
	// BaseBranch
	if platformConfigB.BaseBranch != "" {
		platformConfigA.BaseBranch = platformConfigB.BaseBranch
	}
	// BranchPrefix
	if platformConfigB.BranchPrefix != "" {
		platformConfigA.BranchPrefix = platformConfigB.BranchPrefix
	}
}

func (managerA *Manager) MergeWith(managerB *Manager) {
	if managerB == nil {
		return
	}
	// Id
	managerA.Id = managerB.Id
	// Type
	managerA.Type = managerB.Type
	// Manager Settings
	if managerA.ManagerConfig == nil {
		managerA.ManagerConfig = &ManagerConfig{}
	}
	managerA.ManagerConfig.MergeWith(managerB.ManagerConfig)

	// Dependency Settings
	if managerA.DependencyConfig == nil {
		managerA.DependencyConfig = &DependencyConfig{}
	}
	managerA.DependencyConfig.MergeWith(managerB.DependencyConfig)
}

func (ManagerConfigA *ManagerConfig) MergeWith(ManagerConfigB *ManagerConfig) {
	if ManagerConfigB == nil {
		return
	}
	// Disabled
	if ManagerConfigB.Disabled != nil {
		ManagerConfigA.Disabled = ManagerConfigB.Disabled
	}
	// ClearFilePatterns
	if ManagerConfigB.ClearFilePatterns != nil {
		ManagerConfigA.ClearFilePatterns = ManagerConfigB.ClearFilePatterns
	}
	// FilePatterns
	if ManagerConfigB.ClearFilePatterns != nil && *ManagerConfigB.ClearFilePatterns {
		ManagerConfigA.FilePatterns = append([]string{}, ManagerConfigB.FilePatterns...)
	} else {
		ManagerConfigA.FilePatterns = lo.Union(ManagerConfigA.FilePatterns, ManagerConfigB.FilePatterns)
	}
	// MatchStrings
	ManagerConfigA.MatchStrings = lo.Union(ManagerConfigA.MatchStrings, ManagerConfigB.MatchStrings)
	// DevcontainerConfig
	if len(ManagerConfigB.DevcontainerConfig) > 0 {
		// Make sure the settings object exists in A
		if ManagerConfigA.DevcontainerConfig == nil {
			ManagerConfigA.DevcontainerConfig = map[string][]*DevcontainerFeatureDependency{}
		}

		// Loop thru the features
		for featureName, featureDependencies := range ManagerConfigB.DevcontainerConfig {
			// Make sure the feature exists in A
			if _, ok := ManagerConfigA.DevcontainerConfig[featureName]; !ok {
				ManagerConfigA.DevcontainerConfig[featureName] = []*DevcontainerFeatureDependency{}
			}
			// Merge the individual feature dependencies
			for _, featureDependency := range featureDependencies {
				// Search for an existing featureDependency in A with the same property
				idx := slices.IndexFunc(ManagerConfigA.DevcontainerConfig[featureName], func(m *DevcontainerFeatureDependency) bool {
					return m.Property == featureDependency.Property
				})
				if idx >= 0 {
					// Found one so merge it
					ManagerConfigA.DevcontainerConfig[featureName][idx].MergeWith(featureDependency)
				} else {
					// Not found, so add it
					newFeature := &DevcontainerFeatureDependency{}
					newFeature.MergeWith(featureDependency)
					ManagerConfigA.DevcontainerConfig[featureName] = append(ManagerConfigA.DevcontainerConfig[featureName], newFeature)
				}
			}
		}
	}
}

func (DependencyConfigA *DependencyConfig) MergeWith(DependencyConfigB *DependencyConfig) {
	if DependencyConfigB == nil {
		return
	}
	// Skip
	if DependencyConfigB.Skip != nil {
		DependencyConfigA.Skip = DependencyConfigB.Skip
	}
	// SkipReason
	if DependencyConfigB.SkipReason != "" {
		DependencyConfigA.SkipReason = DependencyConfigB.SkipReason
	}
	// MaxUpdateType
	if DependencyConfigB.MaxUpdateType != "" {
		DependencyConfigA.MaxUpdateType = DependencyConfigB.MaxUpdateType
	}
	// AllowUnstable
	if DependencyConfigB.AllowUnstable != nil {
		DependencyConfigA.AllowUnstable = DependencyConfigB.AllowUnstable
	}
	// RegistryUrls
	DependencyConfigA.RegistryUrls = lo.Union(DependencyConfigA.RegistryUrls, DependencyConfigB.RegistryUrls)
	// Versioning
	if DependencyConfigB.Versioning != "" {
		DependencyConfigA.Versioning = DependencyConfigB.Versioning
	}
	// ExtractVersion
	if DependencyConfigB.ExtractVersion != "" {
		DependencyConfigA.ExtractVersion = DependencyConfigB.ExtractVersion
	}
	// IgnoreNonMatching
	if DependencyConfigB.IgnoreNonMatching != nil {
		DependencyConfigA.IgnoreNonMatching = DependencyConfigB.IgnoreNonMatching
	}
	// DependencyName
	if DependencyConfigB.DependencyName != "" {
		DependencyConfigA.DependencyName = DependencyConfigB.DependencyName
	}
	// Datasource
	if DependencyConfigB.Datasource != "" {
		DependencyConfigA.Datasource = DependencyConfigB.Datasource
	}
	// PostUpgradeReplacements
	DependencyConfigA.PostUpgradeReplacements = lo.Union(DependencyConfigA.PostUpgradeReplacements, DependencyConfigB.PostUpgradeReplacements)
	// GroupName
	if DependencyConfigB.GroupName != "" {
		DependencyConfigA.GroupName = DependencyConfigB.GroupName
	}
}

func (objA *DevcontainerFeatureDependency) MergeWith(objB *DevcontainerFeatureDependency) {
	if objB == nil {
		return
	}
	// Property
	if objB.Property != "" {
		objA.Property = objB.Property
	}
	// Datasource
	if objB.Datasource != "" {
		objA.Datasource = objB.Datasource
	}
	// DependencyName
	if objB.DependencyName != "" {
		objA.DependencyName = objB.DependencyName
	}
}

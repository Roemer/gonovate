package core

import (
	"slices"

	"github.com/samber/lo"
)

func (configA *Config) MergeWith(configB *Config) {
	if configB == nil {
		return
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
	// Managers
	for _, manager := range configB.Managers {
		managerAIndex := slices.IndexFunc(configA.Managers, func(m *Manager) bool { return m.Id == manager.Id })
		if managerAIndex >= 0 {
			configA.Managers[managerAIndex].MergeWith(manager)
		} else {
			configA.Managers = append(configA.Managers, manager)
		}
		// Convert managerSettings/packageSettings to rules and add them to keep the priority order
		if manager.ManagerSettings != nil {
			configA.Rules = append(configA.Rules, &Rule{
				Matches: &RuleMatch{
					Managers: []string{manager.Id},
				},
				ManagerSettings: (&ManagerSettings{}).MergeWith(manager.ManagerSettings),
			})
		}
		if manager.PackageSettings != nil {
			configA.Rules = append(configA.Rules, &Rule{
				Matches: &RuleMatch{
					Managers: []string{manager.Id},
				},
				PackageSettings: (&PackageSettings{}).MergeWith(manager.PackageSettings),
			})
		}
	}
	// Rules
	configA.Rules = append(configA.Rules, configB.Rules...)
	// Host Rules
	configA.HostRules = append(configA.HostRules, configB.HostRules...)
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
	// Owner
	if platformSettingsB.Owner != "" {
		platformSettingsA.Owner = platformSettingsB.Owner
	}
	// Repository
	if platformSettingsB.Repository != "" {
		platformSettingsA.Repository = platformSettingsB.Repository
	}
	// GitAuthor
	if platformSettingsB.GitAuthor != "" {
		platformSettingsA.GitAuthor = platformSettingsB.GitAuthor
	}
	return platformSettingsA
}

func (managerA *Manager) MergeWith(managerB *Manager) {
	// Manager Settings
	managerA.ManagerSettings = managerA.ManagerSettings.MergeWith(managerB.ManagerSettings)

	// Package Settings
	managerA.PackageSettings = managerA.PackageSettings.MergeWith(managerB.PackageSettings)
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
	// PostUpgradeReplacements
	managerSettingsA.PostUpgradeReplacements = lo.Union(managerSettingsA.PostUpgradeReplacements, managerSettingsB.PostUpgradeReplacements)
	return managerSettingsA
}

func (packageSettingsA *PackageSettings) MergeWith(packageSettingsB *PackageSettings) *PackageSettings {
	// Check if any of the objects is nil and if so, return the other (which might also be nil)
	if packageSettingsA == nil {
		return packageSettingsB
	}
	if packageSettingsB == nil {
		return packageSettingsA
	}
	// Both are set, so merge them
	// MaxUpdateType
	if packageSettingsB.MaxUpdateType != "" {
		packageSettingsA.MaxUpdateType = packageSettingsB.MaxUpdateType
	}
	// AllowUnstable
	if packageSettingsB.AllowUnstable != nil {
		packageSettingsA.AllowUnstable = packageSettingsB.AllowUnstable
	}
	// RegistryUrls
	packageSettingsA.RegistryUrls = lo.Union(packageSettingsA.RegistryUrls, packageSettingsB.RegistryUrls)
	// Versioning
	if packageSettingsB.Versioning != "" {
		packageSettingsA.Versioning = packageSettingsB.Versioning
	}
	// ExtractVersion
	if packageSettingsB.ExtractVersion != "" {
		packageSettingsA.ExtractVersion = packageSettingsB.ExtractVersion
	}
	// IgnoreNonMatching
	if packageSettingsB.IgnoreNonMatching != nil {
		packageSettingsA.IgnoreNonMatching = packageSettingsB.IgnoreNonMatching
	}
	// PackageName
	if packageSettingsB.PackageName != "" {
		packageSettingsA.PackageName = packageSettingsB.PackageName
	}
	// Datasource
	if packageSettingsB.Datasource != "" {
		packageSettingsA.Datasource = packageSettingsB.Datasource
	}
	return packageSettingsA
}

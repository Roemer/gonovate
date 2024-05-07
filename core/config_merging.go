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
	managerSettingsA.Disabled = managerSettingsB.Disabled
	// FilePatterns
	managerSettingsA.FilePatterns = lo.Union(managerSettingsA.FilePatterns, managerSettingsB.FilePatterns)
	// MatchStrings
	managerSettingsA.MatchStrings = lo.Union(managerSettingsA.MatchStrings, managerSettingsB.MatchStrings)
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
	// UseUnstable
	if packageSettingsB.UseUnstable != nil {
		packageSettingsA.UseUnstable = packageSettingsB.UseUnstable
	}
	return packageSettingsA
}

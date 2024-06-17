package core

import (
	"fmt"
	"strings"
	"time"
)

// This type represents a concrete dependency which was found by a manager.
type Dependency struct {
	Name           string
	Datasource     DatasourceType
	MaxUpdateType  UpdateType
	Versioning     string
	ExtractVersion string
	// The current version of the dependency.
	Version string
	// The type of the dependency. Used to allow different handlings per type in the manager.
	Type string
}

func (d *Dependency) String() string {
	parts := []string{}
	if d.Name != "" {
		parts = append(parts, fmt.Sprintf("name: %s", d.Name))
	}
	if d.Version != "" {
		parts = append(parts, fmt.Sprintf("version: %s", d.Version))
	}
	if d.Datasource != "" {
		parts = append(parts, fmt.Sprintf("datasource: %s", d.Datasource))
	}
	if d.Type != "" {
		parts = append(parts, fmt.Sprintf("type: %s", d.Type))
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}

/*func (d *Dependency) ApplySettingsAndRules(possiblePackageRules []*Rule) {
	mergedPackageSettings := &PackageSettings{}
	// Loop thru the rules and apply the ones that match
	for _, rule := range possiblePackageRules {
		isAnyMatch := rule.Matches.IsMatchAll()
		// Check if there is at least one condition that matches
		if !isAnyMatch && rule.Matches != nil {
			// Manager-IDs
			if !isAnyMatch && len(rule.Matches.Managers) > 0 {
				if slices.Contains(rule.Matches.Managers, managerId) {
					isAnyMatch = true
				}
			}
			// Files
			if !isAnyMatch && len(rule.Matches.Files) > 0 {
				isMatch, err := core.FilePathMatchesPattern(currentFile, rule.Matches.Files...)
				if err != nil {
					return nil, err
				}
				if isMatch {
					isAnyMatch = true
				}
			}
			// PackageName
			if !isAnyMatch && len(rule.Matches.PackageNames) > 0 {
				if packageSettings.PackageName != "" && slices.Contains(rule.Matches.PackageNames, packageSettings.PackageName) {
					isAnyMatch = true
				}
			}
			// Datasource
			if !isAnyMatch && len(rule.Matches.Datasources) > 0 {
				if packageSettings.Datasource != "" && slices.Contains(rule.Matches.Datasources, packageSettings.Datasource) {
					isAnyMatch = true
				}
			}
		}
		// The rule has at least one match, add it
		if isAnyMatch {
			// Merge the current rules package settings
			packageSettings.MergeWith(rule.PackageSettings)
			// Make sure that the priority settings are not overwritten
			packageSettings.MergeWith(priorityPackageSettings)
		}
	}
}*/

// ?
type DependencyLookupInfo struct {
	// ReplaceString
	// Start/End-Index
}

// This type contains the updated information for a dependency.
type DependencyUpdate struct {
	// The dependency that will be updated.
	Dependency *Dependency
	// The new version of the dependency.
	NewVersion string
	// The type of the update.
	UpdateType UpdateType
	// The date whem this version was published.
	ReleaseDate time.Time
	// A map of hashes related to this dependency.
	Hashes map[string]string
}

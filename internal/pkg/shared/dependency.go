package shared

import (
	"fmt"
	"strings"
)

// This type represents a concrete dependency which was found by a manager.
type Dependency struct {
	// The name of the dependency.
	Name string
	// The current version of the dependency (unprocessed).
	Version string
	// The type of the dependency. Used to allow different handlings per type in the manager. Optional.
	Type string
	// The datasource of the dependency.
	Datasource DatasourceType
	// The id of the manager from which this dependency was found.
	ManagerId string
	// An object that can contain data which is set/read from the manager to process the dependency.
	ManagerData interface{}
	// The filepath from where this dependency was found.
	FilePath string

	// Defines how much the dependency is allowed to update. Can be "major", "minor", or "patch".
	MaxUpdateType UpdateType
	// This flag defines if unstable releases are allowed. Unstable usually means a version that also has parts with text.
	AllowUnstable *bool
	// A list of registry urls to use. Allows overwriting the default. Depends on the datasource.
	RegistryUrls []string
	// Defines the regexp to use to parse the version into separate parts. See https://github.com/Roemer/gover for more details.
	Versioning string
	// An optional regexp that is used to separate the version part from the rest of the raw version string.
	ExtractVersion string
	// A flag to indicate if versions from a remote that do not match the versioning should be ignored or give an exception.
	IgnoreNonMatching *bool

	// Allows defining regexes that replace further information from dependencies (like hash) after updating
	PostUpgradeReplacements []string

	// Contains the information about the new release if any is found.
	NewRelease *ReleaseInfo
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

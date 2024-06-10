package core

import (
	"fmt"
	"time"
)

// This type represents a concrete dependency which was found by a manager.
type Dependency struct {
	// The name of the dependency.
	Name string
	// The current version of the dependency.
	Version string
	// The datasource of the dependency.
	Datasource DatasourceType
	// The type of the dependency. Used to allow different handlings per type in the manager.
	Type string
}

func (d *Dependency) String() string {
	return fmt.Sprintf("{name: %s, version: %s, datasource: %s, type: %s}", d.Name, d.Version, d.Datasource, d.Name)
}

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

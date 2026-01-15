package common

import (
	"log/slog"
)

// This is the interface for the gonovate engine.
type IEngine interface {
	// Gets the logger for the engine.
	Logger() *slog.Logger
	// Gets the desired manager.
	GetManager(managerSettings *ManagerSettings) (IManager, error)
	// Resolves match-strings with presets.
	ResolveMatchString(matchString string) (string, error)
	// Resolves versioning-strings with presets.
	ResolveVersioning(versioning string) (string, error)
}

// This is the interface that needs to be implemented by all managers.
type IManager interface {
	// Gets the id of the manager.
	Id() string
	// Gets the type of the manager.
	Type() ManagerType
	// Gets the settings with which the manager was created.
	Settings() *ManagerSettings
	// Extracts all dependencies from the manager.
	ExtractDependencies(filePath string) ([]*Dependency, error)
	// Applies a dependency update with the manager.
	ApplyDependencyUpdate(dependency *Dependency) error
}

// This is the interface that needs to be implemented by all datasources.
type IDatasource interface {
	// Gets all possible releases for the dependency.
	GetReleases(dependency *Dependency) ([]*ReleaseInfo, error)
	// Gets the digest for the dependency.
	GetDigest(dependency *Dependency, releaseVersion string) (string, error)
	// Gets additional data for the dependency and the new release.
	GetAdditionalData(dependency *Dependency, newRelease *ReleaseInfo, dataType string) (string, error)
	// Handles the dependency update searching.
	SearchDependencyUpdate(dependency *Dependency) (*ReleaseInfo, error)
}

type ICache interface {
	// Gets the cached releases for the given datasource type and identifier.
	Get(datasourceType DatasourceType, cacheIdentifier string) ([]*ReleaseInfo, error)
	// Sets the cached releases for the given datasource type and identifier.
	Set(datasourceType DatasourceType, cacheIdentifier string, releases []*ReleaseInfo) error
}

package managers

import (
	"fmt"
	"gonovate/core"
	"gonovate/datasources"
	"log/slog"
	"strings"
	"time"

	"github.com/roemer/gover"
)

// This type represents a concrete dependency which was found by a manager.
type Dependency struct {
	// The name of the dependency.
	Name string
	// The current version of the dependency.
	Version string
	// The datasource of the dependency.
	Datasource core.DatasourceType
	// The type of the dependency. Used to allow different handlings per type in the manager.
	Type string
}

func (d *Dependency) String() string {
	return fmt.Sprintf("name: %s, version: %s, datasource: %s, type: %s", d.Name, d.Version, d.Datasource, d.Name)
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
	UpdateType core.UpdateType
	// The date whem this version was published.
	ReleaseDate time.Time
	// A map of hashes related to this dependency.
	Hashes map[string]string
}

// This is the interface that needs to be implemented by all managers.
type IManager2 interface {
	// Extracts a single dependency from the manager.
	ExtractDependency(dependencyName string) (*Dependency, error)
	// Extracts all dependencies from the manager.
	ExtractDependencies(content string) ([]*Dependency, error)
	// Applies a dependency update with the manager.
	ApplyDependencyUpdate(dependencyUpdate *DependencyUpdate) error
}

type IManager interface {
	// Gets all changes
	GetChanges() ([]core.IChange, error)
	// Applies a group of changes
	ApplyChanges(changes []core.IChange) error

	// Internal method to get all changes
	getChanges(mergedManagerSettings *core.ManagerSettings, possiblePackageRules []*core.Rule) ([]core.IChange, error)
	// Internal method to apply a group of changes
	applyChanges(changes []core.IChange) error
}

type managerBase2 struct {
	logger        *slog.Logger
	ManagerConfig *core.Manager
	impl          IManager2
}

type managerBase struct {
	logger        *slog.Logger
	Config        *core.Config
	ManagerConfig *core.Manager
	impl          IManager
}

func GetManager2(logger *slog.Logger, managerConfig *core.Manager) (IManager2, error) {
	switch managerConfig.Type {
	case core.MANAGER_TYPE_GOMOD:
		return NewGoModManager(logger, managerConfig), nil
	}
	return nil, fmt.Errorf("no manager defined for '%s'", managerConfig.Type)
}

func GetManager(logger *slog.Logger, config *core.Config, managerConfig *core.Manager) (IManager, error) {
	switch managerConfig.Type {
	case core.MANAGER_TYPE_INLINE:
		return NewInlineManager(logger, config, managerConfig), nil
	case core.MANAGER_TYPE_REGEX:
		return NewRegexManager(logger, config, managerConfig), nil
	}
	return nil, fmt.Errorf("no manager defined for '%s'", managerConfig.Type)
}

func (manager *managerBase) GetChanges() ([]core.IChange, error) {
	// Filter the settings for this manager, also collect all package settings that might apply for this manager
	mergedManagerSettings, possiblePackageRules := manager.Config.FilterForManager(manager.ManagerConfig)

	// Skip the manager if it is disabled
	if mergedManagerSettings.Disabled != nil && *mergedManagerSettings.Disabled {
		manager.logger.Info(fmt.Sprintf("Manager '%s': Skip as it is disabled", manager.ManagerConfig.Id))
		return nil, nil
	}

	// Get the changes for the manager
	manager.logger.Info(fmt.Sprintf("Manager '%s': Get changes", manager.ManagerConfig.Id))
	changes, err := manager.impl.getChanges(mergedManagerSettings, possiblePackageRules)
	if err != nil {
		manager.logger.Error(fmt.Sprintf("Manager failed with error: %s", err.Error()))
	}
	return changes, err
}

func (manager *managerBase) ApplyChanges(changes []core.IChange) error {
	manager.logger.Info(fmt.Sprintf("Applying %d change(s)", len(changes)))
	err := manager.impl.applyChanges(changes)
	if err != nil {
		manager.logger.Error(fmt.Sprintf("Manager failed with error: %s", err.Error()))
	}
	return err
}

// Searches for a new package version with the correct datasource.
func (manager *managerBase) searchPackageUpdate(currentVersionString string, packageSettings *core.PackageSettings) (*core.ReleaseInfo, *gover.Version, error) {
	// Validate the mandatory fields
	if len(currentVersionString) == 0 {
		return nil, nil, fmt.Errorf("no version defined")
	}
	if len(packageSettings.PackageName) == 0 {
		return nil, nil, fmt.Errorf("no packageName defined")
	}
	if len(packageSettings.Datasource) == 0 {
		return nil, nil, fmt.Errorf("no datasource defined")
	}
	// Sanitize some values like trimming (eg. for forgotten \r in Windows files...)
	currentVersionString, _ = manager.sanitizeString(currentVersionString)
	// Log
	manager.logger.Info(fmt.Sprintf("Searching a '%s' update for '%s' with version '%s' on datasource '%s'", packageSettings.MaxUpdateType, packageSettings.PackageName, currentVersionString, packageSettings.Datasource))

	// Lookup the correct datasource
	ds, err := datasources.GetDatasource(manager.logger, manager.Config, packageSettings.Datasource)
	if err != nil {
		return nil, nil, err
	}

	// Search for a new version
	newReleaseInfo, currentVersion, err := ds.SearchPackageUpdate(currentVersionString, packageSettings)

	// Return the result
	return newReleaseInfo, currentVersion, err
}

// Sanitize the value with trimming (eg. for forgotten \r in Windows files...)
func (manager *managerBase) sanitizeString(value string) (string, int) {
	newValue := strings.Trim(value, " \r\n")
	return newValue, len(value) - len(newValue)
}

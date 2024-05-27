package managers

import (
	"fmt"
	"gonovate/core"
	"gonovate/datasources"
	"log/slog"
	"strings"

	"github.com/roemer/gover"
)

type IManager interface {
	// Gets all changes
	GetChanges() ([]core.IChange, error)
	// Applies a group of changes
	ApplyChanges(changes []core.IChange) error

	// Internal method to get all changes
	getChanges() ([]core.IChange, error)
	// Internal method to apply a group of changes
	applyChanges(changes []core.IChange) error
}

type managerBase struct {
	logger       *slog.Logger
	GlobalConfig *core.Config
	Config       *core.Manager
	impl         IManager
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
	manager.logger.Info(fmt.Sprintf("Get changes for manager %s", manager.Config.Id))
	changes, err := manager.impl.getChanges()
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
func (manager *managerBase) searchPackageUpdate(currentVersionString string, packageSettings *core.PackageSettings, hostRules []*core.HostRule) (*core.ReleaseInfo, *gover.Version, error) {
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
	ds, err := datasources.GetDatasource(manager.logger, packageSettings.Datasource)
	if err != nil {
		return nil, nil, err
	}

	// Search for a new version
	newReleaseInfo, currentVersion, err := ds.SearchPackageUpdate(currentVersionString, packageSettings, hostRules)

	// Return the result
	return newReleaseInfo, currentVersion, err
}

// Sanitize the value with trimming (eg. for forgotten \r in Windows files...)
func (manager *managerBase) sanitizeString(value string) (string, int) {
	newValue := strings.Trim(value, " \r\n")
	return newValue, len(value) - len(newValue)
}

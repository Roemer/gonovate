package managers

import (
	"fmt"
	"gonovate/core"
	"gonovate/datasources"
	"gonovate/platforms"
	"log/slog"
	"strings"

	"github.com/roemer/gover"
)

type IManager interface {
	// Entry point to run a manager
	Run() error
	// The internal that starts the processing of a manager
	process() error
	// Applies a group of changes
	applyChanges(changes []core.IChange) error
}

type managerBase struct {
	logger       *slog.Logger
	GlobalConfig *core.Config
	Config       *core.Manager
	Platform     platforms.IPlatform
	impl         IManager
}

func (manager *managerBase) Run() error {
	err := manager.impl.process()
	if err != nil {
		manager.logger.Error(fmt.Sprintf("Manager failed with error: %s", err.Error()))
	}
	return err
}

func GetManager(logger *slog.Logger, config *core.Config, managerConfig *core.Manager, platform platforms.IPlatform) (IManager, error) {
	switch managerConfig.Type {
	case core.MANAGER_TYPE_INLINE:
		return NewInlineManager(logger, config, managerConfig, platform), nil
	case core.MANAGER_TYPE_REGEX:
		return NewRegexManager(logger, config, managerConfig, platform), nil
	}
	return nil, fmt.Errorf("no manager defined for '%s'", managerConfig.Type)
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

func (manager *managerBase) processChanges(changes []core.IChange) error {
	manager.logger.Debug(fmt.Sprintf("Processing %d change(s)", len(changes)))
	// Special case for the noop platform: apply all changes at once
	if manager.Platform.Type() == core.PLATFORM_TYPE_NOOP {
		if err := manager.impl.applyChanges(changes); err != nil {
			return err
		}
		return nil
	}

	// TODO: Grouping and sorting

	// For the other platforms, process the changes in the groups
	for _, change := range changes {
		// Prepare
		if err := manager.Platform.PrepareForChanges(change); err != nil {
			return err
		}
		// Apply the changes
		if err := manager.impl.applyChanges([]core.IChange{change}); err != nil {
			return err
		}
		// Submit
		if err := manager.Platform.SubmitChanges(change); err != nil {
			return err
		}
		// Publish
		if err := manager.Platform.PublishChanges(change); err != nil {
			return err
		}
		// Notify
		if err := manager.Platform.NotifyChanges(change); err != nil {
			return err
		}
		// Reset
		if err := manager.Platform.ResetToBase(); err != nil {
			return err
		}
	}
	return nil
}

// Sanitize the value with trimming (eg. for forgotten \r in Windows files...)
func (manager *managerBase) sanitizeString(value string) (string, int) {
	newValue := strings.Trim(value, " \r\n")
	return newValue, len(value) - len(newValue)
}

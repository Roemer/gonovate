package managers

import (
	"fmt"
	"gonovate/core"
	"gonovate/datasources"
	"log/slog"
)

type managerBase struct {
	logger       *slog.Logger
	GlobalConfig *core.Config
	Config       *core.Manager
}

// Searches for a new package version with the correct datasource.
func (manager *managerBase) searchPackageUpdate(currentVersion string, packageSettings *core.PackageSettings, hostRules []*core.HostRule) (*core.ReleaseInfo, error) {
	// Validate the mandatory fields
	if len(currentVersion) == 0 {
		return nil, fmt.Errorf("no version defined")
	}
	if len(packageSettings.PackageName) == 0 {
		return nil, fmt.Errorf("no packageName defined")
	}
	if len(packageSettings.Datasource) == 0 {
		return nil, fmt.Errorf("no datasource defined")
	}
	// Log
	manager.logger.Info(fmt.Sprintf("Searching a '%s' update for '%s' with version '%s' on datasource '%s'", packageSettings.MaxUpdateType, packageSettings.PackageName, currentVersion, packageSettings.Datasource))

	// Lookup the correct datasource
	ds, err := datasources.GetDatasource(manager.logger, packageSettings.Datasource)
	if err != nil {
		return nil, err
	}

	// Search for a new version
	newReleaseInfo, err := datasources.SearchPackageUpdate(ds, currentVersion, packageSettings, hostRules)

	// Return the result
	return newReleaseInfo, err
}

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
func (manager *managerBase) searchPackageUpdate(datasourceName string, packageName string, currentVersion string, packageSettings *core.PackageSettings, hostRules []*core.HostRule) (string, bool, error) {
	manager.logger.Info(fmt.Sprintf("Searching a '%s' update for '%s' with version '%s' on datasource '%s'", packageSettings.MaxUpdateType, packageName, currentVersion, datasourceName))

	// Lookup the correct datasource
	ds, err := datasources.GetDatasource(manager.logger, datasourceName)
	if err != nil {
		return "", false, err
	}

	// Search for a new version
	newVersion, hasNewVersion, err := ds.SearchPackageUpdate(packageName, currentVersion, packageSettings, hostRules)

	// Return the result
	return newVersion, hasNewVersion, err
}

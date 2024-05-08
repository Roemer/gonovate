package datasources

import (
	"fmt"
	"gonovate/core"
	"log/slog"

	"github.com/roemer/gover"
)

type datasourcesBase struct {
	logger *slog.Logger
}

type datasource interface {
	SearchPackageUpdate(packageName string, currentVersion string, packageSettings *core.PackageSettings, hostRules []*core.HostRule) (string, bool, error)
}

func GetDatasource(logger *slog.Logger, datasource string) (datasource, error) {
	if datasource == core.DATASOURCE_TYPE_NODEJS {
		return NewNodeJsDatasource(logger), nil
	}
	if datasource == core.DATASOURCE_TYPE_DOCKER {
		return NewDockerDatasource(logger), nil
	}
	return nil, fmt.Errorf("no datasource defined for '%s'", datasource)
}

func (ds *datasourcesBase) getReferenceVersionForUpdateType(updateType string, currentVersion *gover.Version) *gover.Version {
	if updateType == core.UPDATE_TYPE_MAJOR {
		return gover.EmptyVersion
	}
	if updateType == core.UPDATE_TYPE_MINOR {
		return gover.ParseSimple(currentVersion.Major())
	}
	if updateType == core.UPDATE_TYPE_PATCH {
		return gover.ParseSimple(currentVersion.Major(), currentVersion.Minor())
	}
	return nil
}

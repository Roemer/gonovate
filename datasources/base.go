package datasources

import (
	"fmt"
	"gonovate/core"
	"io"
	"log/slog"
	"net/http"

	"github.com/roemer/gover"
)

type datasourcesBase struct {
	logger *slog.Logger
}

type datasource interface {
	SearchPackageUpdate(packageName string, currentVersion string, packageSettings *core.PackageSettings) (string, bool, error)
}

func GetDatasource(logger *slog.Logger, datasource string) (datasource, error) {
	if datasource == core.DATASOURCE_TYPE_NODEJS {
		return NewNodeJsDatasource(logger), nil
	}
	return nil, fmt.Errorf("no datasource defined for '%s'", datasource)
}

func (ds *datasourcesBase) DownloadToMemory(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file '%s'. Status code: %d", url, resp.StatusCode)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	return bodyBytes, nil
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

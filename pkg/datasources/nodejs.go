package datasources

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type NodeJsDatasource struct {
	datasourceBase
}

func NewNodeJsDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &NodeJsDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_NODEJS,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *NodeJsDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://nodejs.org/dist", dependency.RegistryUrls)
	indexFilePath := "index.json"
	ltsOnly := strings.HasSuffix(dependency.Name, "lts")

	// Download the index file
	downloadUrl, err := url.JoinPath(registryUrl, indexFilePath)
	if err != nil {
		return nil, err
	}
	indexFileBytes, err := shared.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return nil, err
	}

	// Parse the data as json
	var jsonData []map[string]interface{}
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*shared.ReleaseInfo{}
	for _, entry := range jsonData {
		versionString := entry["version"].(string)
		ltsValue := entry["lts"]
		dateString := entry["date"].(string)
		if ltsOnly && ltsValue == false {
			continue
		}
		releaseDate, err := time.Parse(time.DateOnly, dateString)
		if err != nil {
			return nil, fmt.Errorf("failed parsing date '%s': %w", dateString, err)
		}
		releases = append(releases, &shared.ReleaseInfo{
			ReleaseDate:   releaseDate,
			VersionString: versionString,
		})
	}
	return releases, nil
}

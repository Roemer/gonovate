package datasources

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/roemer/gonovate/internal/config"
	"github.com/roemer/gonovate/internal/core"
)

type GoVersionDatasource struct {
	datasourceBase
}

func NewGoVersionDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &GoVersionDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   core.DATASOURCE_TYPE_GOVERSION,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GoVersionDatasource) getReleases(dependencySettings *config.DependencySettings) ([]*core.ReleaseInfo, error) {
	baseUrl := "https://go.dev"
	if len(dependencySettings.RegistryUrls) > 0 {
		baseUrl = dependencySettings.RegistryUrls[0]
		ds.logger.Debug(fmt.Sprintf("Using custom registry url: %s", baseUrl))
	}
	indexFilePath := "dl/?mode=json&include=all"
	stableOnly := strings.HasSuffix(dependencySettings.DependencyName, "stable")

	// Download the index file
	downloadUrl := baseUrl
	if !strings.HasSuffix(downloadUrl, "/") {
		downloadUrl += "/"
	}
	downloadUrl += indexFilePath
	indexFileBytes, err := core.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return nil, err
	}

	// Parse the data as json
	var jsonData []map[string]interface{}
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*core.ReleaseInfo{}
	for _, entry := range jsonData {
		versionString := entry["version"].(string)
		stableValue := entry["stable"]
		if stableOnly && stableValue != true {
			continue
		}
		releases = append(releases, &core.ReleaseInfo{
			VersionString: versionString,
		})
	}
	return releases, nil
}

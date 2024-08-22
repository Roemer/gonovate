package datasources

import (
	"encoding/json"
	"log/slog"
	"net/url"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type GradleVersionDatasource struct {
	datasourceBase
}

func NewGradleVersionDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &GradleVersionDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_GRADLEVERSION,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GradleVersionDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://raw.githubusercontent.com", dependency.RegistryUrls)
	indexFilePath := "gradle/gradle/master/released-versions.json"

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
	var jsonData map[string]interface{}
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*shared.ReleaseInfo{}
	for _, entry := range jsonData["finalReleases"].([]interface{}) {
		versionString := entry.(map[string]interface{})["version"].(string)
		releases = append(releases, &shared.ReleaseInfo{
			VersionString: versionString,
		})
	}
	return releases, nil
}

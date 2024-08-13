package datasources

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type GoVersionDatasource struct {
	datasourceBase
}

func NewGoVersionDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &GoVersionDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_GOVERSION,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GoVersionDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	baseUrl := "https://go.dev"
	if len(dependency.RegistryUrls) > 0 {
		baseUrl = dependency.RegistryUrls[0]
		ds.logger.Debug(fmt.Sprintf("Using custom registry url: %s", baseUrl))
	}
	indexFilePath := "dl/?mode=json&include=all"
	stableOnly := strings.HasSuffix(dependency.Name, "stable")

	// Download the index file
	downloadUrl := baseUrl
	if !strings.HasSuffix(downloadUrl, "/") {
		downloadUrl += "/"
	}
	downloadUrl += indexFilePath
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
		stableValue := entry["stable"]
		if stableOnly && stableValue != true {
			continue
		}
		releases = append(releases, &shared.ReleaseInfo{
			VersionString: versionString,
		})
	}
	return releases, nil
}

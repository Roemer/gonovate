package datasources

import (
	"encoding/json"
	"fmt"
	"gonovate/core"
	"log/slog"
	"strings"
)

type GoVersionDatasource struct {
	datasourceBase
}

func NewGoVersionDatasource(logger *slog.Logger) IDatasource {
	newDatasource := &GoVersionDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   core.DATASOURCE_TYPE_GOVERSION,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GoVersionDatasource) getReleases(packageSettings *core.PackageSettings, hostRules []*core.HostRule) ([]*core.ReleaseInfo, error) {
	baseUrl := "https://go.dev"
	if len(packageSettings.RegistryUrls) > 0 {
		baseUrl = packageSettings.RegistryUrls[0]
		ds.logger.Debug(fmt.Sprintf("Using custom registry url: %s", baseUrl))
	}
	indexFilePath := "dl/?mode=json&include=all"
	stableOnly := strings.HasSuffix(packageSettings.PackageName, "stable")

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

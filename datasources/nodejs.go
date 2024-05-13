package datasources

import (
	"encoding/json"
	"fmt"
	"gonovate/core"
	"log/slog"
	"net/url"
	"strings"
	"time"
)

type NodeJsDatasource struct {
	datasourceBase
}

func NewNodeJsDatasource(logger *slog.Logger) *NodeJsDatasource {
	newDatasource := &NodeJsDatasource{}
	newDatasource.logger = logger
	newDatasource.name = core.DATASOURCE_TYPE_NODEJS
	return newDatasource
}

func (ds *NodeJsDatasource) GetReleases(packageSettings *core.PackageSettings, hostRules []*core.HostRule) ([]*core.ReleaseInfo, error) {
	baseUrl := "https://nodejs.org/dist"
	if len(packageSettings.RegistryUrls) > 0 {
		baseUrl = packageSettings.RegistryUrls[0]
		ds.logger.Debug(fmt.Sprintf("Using custom registry url: %s", baseUrl))
	}
	indexFilePath := "index.json"
	ltsOnly := strings.HasSuffix(packageSettings.PackageName, "lts")

	// Download the index file
	downloadUrl, err := url.JoinPath(baseUrl, indexFilePath)
	if err != nil {
		return nil, err
	}
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
		ltsValue := entry["lts"]
		dateString := entry["date"].(string)
		if ltsOnly && ltsValue == false {
			continue
		}
		releaseDate, err := time.Parse(time.DateOnly, dateString)
		if err != nil {
			return nil, fmt.Errorf("failed parsing date '%s': %w", dateString, err)
		}
		releases = append(releases, &core.ReleaseInfo{
			ReleaseDate:   releaseDate,
			VersionString: versionString,
		})
	}
	return releases, nil
}

package datasources

import (
	"encoding/json"
	"fmt"
	"gonovate/core"
	"log/slog"
	"net/url"
	"time"
)

type NpmDatasource struct {
	datasourceBase
}

func NewNpmDatasource(logger *slog.Logger) IDatasource {
	newDatasource := &NpmDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   core.DATASOURCE_TYPE_NPM,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *NpmDatasource) getReleases(packageSettings *core.PackageSettings, hostRules []*core.HostRule) ([]*core.ReleaseInfo, error) {
	baseUrl := "https://registry.npmjs.org"
	if len(packageSettings.RegistryUrls) > 0 {
		baseUrl = packageSettings.RegistryUrls[0]
		ds.logger.Debug(fmt.Sprintf("Using custom registry url: %s", baseUrl))
	}

	// Download the index file
	downloadUrl, err := url.JoinPath(baseUrl, packageSettings.PackageName)
	if err != nil {
		return nil, err
	}
	indexFileBytes, err := core.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return nil, err
	}

	// Parse the data as json
	var jsonData npmResponse
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*core.ReleaseInfo{}
	for _, entry := range jsonData.Versions {
		releaseInfo := &core.ReleaseInfo{
			VersionString: entry.Version,
		}
		// If possible, get a date
		if date, ok := jsonData.Time[entry.Version]; ok {
			releaseInfo.ReleaseDate = date
		}
		releases = append(releases, releaseInfo)
	}

	return releases, nil
}

type npmResponse struct {
	Versions map[string]*npmVersion `json:"versions"`
	Time     map[string]time.Time   `json:"time"`
}

type npmVersion struct {
	Version string `json:"version"`
}

package datasources

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type BrowserVersionDatasource struct {
	datasourceBase
}

func NewBrowserVersionDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &BrowserVersionDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_BROWSERVERSION,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *BrowserVersionDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	switch dependency.Name {
	case "chrome":
		return ds.chrome(dependency)
	case "chrome-for-testing":
		return ds.chromeForTesting(dependency)
	case "firefox":
		return ds.firefox(dependency)
	default:
		return nil, fmt.Errorf("unknown browser '%s'", dependency.Name)
	}
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

func (ds *BrowserVersionDatasource) chrome(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://versionhistory.googleapis.com", dependency.RegistryUrls)
	indexFilePath := "v1/chrome/platforms/linux/channels/stable/versions"

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
	var jsonData chromeVersions
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*shared.ReleaseInfo{}
	for _, entry := range jsonData.Versions {
		versionString := entry.Version
		releases = append(releases, &shared.ReleaseInfo{
			VersionString: versionString,
		})
	}

	return releases, nil
}

func (ds *BrowserVersionDatasource) chromeForTesting(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://googlechromelabs.github.io", dependency.RegistryUrls)
	indexFilePath := "chrome-for-testing/known-good-versions.json"

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
	var jsonData chromeTestingVersions
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*shared.ReleaseInfo{}
	for _, entry := range jsonData.Versions {
		versionString := entry.Version
		releases = append(releases, &shared.ReleaseInfo{
			VersionString: versionString,
		})
	}

	return releases, nil
}

func (ds *BrowserVersionDatasource) firefox(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://product-details.mozilla.org", dependency.RegistryUrls)
	indexFilePath := "1.0/firefox.json"

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
	for _, entry := range jsonData["releases"].(map[string]interface{}) {
		versionString := entry.(map[string]interface{})["version"].(string)
		releases = append(releases, &shared.ReleaseInfo{
			VersionString: versionString,
		})
	}

	return releases, nil
}

type chromeVersions struct {
	Versions []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"versions"`
	NextPageToken string `json:"nextPageToken"`
}

type chromeTestingVersions struct {
	Timestamp time.Time `json:"timestamp"`
	Versions  []struct {
		Version  string `json:"version"`
		Revision string `json:"revision"`
	} `json:"versions"`
}

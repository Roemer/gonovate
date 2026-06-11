package datasources

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/roemer/gonovate/pkg/common"
)

type NugetDatasource struct {
	*datasourceBase
}

func NewNugetDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &NugetDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_NUGET, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *NugetDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://api.nuget.org/v3/registration5-semver1", dependency.RegistryUrls)

	// NuGet package IDs in the URL must be lowercase
	packageId := strings.ToLower(dependency.Name)

	// Build the URL for the registration index
	indexUrl, err := url.JoinPath(registryUrl, packageId, "index.json")
	if err != nil {
		return nil, err
	}

	// Download and parse the registration index
	indexFileBytes, err := common.HttpUtil.DownloadToMemory(indexUrl)
	if err != nil {
		return nil, err
	}

	var registrationIndex nugetRegistrationIndex
	if err := json.Unmarshal(indexFileBytes, &registrationIndex); err != nil {
		return nil, err
	}

	// Collect catalog entries from all pages
	releases := []*common.ReleaseInfo{}
	for _, page := range registrationIndex.Items {
		var entries []nugetRegistrationPageEntry

		if len(page.Items) > 0 {
			// Items are inline in the index response
			entries = page.Items
		} else {
			// Items must be fetched from the page URL separately
			pageBytes, err := common.HttpUtil.DownloadToMemory(page.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to download registration page '%s': %w", page.ID, err)
			}
			var fullPage nugetRegistrationPage
			if err := json.Unmarshal(pageBytes, &fullPage); err != nil {
				return nil, err
			}
			entries = fullPage.Items
		}

		for _, entry := range entries {
			ce := entry.CatalogEntry
			releaseInfo := &common.ReleaseInfo{
				VersionString: ce.Version,
			}
			if !ce.Published.IsZero() {
				releaseInfo.ReleaseDate = ce.Published
			}
			releases = append(releases, releaseInfo)
		}
	}

	return releases, nil
}

type nugetRegistrationIndex struct {
	Items []nugetRegistrationPage `json:"items"`
}

type nugetRegistrationPage struct {
	ID    string                       `json:"@id"`
	Items []nugetRegistrationPageEntry `json:"items"`
}

type nugetRegistrationPageEntry struct {
	CatalogEntry nugetCatalogEntry `json:"catalogEntry"`
}

type nugetCatalogEntry struct {
	Version   string    `json:"version"`
	Published time.Time `json:"published"`
}

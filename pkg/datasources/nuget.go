package datasources

import (
	"encoding/json"
	"net/url"
	"strings"

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
	registryUrl := ds.getRegistryUrl("https://api.nuget.org/v3-flatcontainer", dependency.RegistryUrls)

	// NuGet package IDs in the URL must be lowercase
	packageId := strings.ToLower(dependency.Name)

	// Build the URL for the package versions index
	downloadUrl, err := url.JoinPath(registryUrl, packageId, "index.json")
	if err != nil {
		return nil, err
	}

	// Download the versions index file
	indexFileBytes, err := common.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return nil, err
	}

	// Parse the data as JSON
	var jsonData nugetVersionsResponse
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return nil, err
	}

	// Convert all entries to ReleaseInfo objects
	releases := []*common.ReleaseInfo{}
	for _, version := range jsonData.Versions {
		releases = append(releases, &common.ReleaseInfo{
			VersionString: version,
		})
	}

	return releases, nil
}

type nugetVersionsResponse struct {
	Versions []string `json:"versions"`
}

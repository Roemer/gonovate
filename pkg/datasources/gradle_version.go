package datasources

import (
	"encoding/json"
	"net/url"

	"github.com/roemer/gonovate/pkg/common"
)

type GradleVersionDatasource struct {
	*datasourceBase
}

func NewGradleVersionDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &GradleVersionDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_GRADLEVERSION, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GradleVersionDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://raw.githubusercontent.com", dependency.RegistryUrls)
	indexFilePath := "gradle/gradle/master/released-versions.json"

	// Download the index file
	downloadUrl, err := url.JoinPath(registryUrl, indexFilePath)
	if err != nil {
		return nil, err
	}
	indexFileBytes, err := common.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return nil, err
	}

	// Parse the data as json
	var jsonData map[string]interface{}
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*common.ReleaseInfo{}
	for _, entry := range jsonData["finalReleases"].([]interface{}) {
		versionString := entry.(map[string]interface{})["version"].(string)
		releases = append(releases, &common.ReleaseInfo{
			VersionString: versionString,
		})
	}
	return releases, nil
}

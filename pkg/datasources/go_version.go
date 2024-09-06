package datasources

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/roemer/gonovate/pkg/common"
)

type GoVersionDatasource struct {
	*datasourceBase
}

func NewGoVersionDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &GoVersionDatasource{
		datasourceBase: newDatasourceBase(settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GoVersionDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://go.dev", dependency.RegistryUrls)
	indexFilePath := "dl/?mode=json&include=all"
	stableOnly := strings.HasSuffix(dependency.Name, "stable")

	// Download the index file
	downloadUrl := fmt.Sprintf("%s/%s", registryUrl, indexFilePath)
	indexFileBytes, err := common.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return nil, err
	}

	// Parse the data as json
	var jsonData []map[string]interface{}
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*common.ReleaseInfo{}
	for _, entry := range jsonData {
		versionString := entry["version"].(string)
		stableValue := entry["stable"]
		if stableOnly && stableValue != true {
			continue
		}
		releases = append(releases, &common.ReleaseInfo{
			VersionString: versionString,
		})
	}
	return releases, nil
}

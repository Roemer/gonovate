package datasources

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/roemer/gonovate/pkg/common"
)

type NodeJsDatasource struct {
	*datasourceBase
}

func NewNodeJsDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &NodeJsDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_NODEJS, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *NodeJsDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://nodejs.org/dist", dependency.RegistryUrls)
	indexFilePath := "index.json"
	ltsOnly := strings.HasSuffix(dependency.Name, "lts")

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
	var jsonData []map[string]interface{}
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*common.ReleaseInfo{}
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
		releases = append(releases, &common.ReleaseInfo{
			ReleaseDate:   releaseDate,
			VersionString: versionString,
		})
	}
	return releases, nil
}

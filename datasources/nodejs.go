package datasources

import (
	"encoding/json"
	"fmt"
	"gonovate/core"
	"log/slog"
	"net/url"
	"strings"
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

func (ds *NodeJsDatasource) GetVersionStrings(packageName string, packageSettings *core.PackageSettings, hostRules []*core.HostRule) ([]string, error) {
	baseUrl := "https://nodejs.org/dist"
	if len(packageSettings.RegistryUrls) > 0 {
		baseUrl = packageSettings.RegistryUrls[0]
		ds.logger.Debug(fmt.Sprintf("Using custom registry url: %s", baseUrl))
	}
	indexFilePath := "index.json"
	ltsOnly := strings.HasSuffix(packageName, "lts")

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
	versions := []string{}
	for _, entry := range jsonData {
		versionString := entry["version"].(string)
		ltsValue := entry["lts"]
		if ltsOnly && ltsValue == false {
			continue
		}
		versions = append(versions, versionString)
	}
	return versions, nil
}

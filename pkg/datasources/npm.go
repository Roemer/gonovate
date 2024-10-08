package datasources

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/roemer/gonovate/pkg/common"
)

type NpmDatasource struct {
	*datasourceBase
}

func NewNpmDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &NpmDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_NPM, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *NpmDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://registry.npmjs.org", dependency.RegistryUrls)

	// Download the index file
	downloadUrl, err := url.JoinPath(registryUrl, dependency.Name)
	if err != nil {
		return nil, err
	}
	indexFileBytes, err := common.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return nil, err
	}

	// Parse the data as json
	var jsonData npmResponse
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*common.ReleaseInfo{}
	for _, entry := range jsonData.Versions {
		releaseInfo := &common.ReleaseInfo{
			VersionString:  entry.Version,
			AdditionalData: map[string]string{},
		}
		// If possible, get a date
		if date, ok := jsonData.Time[entry.Version]; ok {
			releaseInfo.ReleaseDate = date
		}
		// If possible, get checksums
		if entry.Dist != nil {
			if entry.Dist.Shasum != "" {
				releaseInfo.AdditionalData["sha1"] = entry.Dist.Shasum
			}
			if entry.Dist.Integrity != "" {
				parts := strings.SplitN(entry.Dist.Integrity, "-", 2)
				algo := parts[0]
				checksumBase64 := parts[1]
				checksumHex, err := common.Base64ToHex(checksumBase64)
				if err != nil {
					// TODO: Log warning?
				} else {
					releaseInfo.AdditionalData[algo] = checksumHex
				}
			}
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
	Dist    *struct {
		Shasum    string `json:"shasum"`
		Integrity string `json:"integrity"`
	} `json:"dist"`
}

package datasources

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type NpmDatasource struct {
	datasourceBase
}

func NewNpmDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &NpmDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_NPM,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *NpmDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	baseUrl := "https://registry.npmjs.org"
	if len(dependency.RegistryUrls) > 0 {
		baseUrl = dependency.RegistryUrls[0]
		ds.logger.Debug(fmt.Sprintf("Using custom registry url: %s", baseUrl))
	}

	// Download the index file
	downloadUrl, err := url.JoinPath(baseUrl, dependency.Name)
	if err != nil {
		return nil, err
	}
	indexFileBytes, err := shared.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return nil, err
	}

	// Parse the data as json
	var jsonData npmResponse
	if err := json.Unmarshal(indexFileBytes, &jsonData); err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*shared.ReleaseInfo{}
	for _, entry := range jsonData.Versions {
		releaseInfo := &shared.ReleaseInfo{
			VersionString: entry.Version,
			Hashes:        map[string]string{},
		}
		// If possible, get a date
		if date, ok := jsonData.Time[entry.Version]; ok {
			releaseInfo.ReleaseDate = date
		}
		// If possible, get checksums
		if entry.Dist != nil {
			if entry.Dist.Shasum != "" {
				releaseInfo.Hashes["sha1"] = entry.Dist.Shasum
			}
			if entry.Dist.Integrity != "" {
				parts := strings.SplitN(entry.Dist.Integrity, "-", 2)
				algo := parts[0]
				checksumBase64 := parts[1]
				checksumHex, err := shared.Base64ToHex(checksumBase64)
				if err != nil {
					// TODO: Log warning?
				} else {
					releaseInfo.Hashes[algo] = checksumHex
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

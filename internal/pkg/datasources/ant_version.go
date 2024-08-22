package datasources

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type AntVersionDatasource struct {
	datasourceBase
}

func NewAntVersionDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &AntVersionDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_ANTVERSION,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *AntVersionDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://archive.apache.org", dependency.RegistryUrls)
	indexFilePath := "dist/ant/binaries"

	// Download the index file
	downloadUrl, err := url.JoinPath(registryUrl, indexFilePath)
	if err != nil {
		return nil, err
	}
	indexFileBytes, err := shared.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*shared.ReleaseInfo{}
	lineRegex := regexp.MustCompile(`^.*<a href="apache-ant-([0-9\.]+)-bin.tar.gz">.*$`)
	scanner := bufio.NewScanner(bytes.NewReader(indexFileBytes))
	for scanner.Scan() {
		line := scanner.Text()
		if match := lineRegex.FindStringSubmatch(line); match != nil {
			versionString := match[1]
			releases = append(releases, &shared.ReleaseInfo{
				VersionString: versionString,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed processing the line scanner")
	}

	return releases, nil
}

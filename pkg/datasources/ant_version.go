package datasources

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"regexp"

	"github.com/roemer/gonovate/pkg/common"
)

type AntVersionDatasource struct {
	*datasourceBase
}

func NewAntVersionDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &AntVersionDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_ANTVERSION, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *AntVersionDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://archive.apache.org", dependency.RegistryUrls)
	indexFilePath := "dist/ant/binaries"

	// Download the index file
	downloadUrl, err := url.JoinPath(registryUrl, indexFilePath)
	if err != nil {
		return nil, err
	}
	indexFileBytes, err := common.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*common.ReleaseInfo{}
	lineRegex := regexp.MustCompile(`^.*<a href="apache-ant-([0-9\.]+)-bin.tar.gz">.*$`)
	scanner := bufio.NewScanner(bytes.NewReader(indexFileBytes))
	for scanner.Scan() {
		line := scanner.Text()
		if match := lineRegex.FindStringSubmatch(line); match != nil {
			versionString := match[1]
			releases = append(releases, &common.ReleaseInfo{
				VersionString: versionString,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed processing the line scanner")
	}

	return releases, nil
}

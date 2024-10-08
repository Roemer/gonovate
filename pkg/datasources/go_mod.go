package datasources

import (
	"net/url"
	"strings"

	"github.com/roemer/gonovate/pkg/common"
)

type GoModDatasource struct {
	*datasourceBase
}

func NewGoModDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &GoModDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_GOMOD, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GoModDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://proxy.golang.org", dependency.RegistryUrls)

	// Download the list of versions
	downloadUrl, err := url.JoinPath(registryUrl, dependency.Name, "@v", "list")
	if err != nil {
		return nil, err
	}
	data, err := common.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return nil, err
	}

	// Split the versons by newline
	versions := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	// Convert all entries to objects
	releases := []*common.ReleaseInfo{}
	for _, version := range versions {
		version = strings.TrimSpace(version)
		if version == "" {
			continue
		}
		releases = append(releases, &common.ReleaseInfo{
			VersionString: version,
		})
	}
	return releases, nil

	// To get pseudo-version: $base/$module/@latest
}

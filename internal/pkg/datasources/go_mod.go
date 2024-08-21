package datasources

import (
	"log/slog"
	"net/url"
	"strings"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type GoModDatasource struct {
	datasourceBase
}

func NewGoModDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &GoModDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_GOVERSION,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GoModDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	registryUrl := ds.getRegistryUrl("https://proxy.golang.org", dependency.RegistryUrls)

	// Download the list of versions
	downloadUrl, err := url.JoinPath(registryUrl, dependency.Name, "@v", "list")
	if err != nil {
		return nil, err
	}
	data, err := shared.HttpUtil.DownloadToMemory(downloadUrl)
	if err != nil {
		return nil, err
	}

	// Split the versons by newline
	versions := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	// Convert all entries to objects
	releases := []*shared.ReleaseInfo{}
	for _, version := range versions {
		version = strings.TrimSpace(version)
		if version == "" {
			continue
		}
		releases = append(releases, &shared.ReleaseInfo{
			VersionString: version,
		})
	}
	return releases, nil

	// To get pseudo-version: $base/$module/@latest
}

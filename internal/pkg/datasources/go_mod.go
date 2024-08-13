package datasources

import (
	"fmt"
	"log/slog"
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
	baseUrl := "https://proxy.golang.org"
	if len(dependency.RegistryUrls) > 0 {
		baseUrl = dependency.RegistryUrls[0]
		ds.logger.Debug(fmt.Sprintf("Using custom registry url: %s", baseUrl))
	}

	// Download the list of versions
	url := fmt.Sprintf("%s/%s/@v/list", baseUrl, dependency.Name)
	data, err := shared.HttpUtil.DownloadToMemory(url)
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

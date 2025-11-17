package datasources

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/roemer/gonovate/pkg/common"
)

type HelmDatasource struct {
	*datasourceBase
}

func NewHelmDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &HelmDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_HELM, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *HelmDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	// Get index.yaml file
	helmRepository := dependency.RegistryUrls[0]

	// TODO:
	if strings.HasPrefix(helmRepository, "file:") {
		return nil, nil
	}
	if strings.HasPrefix(helmRepository, "oci:") {
		return nil, nil
	}

	// Todo: Cache
	indexUrl, err := url.JoinPath(helmRepository, "index.yaml")
	if err != nil {
		return nil, err
	}
	ds.logger.Debug(fmt.Sprintf("Fetching index from %s", indexUrl))
	indexBytes, err := common.HttpUtil.DownloadToMemory(indexUrl)
	if err != nil {
		return nil, err
	}

	indexEntries := struct {
		Entries map[string][]struct {
			Name    string    `yaml:"name"`
			Version string    `yaml:"version"`
			Created time.Time `yaml:"created"`
			Digest  string    `yaml:"digest"`
		} `yaml:"entries"`
	}{}

	if err := yaml.Unmarshal([]byte(indexBytes), &indexEntries); err != nil {
		return nil, fmt.Errorf("failed unmarshalling index.yaml")
	}

	releases := []*common.ReleaseInfo{}
	dependencyEntries, ok := indexEntries.Entries[dependency.Name]
	if !ok {
		return releases, nil
	}
	for _, entry := range dependencyEntries {
		if entry.Name != dependency.Name {
			continue
		}
		newRelease := &common.ReleaseInfo{
			VersionString: entry.Version,
			ReleaseDate:   entry.Created,
			Digest:        entry.Digest,
		}
		releases = append(releases, newRelease)
	}
	return releases, nil
}

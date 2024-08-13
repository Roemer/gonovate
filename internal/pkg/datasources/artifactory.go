package datasources

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	artifactory_config "github.com/jfrog/jfrog-client-go/config"
	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type ArtifactoryDatasource struct {
	datasourceBase
}

func NewArtifactoryDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &ArtifactoryDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_ARTIFACTORY,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *ArtifactoryDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	// Get the base url for artifactory
	if len(dependency.RegistryUrls) == 0 {
		return nil, fmt.Errorf("no registry url for Artifactory for dependencyName '%s'", dependency.Name)
	}
	registryUrl := dependency.RegistryUrls[0]

	// Get a host rule if any was defined
	relevantHostRule := ds.Config.FilterHostConfigsForHost(registryUrl)
	token := ""
	if relevantHostRule != nil {
		token = relevantHostRule.TokendExpanded()
	}

	// Create the client
	artifactoryManager, err := ds.createManager(registryUrl, token)
	if err != nil {
		return nil, err
	}

	// Search with the pattern
	params := services.NewSearchParams()
	params.Pattern = dependency.Name
	items, err := ds.getSearchResults(artifactoryManager, params)
	if err != nil {
		return nil, err
	}

	// Build the list of releases
	releases := []*shared.ReleaseInfo{}
	for _, item := range items {
		releases = append(releases, &shared.ReleaseInfo{
			VersionString: item.Name,
			ReleaseDate:   item.Modified,
			Hashes: map[string]string{
				"md5":    item.Md5,
				"sha256": item.Sha256,
			},
		})
	}
	return releases, nil
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

type artifactorySearchResultItem struct {
	Repo       string                            `json:"repo"`
	Path       string                            `json:"path"`
	Name       string                            `json:"name"`
	Created    time.Time                         `json:"created"`
	Modified   time.Time                         `json:"modified"`
	Type       string                            `json:"type"`
	Size       int                               `json:"size"`
	Md5        string                            `json:"actual_md5"`
	Sha1       string                            `json:"actual_sha1"`
	Sha256     string                            `json:"sha256"`
	Properties []artifactorySearchResultProperty `json:"properties"`
}

type artifactorySearchResultProperty struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (ds *ArtifactoryDatasource) createManager(baseUrl string, token string) (artifactory.ArtifactoryServicesManager, error) {
	artifactoryDetails := auth.NewArtifactoryDetails()
	artifactoryDetails.SetUrl(baseUrl)
	if len(token) > 0 {
		artifactoryDetails.SetAccessToken(token)
	}

	configBuilder, err := artifactory_config.NewConfigBuilder().
		SetServiceDetails(artifactoryDetails).
		Build()
	if err != nil {
		return nil, err
	}

	artifactoryManager, err := artifactory.New(configBuilder)
	return artifactoryManager, err
}

func (ds *ArtifactoryDatasource) getSearchResults(artifactoryManager artifactory.ArtifactoryServicesManager, searchParams services.SearchParams) ([]*artifactorySearchResultItem, error) {
	searchResultItems := []*artifactorySearchResultItem{}

	reader, err := artifactoryManager.SearchFiles(searchParams)
	if err != nil {
		return searchResultItems, err
	}
	defer reader.Close()

	// Read the items from the reader
	if reader != nil {
		for searchResultItem := new(artifactorySearchResultItem); reader.NextRecord(searchResultItem) == nil; searchResultItem = new(artifactorySearchResultItem) {
			searchResultItems = append(searchResultItems, searchResultItem)
		}
	}

	return searchResultItems, nil
}

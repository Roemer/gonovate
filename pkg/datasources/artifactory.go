package datasources

import (
	"fmt"
	"time"

	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	artifactory_config "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/roemer/gonovate/pkg/common"
)

type ArtifactoryDatasource struct {
	*datasourceBase
}

func NewArtifactoryDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &ArtifactoryDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_ARTIFACTORY, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *ArtifactoryDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	// Get the base url for artifactory
	if len(dependency.RegistryUrls) == 0 {
		return nil, fmt.Errorf("no registry url for Artifactory for dependencyName '%s'", dependency.Name)
	}
	registryUrl := dependency.RegistryUrls[0]

	// Get a host rule if any was defined
	relevantHostRule := ds.getHostRuleForHost(registryUrl)
	token := ""
	user := ""
	password := ""
	if relevantHostRule != nil {
		token = relevantHostRule.TokendExpanded()
		user = relevantHostRule.Username
		password = relevantHostRule.Password
	}

	// Create the client
	artifactoryManager, err := ds.createManager(registryUrl, token, user, password)
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
	releases := []*common.ReleaseInfo{}
	for _, item := range items {
		releases = append(releases, &common.ReleaseInfo{
			VersionString: item.Name,
			ReleaseDate:   item.Modified,
			AdditionalData: map[string]string{
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

func (ds *ArtifactoryDatasource) createManager(baseUrl string, token string, user string, password string) (artifactory.ArtifactoryServicesManager, error) {
	artifactoryDetails := auth.NewArtifactoryDetails()
	artifactoryDetails.SetUrl(baseUrl)

	// Set authentication info
	if len(user) > 0 {
		artifactoryDetails.SetUser(user)
	}
	if len(password) > 0 {
		artifactoryDetails.SetPassword(password)
	}
	if len(token) > 0 {
		if httpclient.IsApiKey(token) {
			artifactoryDetails.SetApiKey(token)
		} else {
			artifactoryDetails.SetAccessToken(token)
		}
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

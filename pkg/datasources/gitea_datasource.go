package datasources

import (
	"fmt"
	"net/url"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/roemer/gonovate/pkg/common"
)

type GiteaReleasesDatasource struct {
	*datasourceBase
}

func NewGiteaReleasesDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &GiteaReleasesDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_GITEA_RELEASES, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GiteaReleasesDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	owner, repository, err := splitGiteaRepository(dependency.Name)
	if err != nil {
		return nil, err
	}

	client, err := ds.createClient(dependency.RegistryUrls)
	if err != nil {
		return nil, err
	}

	allReleases := []*gitea.Release{}
	listOptions := gitea.ListReleasesOptions{
		ListOptions: gitea.ListOptions{Page: 1, PageSize: 100},
		IsDraft:     common.FalsePtr,
	}
	for {
		giteaReleases, resp, err := client.ListReleases(owner, repository, listOptions)
		if err != nil {
			return nil, err
		}
		allReleases = append(allReleases, giteaReleases...)
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}

	releases := make([]*common.ReleaseInfo, 0, len(allReleases))
	for _, entry := range allReleases {
		releaseDate := entry.PublishedAt
		if releaseDate.IsZero() {
			releaseDate = entry.CreatedAt
		}
		releases = append(releases, &common.ReleaseInfo{
			VersionString: entry.TagName,
			ReleaseDate:   releaseDate,
		})
	}

	return releases, nil
}

func (ds *GiteaReleasesDatasource) createClient(registryUrls []string) (*gitea.Client, error) {
	endpoint, err := normalizeGiteaEndpoint(ds.getRegistryUrl("https://gitea.com", registryUrls))
	if err != nil {
		return nil, err
	}

	parsedEndpoint, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed parsing gitea endpoint '%s': %w", endpoint, err)
	}

	options := []gitea.ClientOption{}
	if relevantHostRule := ds.getHostRuleForHost(parsedEndpoint.Host); relevantHostRule != nil {
		token := relevantHostRule.TokenExpanded()
		if token != "" {
			options = append(options, gitea.SetToken(token))
		}
	}

	return gitea.NewClient(endpoint, options...)
}

func splitGiteaRepository(dependencyName string) (string, string, error) {
	parts := strings.SplitN(dependencyName, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("gitea dependency '%s' must be in the format 'owner/repository'", dependencyName)
	}
	return parts[0], parts[1], nil
}

func normalizeGiteaEndpoint(rawUrl string) (string, error) {
	if rawUrl != "" && !httpSchemeRegex.MatchString(rawUrl) {
		rawUrl = "https://" + rawUrl
	}

	parsed, err := url.Parse(rawUrl)
	if err != nil {
		return "", fmt.Errorf("failed parsing gitea registry url '%s': %w", rawUrl, err)
	}

	path := strings.TrimSuffix(parsed.Path, "/")
	path = strings.TrimSuffix(path, "/api/v1")
	parsed.Path = path
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return strings.TrimSuffix(parsed.String(), "/"), nil
}

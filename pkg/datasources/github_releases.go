package datasources

import (
	"context"
	"strings"

	"github.com/google/go-github/v81/github"
	"github.com/roemer/gonovate/pkg/common"
)

type GitHubReleasesDatasource struct {
	*datasourceBase
}

func NewGitHubReleasesDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &GitHubReleasesDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_GITHUB_RELEASES, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GitHubReleasesDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	client := getGitHubClient(ds.datasourceBase)

	parts := strings.SplitN(dependency.Name, "/", 2)
	owner := parts[0]
	repository := parts[1]

	allReleases := []*github.RepositoryRelease{}
	listOptions := &github.ListOptions{PerPage: 100}
	for {
		gitHubReleases, resp, err := client.Repositories.ListReleases(context.Background(), owner, repository, listOptions)
		if err != nil {
			return nil, err
		}
		allReleases = append(allReleases, gitHubReleases...)
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}

	// Convert all entries to objects
	releases := []*common.ReleaseInfo{}
	for _, entry := range allReleases {
		versionString := *entry.Name
		releases = append(releases, &common.ReleaseInfo{
			VersionString: versionString,
			ReleaseDate:   entry.PublishedAt.Time,
		})
	}
	return releases, nil
}

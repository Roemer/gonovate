package datasources

import (
	"context"
	"strings"

	"github.com/google/go-github/v80/github"
	"github.com/roemer/gonovate/pkg/common"
)

type GitHubTagsDatasource struct {
	*datasourceBase
}

func NewGitHubTagsDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &GitHubTagsDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_GITHUB_TAGS, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GitHubTagsDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	client := getGitHubClient(ds.datasourceBase)

	parts := strings.SplitN(dependency.Name, "/", 2)
	owner := parts[0]
	repository := parts[1]

	allTags := []*github.RepositoryTag{}
	listOptions := &github.ListOptions{PerPage: 100}
	for {
		gitHubTags, resp, err := client.Repositories.ListTags(context.Background(), owner, repository, listOptions)
		if err != nil {
			return nil, err
		}
		allTags = append(allTags, gitHubTags...)
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}

	// Convert all entries to objects
	releases := []*common.ReleaseInfo{}
	for _, entry := range allTags {
		versionString := *entry.Name
		releases = append(releases, &common.ReleaseInfo{
			VersionString: versionString,
		})
	}
	return releases, nil
}

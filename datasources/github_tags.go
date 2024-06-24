package datasources

import (
	"context"
	"gonovate/core"
	"log/slog"
	"strings"

	"github.com/google/go-github/v62/github"
)

type GitHubTagsDatasource struct {
	datasourceBase
}

func NewGitHubTagsDatasource(logger *slog.Logger, config *core.Config) IDatasource {
	newDatasource := &GitHubTagsDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   core.DATASOURCE_TYPE_GITHUB_TAGS,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GitHubTagsDatasource) getReleases(packageSettings *core.PackageSettings) ([]*core.ReleaseInfo, error) {
	client := github.NewClient(nil)
	// TODO: WithAuthToken(token)

	parts := strings.SplitN(packageSettings.PackageName, "/", 2)
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
	releases := []*core.ReleaseInfo{}
	for _, entry := range allTags {
		versionString := *entry.Name
		releases = append(releases, &core.ReleaseInfo{
			VersionString: versionString,
		})
	}
	return releases, nil
}

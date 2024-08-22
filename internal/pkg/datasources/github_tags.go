package datasources

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/go-github/v63/github"
	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type GitHubTagsDatasource struct {
	datasourceBase
}

func NewGitHubTagsDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &GitHubTagsDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_GITHUB_TAGS,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GitHubTagsDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
	client := github.NewClient(nil)

	// Get a host rule if any was defined
	relevantHostRule := ds.Config.FilterHostConfigsForHost("api.github.com")
	// Add the token to the client
	if relevantHostRule != nil {
		token := relevantHostRule.TokendExpanded()
		client = client.WithAuthToken(token)
	}

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
	releases := []*shared.ReleaseInfo{}
	for _, entry := range allTags {
		versionString := *entry.Name
		releases = append(releases, &shared.ReleaseInfo{
			VersionString: versionString,
		})
	}
	return releases, nil
}

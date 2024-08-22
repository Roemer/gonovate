package datasources

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/go-github/v63/github"
	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type GitHubReleasesDatasource struct {
	datasourceBase
}

func NewGitHubReleasesDatasource(logger *slog.Logger, config *config.RootConfig) IDatasource {
	newDatasource := &GitHubReleasesDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   shared.DATASOURCE_TYPE_GITHUB_RELEASES,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GitHubReleasesDatasource) getReleases(dependency *shared.Dependency) ([]*shared.ReleaseInfo, error) {
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
	releases := []*shared.ReleaseInfo{}
	for _, entry := range allReleases {
		versionString := *entry.Name
		releases = append(releases, &shared.ReleaseInfo{
			VersionString: versionString,
			ReleaseDate:   entry.PublishedAt.Time,
		})
	}
	return releases, nil
}

package datasources

import (
	"context"
	"gonovate/core"
	"log/slog"
	"strings"

	"github.com/google/go-github/v62/github"
)

type GitHubReleasesDatasource struct {
	datasourceBase
}

func NewGitHubReleasesDatasource(logger *slog.Logger, config *core.Config) IDatasource {
	newDatasource := &GitHubReleasesDatasource{
		datasourceBase: datasourceBase{
			logger: logger,
			name:   core.DATASOURCE_TYPE_GITHUB_RELEASES,
			Config: config,
		},
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GitHubReleasesDatasource) getReleases(packageSettings *core.PackageSettings) ([]*core.ReleaseInfo, error) {
	client := github.NewClient(nil)
	// TODO: WithAuthToken(token)

	parts := strings.SplitN(packageSettings.PackageName, "/", 2)
	owner := parts[0]
	repository := parts[1]

	gitHubReleases, _, err := client.Repositories.ListReleases(context.Background(), owner, repository, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*core.ReleaseInfo{}
	for _, entry := range gitHubReleases {
		versionString := *entry.Name
		releases = append(releases, &core.ReleaseInfo{
			VersionString: versionString,
			ReleaseDate:   entry.PublishedAt.Time,
		})
	}
	return releases, nil
}

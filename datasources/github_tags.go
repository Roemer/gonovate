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

	gitHubTags, _, err := client.Repositories.ListTags(context.Background(), owner, repository, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Convert all entries to objects
	releases := []*core.ReleaseInfo{}
	for _, entry := range gitHubTags {
		versionString := *entry.Name
		releases = append(releases, &core.ReleaseInfo{
			VersionString: versionString,
		})
	}
	return releases, nil
}

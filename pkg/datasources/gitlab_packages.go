package datasources

import (
	"strings"

	"github.com/roemer/gonovate/pkg/common"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GitLabPackagesDatasource struct {
	*datasourceBase
}

func NewGitLabPackagesDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &GitLabPackagesDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_GITLAB_PACKAGES, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GitLabPackagesDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	parts := strings.SplitN(dependency.Name, ":", 2)
	projectPath := parts[0]
	packageName := parts[1]

	client, err := ds.createClient(dependency.RegistryUrls)
	if err != nil {
		return nil, err
	}

	gitLabPackages, _, err := client.Packages.ListProjectPackages(projectPath, &gitlab.ListProjectPackagesOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
		PackageName: gitlab.Ptr(packageName),
	})
	if err != nil {
		return nil, err
	}

	releases := []*common.ReleaseInfo{}
	for _, gitLabPackage := range gitLabPackages {
		releases = append(releases, &common.ReleaseInfo{
			ReleaseDate:   *gitLabPackage.CreatedAt,
			VersionString: gitLabPackage.Version,
		})
	}

	return releases, nil
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

func (ds *GitLabPackagesDatasource) createClient(registryUrls []string) (*gitlab.Client, error) {
	registryUrl := ds.getRegistryUrl("https://gitlab.com/api/v4", registryUrls)

	// Get a host rule if any was defined
	relevantHostRule := ds.getHostRuleForHost(registryUrl)
	token := ""
	if relevantHostRule != nil {
		token = relevantHostRule.TokendExpanded()
	}

	// Create the client
	return gitlab.NewClient(token, gitlab.WithBaseURL(registryUrl))
}

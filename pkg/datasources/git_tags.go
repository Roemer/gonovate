package datasources

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/roemer/goext"
	"github.com/roemer/gonovate/pkg/common"
)

type GitTagsDatasource struct {
	*datasourceBase
}

func NewGitTagsDatasource(settings *common.DatasourceSettings) common.IDatasource {
	newDatasource := &GitTagsDatasource{
		datasourceBase: newDatasourceBase(common.DATASOURCE_TYPE_GIT_TAGS, settings),
	}
	newDatasource.impl = newDatasource
	return newDatasource
}

func (ds *GitTagsDatasource) GetReleases(dependency *common.Dependency) ([]*common.ReleaseInfo, error) {
	gitTagsStdout, gitTagsStderr, err := goext.CmdRunners.Default.RunGetOutput("git", "ls-remote", "--tags", dependency.Name)
	if err != nil {
		return nil, fmt.Errorf("failed getting git tags: %v - %s", err, gitTagsStderr)
	}

	lineRegexp := regexp.MustCompile(`^[a-f0-9]+\s+refs/tags/(.*)$`)
	releases := []*common.ReleaseInfo{}
	scanner := bufio.NewScanner(strings.NewReader(gitTagsStdout))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, "^{}") {
			continue
		}
		if matches := lineRegexp.FindStringSubmatch(line); matches != nil {
			tagName := matches[1]
			releases = append(releases, &common.ReleaseInfo{
				VersionString: tagName,
			})
		}
	}

	return releases, nil
}

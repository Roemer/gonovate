package platforms

import (
	"fmt"
	"net/url"

	"code.gitea.io/sdk/gitea"
	"github.com/roemer/gonovate/pkg/common"
)

type GiteaPlatform struct {
	*GitPlatform
}

func NewGiteaPlatform(settings *common.PlatformSettings) *GiteaPlatform {
	platform := &GiteaPlatform{
		GitPlatform: NewGitPlatform(settings),
	}
	return platform
}

func (p *GiteaPlatform) Type() common.PlatformType {
	return common.PLATFORM_TYPE_GITEA
}

func (p *GiteaPlatform) FetchProject(project *common.Project) error {
	// Prepare the data for the API
	owner, repository := project.SplitPath()
	// Create the client
	client, err := p.createClient()
	if err != nil {
		return err
	}
	// Get the repository
	platformProject, _, err := client.GetRepo(owner, repository)
	if err != nil {
		return err
	}
	if platformProject == nil {
		return fmt.Errorf("could not find project: %s", project.Path)
	}
	cloneUrl := platformProject.CloneURL
	cloneUrlWithCredentials, err := url.Parse(cloneUrl)
	if err != nil {
		return err
	}
	cloneUrlWithCredentials.User = url.UserPassword("oauth2", p.settings.TokendExpanded())
	_, _, err = common.Git.Run("clone", cloneUrlWithCredentials.String(), ".gonovate-clone")
	return err
}

func (p *GiteaPlatform) NotifyChanges(project *common.Project, updateGroup *common.UpdateGroup) error {
	return nil
}

func (p *GiteaPlatform) Cleanup(cleanupSettings *PlatformCleanupSettings) error {
	return nil
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

func (p *GiteaPlatform) createClient() (*gitea.Client, error) {
	if p.settings == nil || p.settings.Token == "" {
		return nil, fmt.Errorf("no platform token defined")
	}
	endpoint := "https://gitea.com/api"
	token := p.settings.TokendExpanded()
	if p.settings.Endpoint != "" {
		endpoint = p.settings.Endpoint
	}
	return gitea.NewClient(endpoint, gitea.SetToken(token))
}

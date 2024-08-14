package platforms

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
	"github.com/xanzy/go-gitlab"
)

type GitlabPlatform struct {
	GitPlatform
}

func NewGitlabPlatform(logger *slog.Logger, config *config.RootConfig) *GitlabPlatform {
	platform := &GitlabPlatform{
		GitPlatform: *NewGitPlatform(logger, config),
	}
	return platform
}

func (p *GitlabPlatform) Type() shared.PlatformType {
	return shared.PLATFORM_TYPE_GITLAB
}

func (p *GitlabPlatform) FetchProject(project *shared.Project) error {
	client, err := p.createClient()
	if err != nil {
		return err
	}
	platformProject, _, err := client.Projects.GetProject(project.Path, &gitlab.GetProjectOptions{})
	if err != nil {
		return err
	}
	if platformProject == nil {
		return fmt.Errorf("could not find project: %s", project.Path)
	}
	cloneUrl := platformProject.HTTPURLToRepo
	cloneUrlWithCredentials, err := url.Parse(cloneUrl)
	if err != nil {
		return err
	}
	cloneUrlWithCredentials.User = url.UserPassword("oauth2", p.Config.PlatformSettings.TokendExpanded())
	_, _, err = shared.Git.Run("clone", cloneUrlWithCredentials.String(), ".gonovate-clone")
	return err
}

func (p *GitlabPlatform) NotifyChanges(project *shared.Project, updateGroup *shared.UpdateGroup) error {
	// Create the client
	client, err := p.createClient()
	if err != nil {
		return err
	}

	mergeRequests, _, err := client.MergeRequests.ListProjectMergeRequests(project.Path, &gitlab.ListProjectMergeRequestsOptions{
		SourceBranch: gitlab.Ptr(updateGroup.BranchName),
		TargetBranch: gitlab.Ptr(p.Config.PlatformSettings.BaseBranch),
		State:        gitlab.Ptr("opened"),
	})
	if err != nil {
		return err
	}
	if len(mergeRequests) > 0 {
		p.logger.Info(fmt.Sprintf("PR already exists: %s", mergeRequests[0].WebURL))
	} else {
		mr, _, err := client.MergeRequests.CreateMergeRequest(project.Path, &gitlab.CreateMergeRequestOptions{
			Title:        gitlab.Ptr(updateGroup.Title),
			SourceBranch: gitlab.Ptr(updateGroup.BranchName),
			TargetBranch: gitlab.Ptr(p.Config.PlatformSettings.BaseBranch),
		})
		if err != nil {
			return err
		}
		p.logger.Info(fmt.Sprintf("Created PR: %s", mr.WebURL))
	}

	return nil
}

func (p *GitlabPlatform) createClient() (*gitlab.Client, error) {
	if p.Config.PlatformSettings == nil || p.Config.PlatformSettings.Token == "" {
		return nil, fmt.Errorf("no platform token defined")
	}
	endpoint := "https://gitlab.com/api/v4"
	token := p.Config.PlatformSettings.TokendExpanded()
	if p.Config.PlatformSettings.Endpoint != "" {
		endpoint = p.Config.PlatformSettings.Endpoint
	}
	return gitlab.NewClient(token, gitlab.WithBaseURL(endpoint))
}
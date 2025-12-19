package platforms

import (
	"fmt"
	"net/url"
	"slices"

	"code.gitea.io/sdk/gitea"
	"github.com/roemer/gonovate/pkg/common"
	"github.com/samber/lo"
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
	_, _, err = common.Git.Run("clone", cloneUrlWithCredentials.String(), ClonePath)
	return err
}

func (p *GiteaPlatform) LookupAuthor() (string, string, error) {
	client, err := p.createClient()
	if err != nil {
		return "", "", err
	}
	user, _, err := client.GetMyUserInfo()
	if err != nil {
		return "", "", err
	}
	return user.FullName, user.Email, nil
}

func (p *GiteaPlatform) NotifyChanges(project *common.Project, updateGroup *common.UpdateGroup) error {
	// Prepare the data for the API
	owner, repository := project.SplitPath()

	// Create the client
	client, err := p.createClient()
	if err != nil {
		return err
	}

	// Build the content of the PR
	content := ""
	for _, dep := range updateGroup.Dependencies {
		content += fmt.Sprintf("- %s from %s to %s\n", dep.Name, dep.Version, dep.NewRelease.VersionString)
	}

	// Search for an existing PR
	pullRequests, _, err := client.ListRepoPullRequests(owner, repository, gitea.ListPullRequestsOptions{
		State: gitea.StateOpen,
	})
	if err != nil {
		return err
	}
	existingPr, prExists := lo.Find(pullRequests, func(pr *gitea.PullRequest) bool {
		return pr.Head.Ref == updateGroup.BranchName && pr.Base.Ref == p.settings.BaseBranch
	})
	if prExists {
		p.logger.Info(fmt.Sprintf("PR already exists: %s", existingPr.HTMLURL))

		// Update the PR if something changed
		if existingPr.Title != updateGroup.Title || existingPr.Body != content {
			p.logger.Debug("Updating PR")
			if _, _, err := client.EditPullRequest(owner, repository, existingPr.Index, gitea.EditPullRequestOption{
				Title: updateGroup.Title,
				Body:  gitea.OptionalString(content),
			}); err != nil {
				return err
			}
		}
	} else {
		// Create the PR
		pr, _, err := client.CreatePullRequest(owner, repository, gitea.CreatePullRequestOption{
			Title: updateGroup.Title,
			Body:  content,
			Head:  updateGroup.BranchName,
			Base:  p.settings.BaseBranch,
		})
		if err != nil {
			return err
		}
		p.logger.Info(fmt.Sprintf("Created PR: %s", pr.HTMLURL))
	}
	return nil
}

func (p *GiteaPlatform) Cleanup(cleanupSettings *PlatformCleanupSettings) error {
	remoteName := p.getRemoteName()

	// Get the remote branches for gonovate
	gonovateBranches, err := p.getRemoteGonovateBranches(remoteName, cleanupSettings.BranchPrefix)
	if err != nil {
		return err
	}

	// Get the branches that were used in this gonovate run
	usedBranches := lo.FlatMap(cleanupSettings.UpdateGroups, func(x *common.UpdateGroup, _ int) []string {
		return []string{x.BranchName}
	})

	// Prepare the data for the API
	owner, repository := cleanupSettings.Project.SplitPath()

	// Create the client
	client, err := p.createClient()
	if err != nil {
		return err
	}

	// Loop thru the branches and check if they are active or not
	activeBranchCount := 0
	obsoleteBranchCount := 0
	for _, potentialStaleBranch := range gonovateBranches {
		if slices.Contains(usedBranches, potentialStaleBranch) {
			// This branch is used
			activeBranchCount++
			continue
		}
		// This branch is unused, delete the branch and a possible associated PR
		p.logger.Info(fmt.Sprintf("Removing unused branch '%s'", potentialStaleBranch))

		// Search for an existing PR
		existingRequests, _, err := client.ListRepoPullRequests(owner, repository, gitea.ListPullRequestsOptions{
			State: gitea.StateOpen,
		})
		if err != nil {
			return err
		}
		existingPr, prExists := lo.Find(existingRequests, func(pr *gitea.PullRequest) bool {
			return pr.Head.Ref == potentialStaleBranch && pr.Base.Ref == cleanupSettings.BaseBranch
		})
		if prExists {
			// Close the PR
			p.logger.Info(fmt.Sprintf("Closing associated PR: %s", existingPr.HTMLURL))
			if _, _, err := client.EditPullRequest(owner, repository, existingPr.Index, gitea.EditPullRequestOption{
				State: &[]gitea.StateType{gitea.StateClosed}[0],
			}); err != nil {
				return err
			}
		}

		// Delete the unused branch
		p.logger.Debug("Deleting the branch")
		if _, _, err := common.Git.Run("push", remoteName, "--delete", potentialStaleBranch); err != nil {
			return fmt.Errorf("failed to delete the remote branch '%s'", potentialStaleBranch)
		}
		obsoleteBranchCount++
	}

	p.logger.Info(fmt.Sprintf("Finished cleaning branches. Active: %d, Deleted: %d", activeBranchCount, obsoleteBranchCount))

	return nil
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

func (p *GiteaPlatform) createClient() (*gitea.Client, error) {
	if p.settings == nil || p.settings.Token == "" {
		return nil, fmt.Errorf("no platform token defined")
	}
	endpoint := "https://gitea.com"
	token := p.settings.TokendExpanded()
	if p.settings.Endpoint != "" {
		endpoint = p.settings.EndpointExpanded()
	}
	return gitea.NewClient(endpoint, gitea.SetToken(token))
}

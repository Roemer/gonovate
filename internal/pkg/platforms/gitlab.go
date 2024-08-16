package platforms

import (
	"fmt"
	"log/slog"
	"net/url"
	"slices"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
	"github.com/samber/lo"
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

	// Build the content of the MR
	content := ""
	for _, dep := range updateGroup.Dependencies {
		content += fmt.Sprintf("- %s from %s to %s\n", dep.Name, dep.Version, dep.NewRelease.VersionString)
	}

	// Search for an existing MR
	mergeRequests, _, err := client.MergeRequests.ListProjectMergeRequests(project.Path, &gitlab.ListProjectMergeRequestsOptions{
		SourceBranch: gitlab.Ptr(updateGroup.BranchName),
		TargetBranch: gitlab.Ptr(p.Config.PlatformSettings.BaseBranch),
		State:        gitlab.Ptr("opened"),
	})
	if err != nil {
		return err
	}

	if len(mergeRequests) > 0 {
		p.logger.Info(fmt.Sprintf("MR already exists: %s", mergeRequests[0].WebURL))

		// Update the MR if something changed
		if mergeRequests[0].Title != updateGroup.Title || mergeRequests[0].Description != content {
			p.logger.Debug("Updating MR")
			if _, _, err := client.MergeRequests.UpdateMergeRequest(project.Path, mergeRequests[0].IID, &gitlab.UpdateMergeRequestOptions{
				Title:       gitlab.Ptr(updateGroup.Title),
				Description: gitlab.Ptr(content),
			}); err != nil {
				return err
			}
		}
	} else {
		mr, _, err := client.MergeRequests.CreateMergeRequest(project.Path, &gitlab.CreateMergeRequestOptions{
			Title:        gitlab.Ptr(updateGroup.Title),
			Description:  gitlab.Ptr(content),
			SourceBranch: gitlab.Ptr(updateGroup.BranchName),
			TargetBranch: gitlab.Ptr(p.Config.PlatformSettings.BaseBranch),
		})
		if err != nil {
			return err
		}
		p.logger.Info(fmt.Sprintf("Created MR: %s", mr.WebURL))
	}

	return nil
}

func (p *GitlabPlatform) Cleanup(cleanupSettings *PlatformCleanupSettings) error {
	remoteName := p.getRemoteName()

	// Get the remote branches for gonovate
	gonovateBranches, err := p.getRemoteGonovateBranches(remoteName, cleanupSettings.BranchPrefix)
	if err != nil {
		return err
	}

	// Get the branches that were used in this gonovate run
	usedBranches := lo.FlatMap(cleanupSettings.UpdateGroups, func(x *shared.UpdateGroup, _ int) []string {
		return []string{x.BranchName}
	})

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
		// This branch is unused, delete the branch and a possible associated MR
		p.logger.Info(fmt.Sprintf("Removing unused branch '%s'", potentialStaleBranch))

		// Search for an existing MR
		mergeRequests, _, err := client.MergeRequests.ListProjectMergeRequests(cleanupSettings.Project.Path, &gitlab.ListProjectMergeRequestsOptions{
			SourceBranch: gitlab.Ptr(potentialStaleBranch),
			TargetBranch: gitlab.Ptr(cleanupSettings.BaseBranch),
			State:        gitlab.Ptr("opened"),
		})
		if err != nil {
			return err
		}
		// Close all MRs
		for _, mr := range mergeRequests {
			p.logger.Info(fmt.Sprintf("Closing associated MR: %s", mr.WebURL))
			if _, _, err := client.MergeRequests.UpdateMergeRequest(cleanupSettings.Project.Path, mr.IID, &gitlab.UpdateMergeRequestOptions{
				StateEvent: gitlab.Ptr("close"),
			}); err != nil {
				return err
			}
		}

		// Delete the unused branch
		p.logger.Debug("Deleting the branch")
		if _, _, err := shared.Git.Run("push", remoteName, "--delete", potentialStaleBranch); err != nil {
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

package platforms

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/samber/lo"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GitlabPlatform struct {
	*GitPlatform
}

func NewGitlabPlatform(settings *common.PlatformSettings) *GitlabPlatform {
	platform := &GitlabPlatform{
		GitPlatform: NewGitPlatform(settings),
	}
	platform.impl = platform
	return platform
}

func (p *GitlabPlatform) Type() common.PlatformType {
	return common.PLATFORM_TYPE_GITLAB
}

func (p *GitlabPlatform) FetchProject(project *common.Project) error {
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
	cloneUrlWithCredentials.User = url.UserPassword("oauth2", p.settings.TokenExpanded())
	_, _, err = common.Git.Run("clone", cloneUrlWithCredentials.String(), ClonePath)
	return err
}

func (p *GitlabPlatform) LookupAuthor() (string, string, error) {
	client, err := p.createClient()
	if err != nil {
		return "", "", err
	}
	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return "", "", err
	}
	return user.Name, user.Email, nil
}

func (p *GitlabPlatform) NotifyChanges(project *common.Project, updateGroup *common.UpdateGroup) error {
	// Create the client
	client, err := p.createClient()
	if err != nil {
		return err
	}

	// Build the content of the MR
	content := ""
	for _, dep := range updateGroup.Dependencies {
		content += fmt.Sprintf("- %s from %s to %s\n", dep.Dependency.Name, dep.Dependency.Version, dep.NewRelease.VersionString)
	}
	// Trim spaces / newlines
	content = strings.TrimSpace(content)

	// Search for an existing MR
	mergeRequests, _, err := client.MergeRequests.ListProjectMergeRequests(project.Path, &gitlab.ListProjectMergeRequestsOptions{
		SourceBranch: gitlab.Ptr(updateGroup.BranchName),
		TargetBranch: gitlab.Ptr(p.settings.BaseBranch),
		State:        gitlab.Ptr("opened"),
	})
	if err != nil {
		return err
	}

	if len(mergeRequests) > 0 {
		p.logger.Info(fmt.Sprintf("MR already exists: %s", mergeRequests[0].WebURL))

		// Calculate the new reviewer list
		existingReviewerNames := lo.Map(mergeRequests[0].Reviewers, func(reviewer *gitlab.BasicUser, _ int) string {
			return reviewer.Username
		})
		existingReviewerIds := lo.Map(mergeRequests[0].Reviewers, func(reviewer *gitlab.BasicUser, _ int) int64 {
			return reviewer.ID
		})
		newReviewerNames := lo.Without(updateGroup.Reviewers, existingReviewerNames...)
		newReviewerIds, err := p.getUserIds(client, newReviewerNames)
		if err != nil {
			return err
		}

		// Update the MR if something changed
		if mergeRequests[0].Title != updateGroup.Title ||
			mergeRequests[0].Description != content ||
			!lo.ElementsMatch(mergeRequests[0].Labels, updateGroup.Labels) ||
			len(newReviewerIds) > 0 {
			p.logger.Debug("Updating MR")
			// Prepare the options for updating the MR
			updateOptions := &gitlab.UpdateMergeRequestOptions{
				Title:       gitlab.Ptr(updateGroup.Title),
				Description: gitlab.Ptr(content),
				Labels:      p.convertLabels(updateGroup.Labels), // Always update the labels
			}
			// Fill the reviewers (if any)
			allReviewerIds := append(existingReviewerIds, newReviewerIds...)
			if len(allReviewerIds) > 0 {
				updateOptions.ReviewerIDs = gitlab.Ptr(allReviewerIds)
			}
			// Perform the update
			if _, _, err := client.MergeRequests.UpdateMergeRequest(project.Path, mergeRequests[0].IID, updateOptions); err != nil {
				return err
			}
		}
	} else {
		// Prepare the options for creating the MR
		createOptions := &gitlab.CreateMergeRequestOptions{
			Title:              gitlab.Ptr(updateGroup.Title),
			Description:        gitlab.Ptr(content),
			SourceBranch:       gitlab.Ptr(updateGroup.BranchName),
			TargetBranch:       gitlab.Ptr(p.settings.BaseBranch),
			RemoveSourceBranch: gitlab.Ptr(true),
		}
		// Fill the labels (if any)
		if len(updateGroup.Labels) > 0 {
			createOptions.Labels = p.convertLabels(updateGroup.Labels)
		}
		// Fill the reviewers (if any)
		reviewerIds, err := p.getUserIds(client, updateGroup.Reviewers)
		if err != nil {
			return err
		}
		if len(reviewerIds) > 0 {
			createOptions.ReviewerIDs = gitlab.Ptr(reviewerIds)
		}
		// Create the MR
		mr, _, err := client.MergeRequests.CreateMergeRequest(project.Path, createOptions)
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
	usedBranches := lo.FlatMap(cleanupSettings.UpdateGroups, func(x *common.UpdateGroup, _ int) []string {
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

func (p *GitlabPlatform) createClient() (*gitlab.Client, error) {
	if p.settings == nil || p.settings.Token == "" {
		return nil, fmt.Errorf("no platform token defined")
	}
	endpoint := p.getEndpoint()
	token := p.settings.TokenExpanded()
	return gitlab.NewClient(token, gitlab.WithBaseURL(endpoint))
}

func (p *GitlabPlatform) getEndpoint() string {
	endpoint := "https://gitlab.com/api/v4"
	if p.settings.Endpoint != "" {
		endpoint = p.settings.EndpointExpanded()
	}
	return endpoint
}

func (p *GitlabPlatform) convertLabels(labels []string) *gitlab.LabelOptions {
	if labels == nil {
		return nil
	} else if len(labels) == 0 {
		empty := []string{}
		return (*gitlab.LabelOptions)(&empty)
	}
	labelsClone := slices.Clone(labels)
	return (*gitlab.LabelOptions)(&labelsClone)
}

func (p *GitlabPlatform) getUserIds(client *gitlab.Client, usernames []string) ([]int64, error) {
	var userIds []int64
	if len(usernames) == 0 {
		return userIds, nil
	}
	for _, username := range usernames {
		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}
		// Try looking up the user id in the cache first
		cacheIdentifier := fmt.Sprintf("gitlab_userid/%s/%s", p.getEndpoint(), username)
		if p.settings.GitLabUserIdCache != nil {
			if cachedId, exists, err := p.settings.GitLabUserIdCache.Get(cacheIdentifier); err != nil {
				return nil, fmt.Errorf("failed to get user id from cache for '%s': %w", username, err)
			} else if exists {
				userIds = append(userIds, cachedId)
				continue
			}
		}
		// Lookup the user id via the API
		users, _, err := client.Users.ListUsers(&gitlab.ListUsersOptions{
			Username: gitlab.Ptr(username),
		})
		// TODO maybe: If user is not found, treat it as a group and return all groups member ids
		if err != nil {
			return nil, fmt.Errorf("failed to lookup user '%s': %w", username, err)
		}
		if len(users) == 0 {
			return nil, fmt.Errorf("user '%s' not found", username)
		}
		userIds = append(userIds, users[0].ID)
		// Cache the user id for future lookups
		if p.settings.GitLabUserIdCache != nil {
			p.settings.GitLabUserIdCache.Set(cacheIdentifier, users[0].ID, time.Hour)
		}
	}
	return userIds, nil
}

package platforms

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/google/go-github/v82/github"
	"github.com/roemer/gonovate/pkg/common"
	"github.com/samber/lo"
)

type GitHubPlatform struct {
	*GitPlatform
}

func NewGitHubPlatform(settings *common.PlatformSettings) *GitHubPlatform {
	platform := &GitHubPlatform{
		GitPlatform: NewGitPlatform(settings),
	}
	platform.impl = platform
	return platform
}

func (p *GitHubPlatform) Type() common.PlatformType {
	return common.PLATFORM_TYPE_GITHUB
}

func (p *GitHubPlatform) FetchProject(project *common.Project) error {
	// Prepare the data for the API
	owner, repository := project.SplitPath()
	// Create the client
	client, err := p.createClient()
	if err != nil {
		return err
	}
	// Get the repository
	platformRepository, _, err := client.Repositories.Get(context.Background(), owner, repository)
	if err != nil {
		return err
	}
	if platformRepository == nil {
		return fmt.Errorf("could not find project: %s", project.Path)
	}
	cloneUrl := *platformRepository.CloneURL
	cloneUrlWithCredentials, err := url.Parse(cloneUrl)
	if err != nil {
		return err
	}
	cloneUrlWithCredentials.User = url.UserPassword("oauth2", p.settings.TokenExpanded())
	_, _, err = common.Git.Run("clone", cloneUrlWithCredentials.String(), ClonePath)
	return err
}

func (p *GitHubPlatform) LookupAuthor() (string, string, error) {
	client, err := p.createClient()
	if err != nil {
		return "", "", err
	}
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", fmt.Errorf("failed to lookup user information")
	}
	if user.Email == nil {
		// Fetch the email separately
		emails, _, err := client.Users.ListEmails(context.Background(), &github.ListOptions{})
		if err != nil {
			return "", "", fmt.Errorf("failed to lookup user email: %w", err)
		}
		if len(emails) == 0 {
			return "", "", fmt.Errorf("email not found in user information. Has the token the correct scopes?")
		}
		// Use the first email
		// TODO maybe: Prioritize the email which is marked as "primary" and "verified"
		user.Email = emails[0].Email
	}
	return user.GetName(), user.GetEmail(), nil
}

func (p *GitHubPlatform) NotifyChanges(project *common.Project, updateGroup *common.UpdateGroup) error {
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
	// Trim spaces / newlines
	content = strings.TrimSpace(content)

	// Search for an existing PR
	existingRequest, _, err := client.PullRequests.List(context.Background(), owner, repository, &github.PullRequestListOptions{
		Head:  updateGroup.BranchName,
		Base:  p.settings.BaseBranch,
		State: "open",
	})
	if err != nil {
		return err
	}

	// The "Head" search parameter does not work without "user:", so just make sure that the returned list really contains the branch
	existingPr, prExists := lo.Find(existingRequest, func(pr *github.PullRequest) bool { return pr.Head.GetRef() == updateGroup.BranchName })
	if prExists {
		p.logger.Info(fmt.Sprintf("PR already exists: %s", existingPr.GetHTMLURL()))

		// Update the PR if something changed
		if existingPr.Title == nil || *existingPr.Title != updateGroup.Title ||
			existingPr.Body == nil || *existingPr.Body != content {
			p.logger.Debug("Updating PR")
			if _, _, err := client.PullRequests.Edit(context.Background(), owner, repository, existingPr.GetNumber(), &github.PullRequest{
				Title: github.Ptr(updateGroup.Title),
				Body:  github.Ptr(content),
			}); err != nil {
				return err
			}
		}
		existingLabels := lo.Map(existingPr.Labels, func(label *github.Label, _ int) string { return label.GetName() })
		if !lo.ElementsMatch(existingLabels, updateGroup.Labels) {
			p.logger.Debug("Updating PR labels")
			_, _, err := client.Issues.ReplaceLabelsForIssue(context.Background(), owner, repository, existingPr.GetNumber(), updateGroup.Labels)
			if err != nil {
				return err
			}
		}
		existingReviewers := lo.Map(existingPr.RequestedReviewers, func(reviewer *github.User, _ int) string { return reviewer.GetLogin() })
		newReviewers := lo.Without(updateGroup.Reviewers, existingReviewers...)
		if len(newReviewers) > 0 {
			p.logger.Debug("Updating PR reviewers")
			_, _, err := client.PullRequests.RequestReviewers(context.Background(), owner, repository, existingPr.GetNumber(), github.ReviewersRequest{
				Reviewers: newReviewers,
			})
			if err != nil {
				return err
			}
		}
	} else {
		// Create the PR
		pr, _, err := client.PullRequests.Create(context.Background(), owner, repository, &github.NewPullRequest{
			Title: github.Ptr(updateGroup.Title),
			Body:  github.Ptr(content),
			Head:  github.Ptr(updateGroup.BranchName),
			Base:  github.Ptr(p.settings.BaseBranch),
		})
		if err != nil {
			return err
		}
		p.logger.Info(fmt.Sprintf("Created PR: %s", pr.GetHTMLURL()))
		if len(updateGroup.Labels) > 0 {
			// Labels need to be added separately
			_, _, err := client.Issues.ReplaceLabelsForIssue(context.Background(), owner, repository, pr.GetNumber(), updateGroup.Labels)
			if err != nil {
				return err
			}
		}
		if len(updateGroup.Reviewers) > 0 {
			// Request reviewers
			_, _, err := client.PullRequests.RequestReviewers(context.Background(), owner, repository, pr.GetNumber(), github.ReviewersRequest{
				Reviewers: updateGroup.Reviewers,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *GitHubPlatform) Cleanup(cleanupSettings *PlatformCleanupSettings) error {
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
		existingRequest, _, err := client.PullRequests.List(context.Background(), owner, repository, &github.PullRequestListOptions{
			Head:  potentialStaleBranch,
			Base:  cleanupSettings.BaseBranch,
			State: "open",
		})
		if err != nil {
			return err
		}
		// The "Head" search parameter does not work without "user:", so just make sure that the returned list really contains the branch
		existingPr, prExists := lo.Find(existingRequest, func(pr *github.PullRequest) bool { return pr.Head.GetRef() == potentialStaleBranch })
		if prExists {
			// Close the PR
			p.logger.Info(fmt.Sprintf("Closing associated PR: %s", *existingPr.HTMLURL))
			if _, _, err := client.PullRequests.Edit(context.Background(), owner, repository, existingPr.GetNumber(), &github.PullRequest{
				State: github.Ptr("closed"),
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

func (p *GitHubPlatform) createClient() (*github.Client, error) {
	if p.settings == nil || p.settings.Token == "" {
		return nil, fmt.Errorf("no platform token defined")
	}
	return github.NewClient(nil).WithAuthToken(p.settings.TokenExpanded()), nil
}

package platforms

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"

	"github.com/google/go-github/v63/github"
	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
	"github.com/samber/lo"
)

type GitHubPlatform struct {
	GitPlatform
}

func NewGitHubPlatform(logger *slog.Logger, config *config.RootConfig) *GitHubPlatform {
	platform := &GitHubPlatform{
		GitPlatform: *NewGitPlatform(logger, config),
	}
	return platform
}

func (p *GitHubPlatform) Type() shared.PlatformType {
	return shared.PLATFORM_TYPE_GITHUB
}

func (p *GitHubPlatform) FetchProject(project *shared.Project) error {
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
	cloneUrlWithCredentials.User = url.UserPassword("oauth2", p.Config.PlatformSettings.TokendExpanded())
	_, _, err = shared.Git.Run("clone", cloneUrlWithCredentials.String(), ".gonovate-clone")
	return err
}

func (p *GitHubPlatform) NotifyChanges(project *shared.Project, updateGroup *shared.UpdateGroup) error {
	// Prepare the data for the API
	owner, repository := project.SplitPath()

	// Create the client
	client, err := p.createClient()
	if err != nil {
		return err
	}

	// Search for an existing PR
	existingRequest, _, err := client.PullRequests.List(context.Background(), owner, repository, &github.PullRequestListOptions{
		Head:  updateGroup.BranchName,
		Base:  p.Config.PlatformSettings.BaseBranch,
		State: "open",
	})
	if err != nil {
		return err
	}

	// Build the content of the PR
	content := ""
	for _, dep := range updateGroup.Dependencies {
		content += fmt.Sprintf("- %s from %s to %s\n", dep.Name, dep.Version, dep.NewRelease.VersionString)
	}

	// The "Head" search parameter does not work without "user:", so just make sure that the returned list really contains the branch
	existingPr, prExists := lo.Find(existingRequest, func(pr *github.PullRequest) bool { return pr.Head.GetRef() == updateGroup.BranchName })
	if prExists {
		p.logger.Info(fmt.Sprintf("PR already exists: %s", *existingPr.HTMLURL))

		// Update the PR if something changed
		if existingPr.Title == nil || *existingPr.Title != updateGroup.Title || existingPr.Body == nil || *existingPr.Body != content {
			p.logger.Debug("Updating PR")
			if _, _, err := client.PullRequests.Edit(context.Background(), owner, repository, existingPr.GetNumber(), &github.PullRequest{
				Title: github.String(updateGroup.Title),
				Body:  github.String(content),
			}); err != nil {
				return err
			}
		}
	} else {
		// Create the PR
		pr, _, err := client.PullRequests.Create(context.Background(), owner, repository, &github.NewPullRequest{
			Title: github.String(updateGroup.Title),
			Body:  github.String(content),
			Head:  github.String(updateGroup.BranchName),
			Base:  github.String(p.Config.PlatformSettings.BaseBranch),
		})
		if err != nil {
			return err
		}
		p.logger.Info(fmt.Sprintf("Created PR: %s", *pr.HTMLURL))
	}
	return nil
}

func (p *GitHubPlatform) Cleanup(cleanupSettings *PlatformCleanupSettings) error {
	// Get the remote branches
	stdout, _, err := shared.Git.Run("branch", "-r")
	if err != nil {
		return err
	}
	allBranches := strings.Split(stdout, "\n")
	// Remove the origin prefix
	allBranches = lo.Map(allBranches, func(x string, _ int) string { return strings.TrimPrefix(x, "origin/") })

	// Filter to those that are relevant for gonovate
	gonovateBranches := lo.Filter(allBranches, func(x string, _ int) bool {
		return strings.HasPrefix(x, cleanupSettings.BranchPrefix)
	})

	// Get the branches that were used in this gonovate run
	usedBranches := lo.FlatMap(cleanupSettings.UpdateGroups, func(x *shared.UpdateGroup, _ int) []string {
		return []string{x.BranchName}
	})

	for _, potentialStaleBranch := range gonovateBranches {
		if slices.Contains(usedBranches, potentialStaleBranch) {
			// This branch is used
			continue
		}
		// This branch is unused, delete the branch and a possible associated PR
		p.logger.Info(fmt.Sprintf("Unused branch '%s'", potentialStaleBranch))

		// TODO: Search for PR
		// TODO: If a PR is found, delete/close it
		// TODO: Delete the unused branch
	}

	return nil
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

func (p *GitHubPlatform) createClient() (*github.Client, error) {
	if p.Config.PlatformSettings == nil || p.Config.PlatformSettings.Token == "" {
		return nil, fmt.Errorf("no platform token defined")
	}
	return github.NewClient(nil).WithAuthToken(p.Config.PlatformSettings.TokendExpanded()), nil
}

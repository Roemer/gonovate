package platforms

import (
	"context"
	"fmt"
	"gonovate/core"
	"log/slog"
	"net/url"

	"github.com/google/go-github/v62/github"
	"github.com/samber/lo"
)

type GithubPlatform struct {
	GitPlatform
}

func NewGithubPlatform(logger *slog.Logger, config *core.Config) *GithubPlatform {
	platform := &GithubPlatform{
		GitPlatform: *NewGitPlatform(logger, config),
	}
	return platform
}

func (p *GithubPlatform) Type() string {
	return core.PLATFORM_TYPE_GITHUB
}

func (p *GithubPlatform) SearchProjects() ([]*core.Project, error) {
	// TODO
	return nil, nil
}

func (p *GithubPlatform) FetchProject(project *core.Project) error {
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
	_, _, err = core.Git.Run("clone", cloneUrlWithCredentials.String(), ".gonovate-clone")
	return err
}

func (p *GithubPlatform) NotifyChanges(project *core.Project, changeSet *core.ChangeSet) error {
	// Prepare the data for the API
	owner, repository := project.SplitPath()

	// Create the client
	client, err := p.createClient()
	if err != nil {
		return err
	}
	existingRequest, _, err := client.PullRequests.List(context.Background(), owner, repository, &github.PullRequestListOptions{
		Head:  changeSet.Id,
		Base:  p.BaseBranch,
		State: "open",
	})
	if err != nil {
		return err
	}
	// The Head search parameter does not work without "user:", so just make sure that the returned list really contains the branch
	existingPr, prExists := lo.Find(existingRequest, func(pr *github.PullRequest) bool { return pr.Head.GetRef() == changeSet.Id })
	if prExists {
		p.logger.Info(fmt.Sprintf("PR already exists: %s", *existingPr.HTMLURL))
	} else {
		// Create the PR
		pr, _, err := client.PullRequests.Create(context.Background(), owner, repository, &github.NewPullRequest{
			Title: github.String(changeSet.Title),
			Head:  github.String(changeSet.Id),
			Base:  github.String(p.BaseBranch),
		})
		if err != nil {
			return err
		}
		p.logger.Info(fmt.Sprintf("Created PR: %s", *pr.HTMLURL))
	}
	return nil
}

func (p *GithubPlatform) createClient() (*github.Client, error) {
	if p.Config.PlatformSettings == nil || p.Config.PlatformSettings.Token == "" {
		return nil, fmt.Errorf("no platform token defined")
	}
	return github.NewClient(nil).WithAuthToken(p.Config.PlatformSettings.TokendExpanded()), nil
}

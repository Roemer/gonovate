package platforms

import (
	"context"
	"fmt"
	"gonovate/core"
	"log/slog"
	"os"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/samber/lo"
)

type GithubPlatform struct {
	gitPlatform
}

func NewGithubPlatform(logger *slog.Logger, config *core.Config) IPlatform {
	platform := &GithubPlatform{
		gitPlatform: gitPlatform{
			platformBase: platformBase{
				logger: logger,
				Config: config,
			},
			BaseBranch: "main",
		},
	}
	return platform
}

func (p *GithubPlatform) SearchProjects() ([]*core.Project, error) {
	// TODO
	return nil, nil
}

func (p *GithubPlatform) FetchProject(project *core.Project) error {
	// TODO
	return nil
}

func (p *GithubPlatform) PrepareForChanges(change *core.Change) error {
	return p.CreateBranch(change)
}

func (p *GithubPlatform) SubmitChanges(change *core.Change) error {
	if err := p.AddAll(); err != nil {
		return err
	}
	return p.Commit(change)
}

func (p *GithubPlatform) PublishChanges(change *core.Change) error {
	return p.PushBranch()
}

func (p *GithubPlatform) NotifyChanges(change *core.Change) error {
	// Prepare the data for the API
	token := ""
	owner := ""
	repository := ""
	// Try get the values from the config file(s)
	if p.Config.PlatformSettings != nil {
		token = p.Config.PlatformSettings.TokendExpanded()
		owner = p.Config.PlatformSettings.Owner
		repository = p.Config.PlatformSettings.Repository
	}
	// Overwrite with environment variables
	if value, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		token = value
	}
	if value, ok := os.LookupEnv("GITHUB_REPOSITORY_OWNER"); ok {
		owner = value
	}
	if value, ok := os.LookupEnv("GITHUB_REPOSITORY"); ok {
		parts := strings.SplitN(value, "/", 2)
		repository = parts[1]
	}

	// Create the client
	client := github.NewClient(nil).WithAuthToken(token)
	existingRequest, _, err := client.PullRequests.List(context.Background(), owner, repository, &github.PullRequestListOptions{
		Head:  change.Data["branchName"],
		Base:  p.BaseBranch,
		State: "open",
	})
	if err != nil {
		return err
	}
	// The Head search parameter does not work without "user:", so just make sure that the returned list really contains the branch
	prExists := lo.ContainsBy(existingRequest, func(pr *github.PullRequest) bool { return pr.Head.GetRef() == change.Data["branchName"] })
	if prExists {
		p.logger.Info(fmt.Sprintf("PR already exists: %s", *existingRequest[0].HTMLURL))
	} else {
		// Create the PR
		pr, _, err := client.PullRequests.Create(context.Background(), owner, repository, &github.NewPullRequest{
			Title: github.String(change.Data["msg"]),
			Head:  github.String(change.Data["branchName"]),
			Base:  github.String(p.BaseBranch),
		})
		if err != nil {
			return err
		}
		p.logger.Info(fmt.Sprintf("Created PR: %s", *pr.HTMLURL))
	}
	return nil
}

func (p *GithubPlatform) ResetToBase() error {
	return p.CheckoutBaseBranch()
}

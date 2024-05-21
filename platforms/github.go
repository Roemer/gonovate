package platforms

import (
	"context"
	"fmt"
	"gonovate/core"
	"log/slog"
	"os"
	"strings"

	"github.com/google/go-github/v62/github"
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
	// Create the PR
	pr, _, err := client.PullRequests.Create(context.Background(), owner, repository, &github.NewPullRequest{
		Title: github.String(change.Data["msg"]),
		Head:  github.String(change.Data["branchName"]),
		Base:  github.String("main"),
	})
	if err != nil {
		return err
	}
	p.logger.Info(fmt.Sprintf("Created PR: %s", *pr.URL))
	return nil
}

func (p *GithubPlatform) ResetToBase() error {
	return p.CheckoutBaseBranch()
}

package platforms

import (
	"gonovate/core"
	"log/slog"
)

type GitlabPlatform struct {
	gitPlatform
}

func NewGitlabPlatform(logger *slog.Logger, config *core.Config) IPlatform {
	platform := &GitlabPlatform{
		gitPlatform: gitPlatform{
			platformBase: platformBase{
				logger: logger,
			},
		},
	}
	return platform
}

func (p *GitlabPlatform) SearchProjects() ([]*core.Project, error) {
	// TODO
	return nil, nil
}

func (p *GitlabPlatform) FetchProject(project *core.Project) error {
	// TODO
	return nil
}

func (p *GitlabPlatform) PrepareForChanges(packageName, oldVersion, newVersion string) error {
	return p.CreateBranch(packageName, oldVersion, newVersion)
}

func (p *GitlabPlatform) SubmitChanges(packageName, oldVersion, newVersion string) error {
	if err := p.AddAll(); err != nil {
		return err
	}
	return p.Commit(packageName, oldVersion, newVersion)
}

func (p *GitlabPlatform) PublishChanges() error {
	return p.PushBranch()
}

func (p *GitlabPlatform) ResetToBase() error {
	return p.CheckoutBaseBranch()
}

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
				Config: config,
			},
			BaseBranch: "main",
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

func (p *GitlabPlatform) PrepareForChanges(change core.IChange) error {
	return p.CreateBranch(change.GetMeta())
}

func (p *GitlabPlatform) SubmitChanges(change core.IChange) error {
	if err := p.AddAll(); err != nil {
		return err
	}
	return p.Commit(change.GetMeta())
}

func (p *GitlabPlatform) PublishChanges(change core.IChange) error {
	return p.PushBranch()
}

func (p *GitlabPlatform) NotifyChanges(change core.IChange) error {
	// TODO: Create MR
	return nil
}

func (p *GitlabPlatform) ResetToBase() error {
	return p.CheckoutBaseBranch()
}

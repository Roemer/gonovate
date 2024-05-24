package platforms

import (
	"gonovate/core"
	"log/slog"
)

type GitlabPlatform struct {
	GitPlatform
}

func NewGitlabPlatform(logger *slog.Logger, config *core.Config) *GitlabPlatform {
	platform := &GitlabPlatform{
		GitPlatform: *NewGitPlatform(logger, config),
	}
	return platform
}

func (p *GitlabPlatform) Type() string {
	return core.PLATFORM_TYPE_GITLAB
}

func (p *GitlabPlatform) SearchProjects() ([]*core.Project, error) {
	// TODO
	return nil, nil
}

func (p *GitlabPlatform) FetchProject(project *core.Project) error {
	// TODO
	return nil
}

func (p *GitlabPlatform) NotifyChanges(change core.IChange) error {
	// TODO: Create MR
	return nil
}

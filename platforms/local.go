package platforms

import (
	"gonovate/core"
	"log/slog"
)

type LocalPlatform struct {
	platformBase
}

func NewLocalPlatform(logger *slog.Logger, config *core.Config) IPlatform {
	platform := &LocalPlatform{
		platformBase: platformBase{
			logger: logger,
			Config: config,
		},
	}
	return platform
}

func (p *LocalPlatform) Type() string {
	return core.PLATFORM_TYPE_LOCAL
}

func (p *LocalPlatform) SearchProjects() ([]*core.Project, error) {
	return nil, nil
}

func (p *LocalPlatform) FetchProject(project *core.Project) error {
	return nil
}

func (p *LocalPlatform) PrepareForChanges(change core.IChange) error {
	return nil
}

func (p *LocalPlatform) SubmitChanges(change core.IChange) error {
	return nil
}

func (p *LocalPlatform) PublishChanges(change core.IChange) error {
	return nil
}

func (p *LocalPlatform) NotifyChanges(change core.IChange) error {
	return nil
}

func (p *LocalPlatform) ResetToBase() error {
	return nil
}

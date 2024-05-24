package platforms

import (
	"gonovate/core"
	"log/slog"
)

type NoopPlatform struct {
	platformBase
}

func NewNoopPlatform(logger *slog.Logger, config *core.Config) IPlatform {
	platform := &NoopPlatform{
		platformBase: platformBase{
			logger: logger,
			Config: config,
		},
	}
	return platform
}

func (p *NoopPlatform) Type() string {
	return core.PLATFORM_TYPE_NOOP
}

func (p *NoopPlatform) SearchProjects() ([]*core.Project, error) {
	return nil, nil
}

func (p *NoopPlatform) FetchProject(project *core.Project) error {
	return nil
}

func (p *NoopPlatform) PrepareForChanges(change core.IChange) error {
	return nil
}

func (p *NoopPlatform) SubmitChanges(change core.IChange) error {
	return nil
}

func (p *NoopPlatform) PublishChanges(change core.IChange) error {
	return nil
}

func (p *NoopPlatform) NotifyChanges(change core.IChange) error {
	return nil
}

func (p *NoopPlatform) ResetToBase() error {
	return nil
}

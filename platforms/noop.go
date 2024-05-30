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

func (p *NoopPlatform) PrepareForChanges(changeSet *core.ChangeSet) error {
	return nil
}

func (p *NoopPlatform) SubmitChanges(changeSet *core.ChangeSet) error {
	return nil
}

func (p *NoopPlatform) PublishChanges(changeSet *core.ChangeSet) error {
	return nil
}

func (p *NoopPlatform) NotifyChanges(project *core.Project, changeSet *core.ChangeSet) error {
	return nil
}

func (p *NoopPlatform) ResetToBase() error {
	return nil
}

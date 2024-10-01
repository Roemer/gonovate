package platforms

import (
	"log/slog"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type NoopPlatform struct {
	platformBase
}

func NewNoopPlatform(logger *slog.Logger, config *config.RootConfig) IPlatform {
	platform := &NoopPlatform{
		platformBase: platformBase{
			logger: logger,
			Config: config,
		},
	}
	return platform
}

func (p *NoopPlatform) Type() shared.PlatformType {
	return shared.PLATFORM_TYPE_NOOP
}

func (p *NoopPlatform) FetchProject(project *shared.Project) error {
	return nil
}

func (p *NoopPlatform) PrepareForChanges(updateGroup *shared.UpdateGroup) error {
	return nil
}

func (p *NoopPlatform) SubmitChanges(updateGroup *shared.UpdateGroup) error {
	return nil
}

func (p *NoopPlatform) PublishChanges(updateGroup *shared.UpdateGroup) error {
	return nil
}

func (p *NoopPlatform) NotifyChanges(project *shared.Project, updateGroup *shared.UpdateGroup) error {
	return nil
}

func (p *NoopPlatform) ResetToBase() error {
	return nil
}

func (p *NoopPlatform) Cleanup(cleanupSettings *PlatformCleanupSettings) error {
	return nil
}

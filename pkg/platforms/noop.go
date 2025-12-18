package platforms

import (
	"github.com/roemer/gonovate/pkg/common"
)

type NoopPlatform struct {
	*platformBase
}

func NewNoopPlatform(settings *common.PlatformSettings) IPlatform {
	platform := &NoopPlatform{
		platformBase: newPlatformBase(settings),
	}
	return platform
}

func (p *NoopPlatform) Type() common.PlatformType {
	return common.PLATFORM_TYPE_NOOP
}

func (p *NoopPlatform) FetchProject(project *common.Project) error {
	return nil
}

func (p *NoopPlatform) PrepareForChanges(updateGroup *common.UpdateGroup) error {
	return nil
}

func (p *NoopPlatform) SubmitChanges(updateGroup *common.UpdateGroup) error {
	return nil
}

func (p *NoopPlatform) IsNewOrChanged(updateGroup *common.UpdateGroup) (bool, error) {
	return true, nil
}

func (p *NoopPlatform) PublishChanges(updateGroup *common.UpdateGroup) error {
	return nil
}

func (p *NoopPlatform) NotifyChanges(project *common.Project, updateGroup *common.UpdateGroup) error {
	return nil
}

func (p *NoopPlatform) ResetToBase(baseBranch string) error {
	return nil
}

func (p *NoopPlatform) Cleanup(cleanupSettings *PlatformCleanupSettings) error {
	return nil
}

package platforms

import (
	"fmt"
	"log/slog"

	"github.com/roemer/gonovate/pkg/common"
)

type IPlatform interface {
	// Returns the type of the platform
	Type() common.PlatformType
	// Fetches the project from the platform in it's initial state.
	FetchProject(project *common.Project) error
	// Prepares the project to accept changes.
	PrepareForChanges(updateGroup *common.UpdateGroup) error
	// Looks up the author to use for commits.
	LookupAuthor() (string, string, error)
	// Submit the changes to the project locally.
	SubmitChanges(updateGroup *common.UpdateGroup) error
	// Checks if the remote already has the same changes.
	IsNewOrChanged(updateGroup *common.UpdateGroup) (bool, error)
	// Publishes the changes to the remote location.
	PublishChanges(updateGroup *common.UpdateGroup) error
	// Notifies the remote about the changes with eg. MRs/PRs.
	NotifyChanges(project *common.Project, updateGroup *common.UpdateGroup) error
	// Resets the project to the initial state for other changes.
	ResetToBase(baseBranch string) error
	// Cleans the platform after a gonovate run.
	Cleanup(cleanupSettings *PlatformCleanupSettings) error
}

var ClonePath = ".gonovate-clone"

type PlatformCleanupSettings struct {
	Project      *common.Project
	UpdateGroups []*common.UpdateGroup
	BaseBranch   string
	BranchPrefix string
}

type platformBase struct {
	logger   *slog.Logger
	settings *common.PlatformSettings
}

func newPlatformBase(settings *common.PlatformSettings) *platformBase {
	return &platformBase{
		logger:   settings.Logger.With(slog.String("platform", string(settings.Platform))),
		settings: settings,
	}
}

func GetPlatform(settings *common.PlatformSettings) (IPlatform, error) {
	switch settings.Platform {
	case common.PLATFORM_TYPE_GIT:
		return NewGitPlatform(settings), nil
	case common.PLATFORM_TYPE_GITEA:
		return NewGiteaPlatform(settings), nil
	case common.PLATFORM_TYPE_GITHUB:
		return NewGitHubPlatform(settings), nil
	case common.PLATFORM_TYPE_GITLAB:
		return NewGitlabPlatform(settings), nil
	case common.PLATFORM_TYPE_NOOP:
		return NewNoopPlatform(settings), nil
	}
	return nil, fmt.Errorf("no platform defined for '%s'", settings.Platform)
}

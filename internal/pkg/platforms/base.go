package platforms

import (
	"fmt"
	"log/slog"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type IPlatform interface {
	// Returns the type of the platform
	Type() shared.PlatformType
	// Fetches the project from the platform in it's initial state.
	FetchProject(project *shared.Project) error
	// Prepares the project to accept changes.
	PrepareForChanges(updateGroup *shared.UpdateGroup) error
	// Submit the changes to the project locally.
	SubmitChanges(updateGroup *shared.UpdateGroup) error
	// Publishes the changes to the remote location.
	PublishChanges(updateGroup *shared.UpdateGroup) error
	// Notifies the remote about the changes with eg. MRs/PRs.
	NotifyChanges(project *shared.Project, updateGroup *shared.UpdateGroup) error
	// Resets the project to the initial state for other changes.
	ResetToBase() error
	// Cleans the platform after a gonovate run.
	Cleanup(cleanupSettings *PlatformCleanupSettings) error
}

type PlatformCleanupSettings struct {
	Project      *shared.Project
	UpdateGroups []*shared.UpdateGroup
	BaseBranch   string
	BranchPrefix string
}

type platformBase struct {
	logger *slog.Logger
	Config *config.RootConfig
}

func GetPlatform(logger *slog.Logger, config *config.RootConfig) (IPlatform, error) {
	switch config.Platform {
	case shared.PLATFORM_TYPE_GIT:
		return NewGitPlatform(logger, config), nil
	case shared.PLATFORM_TYPE_GITHUB:
		return NewGitHubPlatform(logger, config), nil
	case shared.PLATFORM_TYPE_GITLAB:
		return NewGitlabPlatform(logger, config), nil
	case shared.PLATFORM_TYPE_NOOP:
		return NewNoopPlatform(logger, config), nil
	}
	return nil, fmt.Errorf("no platform defined for '%s'", config.Platform)
}

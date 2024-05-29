package platforms

import (
	"fmt"
	"gonovate/core"
	"log/slog"
)

type IPlatform interface {
	// Returns the type of the platform
	Type() string
	// Ge the base branch
	BaseBranch() string
	// Searches on the platform for projects to run gonovate on.
	SearchProjects() ([]*core.Project, error)
	// Fetches the project from the platform in it's initial state.
	FetchProject(project *core.Project) error
	// Prepares the project to accept changes.
	PrepareForChanges(changeSet *core.ChangeSet) error
	// Submit the changes to the project locally.
	SubmitChanges(changeSet *core.ChangeSet) error
	// Publishes the changes to the remote location.
	PublishChanges(changeSet *core.ChangeSet) error
	// Notifies the remote about the changes with eg. MRs/PRs.
	NotifyChanges(project *core.Project, changeSet *core.ChangeSet) error
	// Resets the project to the initial state for other changes.
	ResetToBase() error
}

type platformBase struct {
	logger     *slog.Logger
	Config     *core.Config
	baseBranch string
}

func (p *platformBase) BaseBranch() string {
	return p.baseBranch
}

func GetPlatform(logger *slog.Logger, config *core.Config) (IPlatform, error) {
	switch config.Platform {
	case core.PLATFORM_TYPE_GIT:
		return NewGitPlatform(logger, config), nil
	case core.PLATFORM_TYPE_GITHUB:
		return NewGitHubPlatform(logger, config), nil
	case core.PLATFORM_TYPE_GITLAB:
		return NewGitlabPlatform(logger, config), nil
	case core.PLATFORM_TYPE_NOOP:
		return NewNoopPlatform(logger, config), nil
	}
	return nil, fmt.Errorf("no platform defined for '%s'", config.Platform)
}

package platforms

import (
	"fmt"
	"gonovate/core"
	"log/slog"
)

type IPlatform interface {
	// Searches on the platform for projects to run gonovate on.
	SearchProjects() ([]*core.Project, error)
	// Fetches the project from the platform in it's initial state.
	FetchProject(project *core.Project) error
	// Prepares the project to accept changes.
	PrepareForChanges(change *core.Change) error
	// Submit the changes to the project locally.
	SubmitChanges(change *core.Change) error
	// Publishes the changes to the remote location.
	PublishChanges(change *core.Change) error
	// Notifies the remote about the changes with eg. MRs/PRs.
	NotifyChanges(change *core.Change) error
	// Resets the project to the initial state for other changes.
	ResetToBase() error
}

type platformBase struct {
	logger *slog.Logger
	Config *core.Config
}

func GetPlatform(logger *slog.Logger, config *core.Config) (IPlatform, error) {
	switch config.Platform {
	case core.PLATFORM_TYPE_GITLAB:
		return NewGitlabPlatform(logger, config), nil
	}
	return nil, fmt.Errorf("no platform defined for '%s'", config.Platform)
}

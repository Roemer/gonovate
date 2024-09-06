package common

import (
	"log/slog"
	"os"
)

type PlatformSettings struct {
	// The logger to use for the platform.
	Logger *slog.Logger
	// The type of the platform.
	Platform PlatformType
	// The token which is used to interact with the platform. Is expanded from environment variables.
	Token string
	// The endpoint to use when interacting with the platform.
	Endpoint string
	// The author to use when interacting with git.
	GitAuthor string
	// The name of the base branch.
	BaseBranch string
}

func (ps *PlatformSettings) TokendExpanded() string {
	return os.ExpandEnv(ps.Token)
}

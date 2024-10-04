package common

import (
	"log/slog"

	"github.com/roemer/gonovate/pkg/presets"
)

// This struct contains settings relevant for managers.
type ManagerSettings struct {
	// The logger to use for the manager.
	Logger *slog.Logger
	// The id of the manager.
	Id string
	// The type of the manager.
	ManagerType ManagerType
	// A flag which is set when the manager is disabled.
	Disabled *bool
	// A list of patterns with the files that the manager should process.
	FilePatterns []string
	// Settings for the RegexManager.
	RegexManagerSettings *RegexManagerSettings
	// Settings for the DevcontainerManager.
	DevcontainerManagerSettings *DevcontainerManagerSettings
}

// Settings relevant for the regex manager.
type RegexManagerSettings struct {
	MatchStringPresets map[string]*presets.MatchStringPreset
	MatchStrings       []string
}

// Settings relevant for the devcontainer manager.
type DevcontainerManagerSettings struct {
	FeatureDependencies map[string][]*DevcontainerManagerFeatureDependency
}

type DevcontainerManagerFeatureDependency struct {
	Property       string
	Datasource     DatasourceType
	DependencyName string
}

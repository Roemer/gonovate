package config

import (
	"os"

	"github.com/roemer/gonovate/internal/pkg/shared"
)

// This type represents the root config object.
type RootConfig struct {
	// The type of the platform to use.
	Platform shared.PlatformType `json:"platform"`
	// Settings that are relevant for the specified platform.
	PlatformSettings *PlatformSettings `json:"platformSettings"`
	// A map of presets for matchstrings that can be used and referenced.
	MatchStringPresets map[string]*MatchStringPreset `json:"matchStringPresets"`
	// A map of presets for versionings that can be used and referenced.
	VersioningPresets map[string]string `json:"versioningPresets"`
	// A list of presets to also load before loading this config. All configs are merged together.
	Extends []string `json:"extends"`
	// A list of patterns that will be completely ignored.
	IgnorePatterns []string `json:"ignorePatterns"`
	// A list of configurations for managers
	Managers []*ManagerConfig `json:"managers"`
	// A list of rules that can apply to managers or dependencies.
	Rules []*Rule `json:"rules"`
	// A list of rules that can apply to hosts.
	HostRules []*HostRule `json:"hostRules"`
}

type MatchStringPreset struct {
	MatchString       string   `json:"matchString"`
	ParameterDefaults []string `json:"parameterDefaults"`
}

// This type defines settings regarding the platform.
type PlatformSettings struct {
	Token     string   `json:"token"`
	GitAuthor string   `json:"gitAuthor"`
	Endpoint  string   `json:"endpoint"`
	Inplace   *bool    `json:"inplace"`
	Projects  []string `json:"projects"`
	// The name of the base branch, defaults to "main".
	BaseBranch string `json:"baseBranch"`
	// The prefix for branches created by gonovate. Defaults to "gonovate/".
	BranchPrefix string `json:"branchPrefix"`
}

func (ps *PlatformSettings) TokendExpanded() string {
	return os.ExpandEnv(ps.Token)
}

// This type represents the config for a manager with its settings and settings that apply for all dependencies.
type ManagerConfig struct {
	Id   string             `json:"id"`
	Type shared.ManagerType `json:"type"`
	// These settings are immediately converted to rules to keep the right order, so they should not be used directly
	ManagerSettings    *ManagerSettings    `json:"managerSettings"`
	DependencySettings *DependencySettings `json:"dependencySettings"`
}

type ManagerSettings struct {
	// General settings
	Disabled     *bool    `json:"disabled"`
	FilePatterns []string `json:"filePatterns"`
	// Specific settings for RegexManager
	MatchStrings []string `json:"matchStrings"`
	// Specific settings for DevcontainerManager
	DevcontainerSettings map[string][]*DevcontainerFeatureDependency `json:"devcontainerSettings"`
}

type DependencySettings struct {
	// A flag that allows disabling individual dependencies.
	Skip *bool `json:"skip"`
	// An optional text to describe, why a dependency was disabled.
	SkipReason string `json:"skipReason"`
	// Defines how much the dependency is allowed to update. Can be "major", "minor", or "patch".
	MaxUpdateType shared.UpdateType `json:"maxUpdateType"`
	// This flag defines if unstable releases are allowed. Unstable usually means a version that also has parts with text.
	AllowUnstable *bool `json:"allowUnstable"`
	// A list of registry urls to use. Allows overwriting the default. Depends on the datasource.
	RegistryUrls []string `json:"registryUrls"`
	// Defines the regexp to use to parse the version into separate parts. See gover for more details.
	Versioning string `json:"versioning"`
	// An optional regexp that is used to separate the version part from the rest from the raw string from external sources.
	ExtractVersion string `json:"extractVersion"`
	// A flag to indicate if versions from a remote that do not match the versioning should be ignored or give an exception.
	IgnoreNonMatching *bool `json:"ignoreNonMatching"`
	// Allows hard-coding a dependencyName in rules. Is used if it is not captured via matchString.
	DependencyName string `json:"dependencyName"`
	// Allows hard-coding a datasource in rules. Is used if it is not captured via matchString.
	Datasource shared.DatasourceType `json:"datasource"`
	// Allows defining regexes that replace further information from dependencies (like hash) after updating
	PostUpgradeReplacements []string `json:"postUpgradeReplacements"`
	// An optional name of a group to group dependency updates together.
	GroupName string `json:"groupName"`
}

type Rule struct {
	Matches            *RuleMatch          `json:"matches"`
	ManagerSettings    *ManagerSettings    `json:"managerSettings"`
	DependencySettings *DependencySettings `json:"dependencySettings"`
}

type RuleMatch struct {
	Managers        []string                `json:"managers"`
	ManagerTypes    []shared.ManagerType    `json:"managerTypes"`
	Files           []string                `json:"files"`
	DependencyNames []string                `json:"dependencyNames"`
	Datasources     []shared.DatasourceType `json:"datasources"`
}

// A MatchAll rule is a rule that has no matches defined at all, so it will match everything.
func (rm *RuleMatch) IsMatchAll() bool {
	return rm == nil || (len(rm.Managers) == 0 &&
		len(rm.Files) == 0 &&
		len(rm.DependencyNames) == 0 &&
		len(rm.Datasources) == 0)
}

type HostRule struct {
	MatchHost string `json:"matchHost"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Token     string `json:"token"`
}

func (hr *HostRule) UsernameExpanded() string {
	return os.ExpandEnv(hr.Username)
}

func (hr *HostRule) PasswordExpanded() string {
	return os.ExpandEnv(hr.Password)
}

func (hr *HostRule) TokendExpanded() string {
	return os.ExpandEnv(hr.Token)
}

type DevcontainerFeatureDependency struct {
	Property       string                `json:"property"`
	Datasource     shared.DatasourceType `json:"datasource"`
	DependencyName string                `json:"dependencyName"`
}

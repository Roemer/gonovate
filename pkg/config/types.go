package config

import (
	"github.com/roemer/gonovate/pkg/common"
)

// This type represents the gonovate config object.
type GonovateConfig struct {
	// Settings that are relevant for the platform.
	Platform *PlatformConfig `json:"platform" yaml:"platform"`
	// A map of presets for matchstrings that can be used and referenced.
	MatchStringPresets map[string]*MatchStringPreset `json:"matchStringPresets" yaml:"matchStringPresets"`
	// A map of presets for versionings that can be used and referenced.
	VersioningPresets map[string]string `json:"versioningPresets" yaml:"versioningPresets"`
	// A list of presets to also load before loading this config. All configs are merged together.
	Extends []string `json:"extends" yaml:"extends"`
	// A list of patterns that will be completely ignored.
	IgnorePatterns []string `json:"ignorePatterns" yaml:"ignorePatterns"`
	// A list of configurations for managers
	Managers []*Manager `json:"managers" yaml:"managers"`
	// A list of rules that can apply to managers or dependencies.
	Rules []*Rule `json:"rules" yaml:"rules"`
	// A list of rules that can apply to hosts.
	HostRules []*common.HostRule `json:"hostRules" yaml:"hostRules"`
}

type MatchStringPreset struct {
	MatchString       string   `json:"matchString" yaml:"matchString"`
	ParameterDefaults []string `json:"parameterDefaults" yaml:"parameterDefaults"`
}

// This type defines configurations regarding the platform.
type PlatformConfig struct {
	// The type of the platform to use.
	Type      common.PlatformType `json:"type" yaml:"type"`
	Token     string              `json:"token" yaml:"token"`
	GitAuthor string              `json:"gitAuthor" yaml:"gitAuthor"`
	Endpoint  string              `json:"endpoint" yaml:"endpoint"`
	Inplace   *bool               `json:"inplace" yaml:"inplace"`
	Projects  []string            `json:"projects" yaml:"projects"`
	// The name of the base branch, defaults to "main".
	BaseBranch string `json:"baseBranch" yaml:"baseBranch"`
	// The prefix for branches created by gonovate. Defaults to "gonovate/".
	BranchPrefix string `json:"branchPrefix" yaml:"branchPrefix"`
}

// This type represents an instance of manager with its configs and configs that apply for all dependencies within this manager.
type Manager struct {
	Id   string             `json:"id" yaml:"id"`
	Type common.ManagerType `json:"type" yaml:"type"`
	// These settings are immediately converted to rules to keep the right order, so they should not be used directly
	ManagerConfig    *ManagerConfig    `json:"managerConfig" yaml:"managerConfig"`
	DependencyConfig *DependencyConfig `json:"dependencyConfig" yaml:"dependencyConfig"`
}

type ManagerConfig struct {
	// General settings
	Disabled          *bool    `json:"disabled" yaml:"disabled"`
	FilePatterns      []string `json:"filePatterns" yaml:"filePatterns"`
	ClearFilePatterns *bool    `json:"clearFilePatterns" yaml:"clearFilePatterns"`
	// Specific settings for RegexManager
	MatchStrings []string `json:"matchStrings" yaml:"matchStrings"`
	// Specific settings for DevcontainerManager
	DevcontainerConfig map[string][]*DevcontainerFeatureDependency `json:"devcontainerConfig" yaml:"devcontainerConfig"`
}

type DevcontainerFeatureDependency struct {
	Property       string                `json:"property" yaml:"property"`
	Datasource     common.DatasourceType `json:"datasource" yaml:"datasource"`
	DependencyName string                `json:"dependencyName" yaml:"dependencyName"`
}

type DependencyConfig struct {
	// A flag that allows disabling individual dependencies.
	Skip *bool `json:"skip" yaml:"skip"`
	// An optional text to describe, why a dependency was disabled.
	SkipReason string `json:"skipReason" yaml:"skipReason"`
	// Defines how much the dependency is allowed to update. Can be "major", "minor", or "patch".
	MaxUpdateType common.UpdateType `json:"maxUpdateType" yaml:"maxUpdateType"`
	// This flag defines if unstable releases are allowed. Unstable usually means a version that also has parts with text.
	AllowUnstable *bool `json:"allowUnstable" yaml:"allowUnstable"`
	// A list of registry urls to use. Allows overwriting the default. Depends on the datasource.
	RegistryUrls []string `json:"registryUrls" yaml:"registryUrls"`
	// Defines the regexp to use to parse the version into separate parts. See gover for more details.
	Versioning string `json:"versioning" yaml:"versioning"`
	// An optional regexp that is used to separate the version part from the rest from the raw string from external sources.
	ExtractVersion string `json:"extractVersion" yaml:"extractVersion"`
	// A flag to indicate if versions from a remote that do not match the versioning should be ignored or give an exception.
	IgnoreNonMatching *bool `json:"ignoreNonMatching" yaml:"ignoreNonMatching"`
	// Allows hard-coding a dependencyName in rules. Is used if it is not captured via matchString.
	DependencyName string `json:"dependencyName" yaml:"dependencyName"`
	// Allows hard-coding a datasource in rules. Is used if it is not captured via matchString.
	Datasource common.DatasourceType `json:"datasource" yaml:"datasource"`
	// Allows defining regexes that replace further information from dependencies (like hash) after updating
	PostUpgradeReplacements []string `json:"postUpgradeReplacements" yaml:"postUpgradeReplacements"`
	// An optional name of a group to group dependency updates together.
	GroupName string `json:"groupName" yaml:"groupName"`
}

type Rule struct {
	Matches          *RuleMatch        `json:"matches" yaml:"matches"`
	ManagerConfig    *ManagerConfig    `json:"managerConfig" yaml:"managerConfig"`
	DependencyConfig *DependencyConfig `json:"dependencyConfig" yaml:"dependencyConfig"`
}

type RuleMatch struct {
	Managers        []string                `json:"managers" yaml:"managers"`
	ManagerTypes    []common.ManagerType    `json:"managerTypes" yaml:"managerTypes"`
	Files           []string                `json:"files" yaml:"files"`
	DependencyNames []string                `json:"dependencyNames" yaml:"dependencyNames"`
	Datasources     []common.DatasourceType `json:"datasources" yaml:"datasources"`
}

// A MatchAll rule is a rule that has no matches defined at all, so it will match everything.
func (rm *RuleMatch) IsMatchAll() bool {
	return rm == nil || (len(rm.Datasources) == 0 &&
		len(rm.DependencyNames) == 0 &&
		len(rm.Files) == 0 &&
		len(rm.ManagerTypes) == 0 &&
		len(rm.Managers) == 0)
}

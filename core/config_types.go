package core

import (
	"encoding/json"
	"os"
)

type Config struct {
	Platform         string            `json:"platform"`
	PlatformSettings *PlatformSettings `json:"platformSettings"`
	Extends          []string          `json:"extends"`
	IgnorePatterns   []string          `json:"ignorePatterns"`
	Managers         []*Manager        `json:"managers"`
	Rules            []*Rule           `json:"rules"`
	HostRules        []*HostRule       `json:"hostRules"`
}

func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}

type PlatformSettings struct {
	Token     string   `json:"token"`
	GitAuthor string   `json:"gitAuthor"`
	Endpoint  string   `json:"endpoint"`
	Direct    *bool    `json:"direct"`
	Projects  []string `json:"projects"`
}

func (ps *PlatformSettings) TokendExpanded() string {
	return os.ExpandEnv(ps.Token)
}

type Manager struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	// The settings are converted to rules to keep the right order, so they should not be used directly
	ManagerSettings *ManagerSettings `json:"managerSettings"`
	PackageSettings *PackageSettings `json:"packageSettings"`
}

func (m *Manager) String() string {
	b, _ := json.MarshalIndent(m, "", "  ")
	return string(b)
}

type Rule struct {
	Matches         *RuleMatch       `json:"matches"`
	ManagerSettings *ManagerSettings `json:"managerSettings"`
	PackageSettings *PackageSettings `json:"packageSettings"`
}

type RuleMatch struct {
	Managers     []string `json:"managers"`
	Files        []string `json:"files"`
	PackageNames []string `json:"packageNames"`
	Datasources  []string `json:"datasources"`
}

// A MatchAll rule is a rule that has no matches defined at all, so it will match all.
func (rm *RuleMatch) IsMatchAll() bool {
	return rm == nil || (len(rm.Managers) == 0 &&
		len(rm.Files) == 0 &&
		len(rm.PackageNames) == 0 &&
		len(rm.Datasources) == 0)
}

type ManagerSettings struct {
	// General settings
	Disabled     *bool    `json:"disabled"`
	FilePatterns []string `json:"filePatterns"`
	// Specific settings for RegexManager
	MatchStrings []string `json:"matchStrings"`
}

type PackageSettings struct {
	// Defines how much the dependency is allowed to update. Can be "major", "minor", or "patch".
	MaxUpdateType string `json:"maxUpdateType"`
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
	// Allows hard-coding a packageName in rules. Is used if it is not captured via matchString.
	PackageName string `json:"packageName"`
	// Allows hard-coding a datasource in rules. Is used if it is not captured via matchString.
	Datasource string `json:"datasource"`
	// Allows defining regexes that replace further information from packages (like hash) after updating
	PostUpgradeReplacements []string `json:"postUpgradeReplacements"`
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

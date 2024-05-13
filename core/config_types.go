package core

import (
	"encoding/json"
	"os"
)

type Config struct {
	Platform       string      `json:"platform"`
	Extends        []string    `json:"extends"`
	IgnorePatterns []string    `json:"ignorePatterns"`
	Managers       []*Manager  `json:"managers"`
	Rules          []*Rule     `json:"rules"`
	HostRules      []*HostRule `json:"hostRules"`
}

func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
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
	Managers    []string `json:"managers"`
	Files       []string `json:"files"`
	Packages    []string `json:"packages"`
	Datasources []string `json:"datasources"`
}

// A MatchAll rule is a rule that has no matches defined at all, so it will match all.
func (rm *RuleMatch) IsMatchAll() bool {
	return rm == nil || (len(rm.Managers) == 0 &&
		len(rm.Files) == 0 &&
		len(rm.Packages) == 0 &&
		len(rm.Datasources) == 0)
}

type ManagerSettings struct {
	// General settings
	Disabled     bool     `json:"disabled"`
	FilePatterns []string `json:"filePatterns"`
	// Specific settings for RegexManager
	MatchStrings []string `json:"matchStrings"`
}

type PackageSettings struct {
	MaxUpdateType     string   `json:"maxUpdateType"`
	AllowUnstable     *bool    `json:"allowUnstable"`
	RegistryUrls      []string `json:"registryUrls"`
	UseUnstable       *bool    `json:"useUnstable"`
	Versioning        string   `json:"versioning"`
	ExtractVersion    string   `json:"extractVersion"`
	IgnoreNonMatching *bool    `json:"ignoreNonMatching"`
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

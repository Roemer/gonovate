package core

import (
	"encoding/json"
)

type Config struct {
	Platform       string     `json:"platform"`
	Extends        []string   `json:"extends"`
	IgnorePatterns []string   `json:"ignorePatterns"`
	Managers       []*Manager `json:"managers"`
	Rules          []*Rule    `json:"rules"`
}

func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}

type Manager struct {
	Id           string   `json:"id"`
	Type         string   `json:"type"`
	MatchStrings []string `json:"matchStrings"`
	// The settings are converted to rules to keep the right order, so they should not be used outside
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
	Managers []string `json:"managers"`
	Files    []string `json:"files"`
}

type ManagerSettings struct {
	Disabled     bool     `json:"disabled"`
	FilePatterns []string `json:"filePatterns"`
}

type PackageSettings struct {
	MaxUpdateType string `json:"maxUpdateType"`
	AllowUnstable *bool  `json:"allowUnstable"`
}

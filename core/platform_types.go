package core

import "github.com/roemer/gover"

type Project struct {
}

// The interface for a manager specific change object
type IChange interface {
	GetMeta() *ChangeMeta
}

// Contains metadata for a change
type ChangeMeta struct {
	Datasource              string
	PackageName             string
	File                    string
	CurrentVersion          *gover.Version
	NewRelease              *ReleaseInfo
	PostUpgradeReplacements []string
	// Contains data that is generated while the change is processed and which is needed by other steps
	Data map[string]string
}

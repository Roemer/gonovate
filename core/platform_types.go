package core

import (
	"strings"

	"github.com/roemer/gover"
)

type Project struct {
	Path string
}

// Splits the path into "owner" and "repository"
func (p *Project) SplitPath() (string, string) {
	parts := strings.SplitN(p.Path, "/", 2)
	return parts[0], parts[1]
}

// Contains a list of changes which resulted from grouping changes
type ChangeSet struct {
	Title   string
	Id      string
	Changes []IChange
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
}

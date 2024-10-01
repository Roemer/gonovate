package common

type UpdateGroup struct {
	Title        string
	BranchName   string
	Dependencies []*Dependency
}

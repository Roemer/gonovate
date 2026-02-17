package common

type UpdateGroup struct {
	Title        string
	BranchName   string
	Dependencies []*DependencyWithUpdate
	Labels       []string
	Reviewers    []string
}

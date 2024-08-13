package shared

type UpdateGroup struct {
	Title        string
	BranchName   string
	Dependencies []*Dependency
}

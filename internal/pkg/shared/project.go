package shared

import "strings"

type Project struct {
	Path string
}

// Splits the path into "owner" and "repository"
func (p *Project) SplitPath() (string, string) {
	parts := strings.SplitN(p.Path, "/", 2)
	return parts[0], parts[1]
}

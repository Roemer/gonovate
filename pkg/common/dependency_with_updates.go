package common

// Contains a dependency and the information about the update to this dependency.
type DependencyWithUpdate struct {
	Dependency *Dependency
	NewRelease *ReleaseInfo
}

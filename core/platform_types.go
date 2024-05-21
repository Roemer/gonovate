package core

type Project struct {
}

type Change struct {
	PackageName string
	OldVersion  string
	NewVersion  string
	// Contains data that is generated while the change is processed and which is needed by other steps
	Data map[string]string
}

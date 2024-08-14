package config

import "github.com/roemer/gonovate/internal/pkg/shared"

// Checks if a project has its own configuration.
func HasProjectConfig() (bool, error) {
	return shared.FileExists("gonovate.json")
}

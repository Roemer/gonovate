package config

import "github.com/roemer/gonovate/internal/pkg/shared"

func HasProjectConfig() (bool, error) {
	return shared.FileExists("gonovate.json")
}

package shared

import (
	"time"

	"github.com/roemer/gover"
)

type ReleaseInfo struct {
	ReleaseDate   time.Time
	Version       *gover.Version
	VersionString string
	Hashes        map[string]string
}

package shared

import (
	"time"

	"github.com/roemer/gover"
)

type ReleaseInfo struct {
	ReleaseDate   time.Time
	Version       *gover.Version
	VersionString string
	// Can contain additional data for the release like hashes/digest
	AdditionalData map[string]string
}

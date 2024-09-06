package common

import (
	"time"

	"github.com/roemer/gover"
)

// This type contains information about a release of a dependency
type ReleaseInfo struct {
	// The time when the release was created
	ReleaseDate time.Time
	// The original versionstring of the release
	VersionString string
	// The parsed version of the release
	Version *gover.Version
	// The digest of the release
	Digest string
	// Can contain additional data for the release like hashes
	AdditionalData map[string]string
}

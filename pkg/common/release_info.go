package common

import (
	"time"

	"github.com/roemer/gover"
)

// This type contains information about a release of a dependency
type ReleaseInfo struct {
	// The time when the release was created
	ReleaseDate time.Time `json:"releaseDate,omitzero"`
	// The original versionstring of the release
	VersionString string `json:"versionString"`
	// The parsed version of the release
	Version *gover.Version `json:"-"`
	// The digest of the release
	Digest string `json:"digest,omitempty"`
	// Can contain additional data for the release like hashes or urls
	AdditionalData map[string]string `json:"additionalData,omitempty"`
}

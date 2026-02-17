package common

import (
	"maps"
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
	// The type of the update this release would be. Can be "major", "minor", or "patch"
	UpdateType UpdateType `json:"-"`
	// Can contain additional data for the release like hashes or urls
	AdditionalData map[string]string `json:"additionalData,omitempty"`
}

func (r *ReleaseInfo) Clone() *ReleaseInfo {
	if r == nil {
		return nil
	}
	var additionalDataCopy map[string]string
	if r.AdditionalData != nil {
		additionalDataCopy = make(map[string]string, len(r.AdditionalData))
		maps.Copy(additionalDataCopy, r.AdditionalData)
	}
	return &ReleaseInfo{
		ReleaseDate:    r.ReleaseDate,
		VersionString:  r.VersionString,
		Version:        r.Version, // Version is immutable, so we can share it
		Digest:         r.Digest,
		UpdateType:     r.UpdateType,
		AdditionalData: additionalDataCopy,
	}
}

package cache

import (
	"time"

	"github.com/roemer/gonovate/pkg/common"
)

type cacheInfo struct {
	DatasourceType  common.DatasourceType // `json:"datasourceType"`
	CacheIdentifier string                // `json:"cacheIdentifier"`
	FetchedAt       time.Time             // `json:"fetchedAt"`
	ExpiresAt       time.Time             // `json:"expiresAt"`
	Releases        []*cacheRelease       // `json:"releases"`
}

type cacheRelease struct {
	ReleaseDate    time.Time         // `json:"releaseDate"`
	VersionString  string            // `json:"versionString"`
	Digest         string            // `json:"digest"`
	AdditionalData map[string]string // `json:"additionalData"`
}

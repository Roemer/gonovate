package common

import (
	"log/slog"

	"github.com/roemer/gonovate/pkg/cache"
)

type GonovateCache struct {
	ReleaseCache      *cache.FileCache[[]*ReleaseInfo]
	GitLabUserIdCache *cache.MemoryCache[int64]
}

func NewGonovateCache(cacheDir string, logger *slog.Logger) *GonovateCache {
	return &GonovateCache{
		ReleaseCache:      cache.NewFileCache[[]*ReleaseInfo](cacheDir, logger),
		GitLabUserIdCache: cache.NewMemoryCache[int64](logger),
	}
}

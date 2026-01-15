package cache

import (
	"log/slog"

	"github.com/roemer/gonovate/pkg/common"
)

type GonovateCache struct {
	fileCache *fileCache
}

func NewGonovateCache(cacheDir string, logger *slog.Logger) *GonovateCache {
	return &GonovateCache{
		fileCache: &fileCache{
			CacheDir: cacheDir,
			Logger:   logger,
		},
	}
}

func (c *GonovateCache) Get(datasourceType common.DatasourceType, cacheIdentifier string) ([]*common.ReleaseInfo, error) {
	return c.fileCache.Get(datasourceType, cacheIdentifier)
}

func (c *GonovateCache) Set(datasourceType common.DatasourceType, cacheIdentifier string, releases []*common.ReleaseInfo) error {
	return c.fileCache.Set(datasourceType, cacheIdentifier, releases)
}

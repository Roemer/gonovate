package cache

import "github.com/roemer/gonovate/internal/core"

var DatasourceCache *datasourceCache = &datasourceCache{entries: map[core.DatasourceType]map[string][]*core.ReleaseInfo{}}

type datasourceCache struct {
	entries map[core.DatasourceType]map[string][]*core.ReleaseInfo
}

func (cache *datasourceCache) GetCache(datasourceType core.DatasourceType, identifier string) []*core.ReleaseInfo {
	entriesForDatasource, ok := cache.entries[datasourceType]
	if !ok {
		return nil
	}
	entriesForId, ok := entriesForDatasource[identifier]
	if !ok {
		return nil
	}
	return entriesForId
}

func (cache *datasourceCache) SetCache(datasourceType core.DatasourceType, identifier string, versions []*core.ReleaseInfo) {
	if _, ok := cache.entries[datasourceType]; !ok {
		cache.entries[datasourceType] = map[string][]*core.ReleaseInfo{}
	}
	cache.entries[datasourceType][identifier] = versions
}

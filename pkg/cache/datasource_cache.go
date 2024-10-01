package cache

import "github.com/roemer/gonovate/internal/pkg/shared"

var DatasourceCache *datasourceCache = &datasourceCache{entries: map[shared.DatasourceType]map[string][]*shared.ReleaseInfo{}}

type datasourceCache struct {
	entries map[shared.DatasourceType]map[string][]*shared.ReleaseInfo
}

func (cache *datasourceCache) GetCache(datasourceType shared.DatasourceType, identifier string) []*shared.ReleaseInfo {
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

func (cache *datasourceCache) SetCache(datasourceType shared.DatasourceType, identifier string, versions []*shared.ReleaseInfo) {
	if _, ok := cache.entries[datasourceType]; !ok {
		cache.entries[datasourceType] = map[string][]*shared.ReleaseInfo{}
	}
	cache.entries[datasourceType][identifier] = versions
}

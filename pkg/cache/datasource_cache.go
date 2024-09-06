package cache

import "github.com/roemer/gonovate/pkg/common"

var DatasourceCache *datasourceCache = &datasourceCache{entries: map[common.DatasourceType]map[string][]*common.ReleaseInfo{}}

type datasourceCache struct {
	entries map[common.DatasourceType]map[string][]*common.ReleaseInfo
}

func (cache *datasourceCache) GetCache(datasourceType common.DatasourceType, identifier string) []*common.ReleaseInfo {
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

func (cache *datasourceCache) SetCache(datasourceType common.DatasourceType, identifier string, versions []*common.ReleaseInfo) {
	if _, ok := cache.entries[datasourceType]; !ok {
		cache.entries[datasourceType] = map[string][]*common.ReleaseInfo{}
	}
	cache.entries[datasourceType][identifier] = versions
}

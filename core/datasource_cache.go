package core

var DatasourceCache *datasourceCache = &datasourceCache{entries: map[DatasourceType]map[string][]*ReleaseInfo{}}

type datasourceCache struct {
	entries map[DatasourceType]map[string][]*ReleaseInfo
}

func (cache *datasourceCache) GetCache(datasourceType DatasourceType, identifier string) []*ReleaseInfo {
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

func (cache *datasourceCache) SetCache(datasourceType DatasourceType, identifier string, versions []*ReleaseInfo) {
	if _, ok := cache.entries[datasourceType]; !ok {
		cache.entries[datasourceType] = map[string][]*ReleaseInfo{}
	}
	cache.entries[datasourceType][identifier] = versions
}

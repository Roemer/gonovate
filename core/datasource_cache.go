package core

import "github.com/roemer/gover"

var DatasourceCache *datasourceCache = &datasourceCache{entries: map[string]map[string][]*gover.Version{}}

type datasourceCache struct {
	entries map[string]map[string][]*gover.Version
}

func (cache *datasourceCache) GetCache(datasourceType string, identifier string) []*gover.Version {
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

func (cache *datasourceCache) SetCache(datasourceType string, identifier string, versions []*gover.Version) {
	if _, ok := cache.entries[datasourceType]; !ok {
		cache.entries[datasourceType] = map[string][]*gover.Version{}
	}
	cache.entries[datasourceType][identifier] = versions
}

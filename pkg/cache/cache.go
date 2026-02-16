package cache

import "time"

type cacheEntry[T any] struct {
	FetchedAt time.Time `json:"fetchedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	CacheData T         `json:"cacheData"`
}

type Cache[T any] interface {
	Get(cacheIdentifier string) (T, bool, error)
	Set(cacheIdentifier string, objectToCache T, ttl time.Duration) error
	Clear(cacheIdentifier string) error
}

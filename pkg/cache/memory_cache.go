package cache

import (
	"log/slog"
	"time"
)

// Make sure the cache interface is implemented
var _ Cache[any] = (*MemoryCache[any])(nil)

type MemoryCache[T any] struct {
	entries map[string]*cacheEntry[T]
	Logger  *slog.Logger
}

func NewMemoryCache[T any](logger *slog.Logger) *MemoryCache[T] {
	return &MemoryCache[T]{
		entries: make(map[string]*cacheEntry[T]),
		Logger:  logger,
	}
}

func (c *MemoryCache[T]) Get(cacheIdentifier string) (T, bool, error) {
	var zeroValue T
	entry, exists := c.entries[cacheIdentifier]
	if !exists {
		c.Logger.Debug("Cache miss", "cacheIdentifier", cacheIdentifier)
		return zeroValue, false, nil
	}
	if entry.ExpiresAt.Before(time.Now()) {
		c.Logger.Debug("Cache expired", "cacheIdentifier", cacheIdentifier)
		// Cache expired, clear it and return nil
		delete(c.entries, cacheIdentifier)
		return zeroValue, false, nil
	}
	c.Logger.Debug("Cache hit", "cacheIdentifier", cacheIdentifier)
	return entry.CacheData, true, nil
}

func (c *MemoryCache[T]) Set(cacheIdentifier string, objectToCache T, ttl time.Duration) error {
	c.entries[cacheIdentifier] = &cacheEntry[T]{
		FetchedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
		CacheData: objectToCache,
	}
	return nil
}

func (c *MemoryCache[T]) Clear(cacheIdentifier string) error {
	delete(c.entries, cacheIdentifier)
	return nil
}

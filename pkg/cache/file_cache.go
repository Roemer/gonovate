package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// Make sure the cache interface is implemented
var _ Cache[any] = (*FileCache[any])(nil)

type FileCache[T any] struct {
	cacheDir string
	Logger   *slog.Logger
}

func NewFileCache[T any](cacheDir string, logger *slog.Logger) *FileCache[T] {
	return &FileCache[T]{
		cacheDir: cacheDir,
		Logger:   logger,
	}
}

func (c *FileCache[T]) Get(cacheIdentifier string) (T, bool, error) {
	var zeroValue T
	cacheFilePath := c.getCacheFilePath(cacheIdentifier)
	// Check if the file exists and if so, open and read it
	fileDescriptor, err := os.Open(cacheFilePath)
	if errors.Is(err, os.ErrNotExist) {
		// File does not exist
		c.Logger.Debug("Cache miss", "cacheIdentifier", cacheIdentifier)
		return zeroValue, false, nil
	} else if err != nil {
		// Error while reading the file
		return zeroValue, false, fmt.Errorf("error reading the cache file '%s': %w", cacheFilePath, err)
	}
	// Read the file
	defer fileDescriptor.Close()
	var dataObject cacheEntry[T]
	if err := json.NewDecoder(fileDescriptor).Decode(&dataObject); err != nil {
		return zeroValue, false, fmt.Errorf("error converting the cache file '%s' to json: %w", cacheFilePath, err)
	}
	if dataObject.ExpiresAt.Before(time.Now()) {
		c.Logger.Debug("Cache expired", "cacheIdentifier", cacheIdentifier)
		// Cache expired, clean it and return nil
		if err := c.Clear(cacheIdentifier); err != nil {
			return zeroValue, false, fmt.Errorf("error clearing expired cache file '%s': %w", cacheFilePath, err)
		}
		return zeroValue, false, nil
	}
	c.Logger.Debug("Cache hit", "cacheIdentifier", cacheIdentifier)
	return dataObject.CacheData, true, nil
}

func (c *FileCache[T]) Set(cacheIdentifier string, objectToCache T, ttl time.Duration) error {
	cacheFilePath := c.getCacheFilePath(cacheIdentifier)
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(cacheFilePath), os.ModePerm); err != nil {
		return fmt.Errorf("error creating the cache directory for file '%s': %w", cacheFilePath, err)
	}
	// Open the file for writing
	fileDescriptor, err := os.Create(cacheFilePath)
	if err != nil {
		return fmt.Errorf("error creating the cache file '%s': %w", cacheFilePath, err)
	}
	defer fileDescriptor.Close()
	// Create the cache info object
	dataObject := cacheEntry[T]{
		FetchedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
		CacheData: objectToCache,
	}
	// Write the file
	encoder := json.NewEncoder(fileDescriptor)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&dataObject); err != nil {
		return fmt.Errorf("error writing the cache file '%s': %w", cacheFilePath, err)
	}
	return nil
}

func (c *FileCache[T]) Clear(cacheIdentifier string) error {
	cacheFilePath := c.getCacheFilePath(cacheIdentifier)
	if err := os.Remove(cacheFilePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error deleting the cache file '%s': %w", cacheFilePath, err)
	}
	return nil
}

func (c *FileCache[T]) getCacheFilePath(cacheIdentifier string) string {
	// Sanitize the cache identifier to create a valid file name by removing invalid characters
	// But we keep / to allow for subdirectories in the cache
	sanitizedIdentifier := NormalizeFilePath(cacheIdentifier, true)
	return filepath.Join(c.cacheDir, sanitizedIdentifier+".json")
}

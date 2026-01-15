package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/samber/lo"
)

type fileCache struct {
	CacheDir string
	Logger   *slog.Logger
}

func (c *fileCache) Get(datasourceType common.DatasourceType, cacheIdentifier string) ([]*common.ReleaseInfo, error) {
	cacheFilePath := c.getCacheFilePath(datasourceType, cacheIdentifier)
	// Check if the file exists and if so, open and read it
	fileDescriptor, err := os.Open(cacheFilePath)
	if errors.Is(err, os.ErrNotExist) {
		// File does not exist
		return nil, nil
	} else if err != nil {
		// Error while reading the file
		return nil, fmt.Errorf("error reading the cache file '%s': %w", cacheFilePath, err)
	}
	// Read the file
	defer fileDescriptor.Close()
	var dataObject cacheInfo
	if err := json.NewDecoder(fileDescriptor).Decode(&dataObject); err != nil {
		return nil, fmt.Errorf("error converting the cache file '%s' to json: %w", cacheFilePath, err)
	}
	if dataObject.ExpiresAt.Before(time.Now()) {
		// Cache expired
		return nil, nil
	}
	if dataObject.CacheIdentifier != cacheIdentifier {
		// The identifier does not match (cleaned identifier might lead to a duplicate)
		c.Logger.Warn(fmt.Sprintf("Cache identifier mismatch for file '%s': expected '%s', got '%s'", cacheFilePath, cacheIdentifier, dataObject.CacheIdentifier))
		return nil, nil
	}
	// Convert the releases and return them
	mappedReleases := lo.Map(dataObject.Releases, func(item *cacheRelease, index int) *common.ReleaseInfo {
		return &common.ReleaseInfo{
			ReleaseDate:    item.ReleaseDate,
			VersionString:  item.VersionString,
			Digest:         item.Digest,
			AdditionalData: item.AdditionalData,
		}
	})
	return mappedReleases, nil
}

func (c *fileCache) Set(datasourceType common.DatasourceType, cacheIdentifier string, releases []*common.ReleaseInfo) error {
	cacheFilePath := c.getCacheFilePath(datasourceType, cacheIdentifier)
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
	// Convert the releases
	mappedReleases := lo.Map(releases, func(item *common.ReleaseInfo, index int) *cacheRelease {
		return &cacheRelease{
			ReleaseDate:    item.ReleaseDate,
			VersionString:  item.VersionString,
			Digest:         item.Digest,
			AdditionalData: item.AdditionalData,
		}
	})
	// Create the cache info object
	dataObject := cacheInfo{
		DatasourceType:  datasourceType,
		CacheIdentifier: cacheIdentifier,
		FetchedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(10 * time.Minute),
		Releases:        mappedReleases,
	}
	// Write the file
	encoder := json.NewEncoder(fileDescriptor)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&dataObject); err != nil {
		return fmt.Errorf("error writing the cache file '%s': %w", cacheFilePath, err)
	}
	return nil
}

func (c *fileCache) getCacheFilePath(datasourceType common.DatasourceType, cacheIdentifier string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9*,\. ]+`)
	clearedIdentifier := re.ReplaceAllString(cacheIdentifier, "-")
	return filepath.Join(c.CacheDir, string(datasourceType), clearedIdentifier+".json")
}

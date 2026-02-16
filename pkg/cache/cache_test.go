package cache

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCache_BasicOperations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cache := NewMemoryCache[string](logger)
	key := "test-key"
	value := "test-value"

	// Should not exist initially
	v, ok, err := cache.Get(key)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, v)

	// Set value
	err = cache.Set(key, value, time.Second)
	assert.NoError(t, err)

	// Should exist now
	v, ok, err = cache.Get(key)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, value, v)

	// Clear value
	err = cache.Clear(key)
	assert.NoError(t, err)
	v, ok, err = cache.Get(key)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestMemoryCache_Expiration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cache := NewMemoryCache[string](logger)
	key := "expire-key"
	value := "expire-value"
	err := cache.Set(key, value, 100*time.Millisecond)
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond)
	v, ok, err := cache.Get(key)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, v)
}

func TestFileCache_BasicOperations(t *testing.T) {
	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cache := NewFileCache[string](tmpDir, logger)
	key := "file-key"
	value := "file-value"

	// Should not exist initially
	v, ok, err := cache.Get(key)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, v)

	// Set value
	err = cache.Set(key, value, time.Second)
	assert.NoError(t, err)

	// Should exist now
	v, ok, err = cache.Get(key)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, value, v)

	// Clear value
	err = cache.Clear(key)
	assert.NoError(t, err)
	v, ok, err = cache.Get(key)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestFileCache_Expiration(t *testing.T) {
	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cache := NewFileCache[string](tmpDir, logger)
	key := "expire-file-key"
	value := "expire-file-value"
	err := cache.Set(key, value, 10*time.Millisecond)
	assert.NoError(t, err)
	time.Sleep(20 * time.Millisecond)
	v, ok, err := cache.Get(key)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, v)
	// File should be deleted after expiration
	filePath := filepath.Join(tmpDir, key+".json")
	_, statErr := os.Stat(filePath)
	assert.True(t, os.IsNotExist(statErr))
}

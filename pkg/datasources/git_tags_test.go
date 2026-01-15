package datasources

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitTagsDatasource_GetReleases(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	cacheDir := filepath.Join(os.TempDir(), "gonovate_cache")
	cache := common.NewGonovateCache(cacheDir, logger)
	ds := NewGitTagsDatasource(&common.DatasourceSettings{Logger: logger, Cache: cache.ReleaseCache})
	dep := &common.Dependency{
		Name:          "https://github.com/Roemer/gonovate.git",
		Version:       "v0.0.3",
		Versioning:    `v([0-9]+)\.([0-9]+)\.([0-9]+)`,
		MaxUpdateType: common.UPDATE_TYPE_MAJOR,
	}
	releases, err := ds.GetReleases(dep)
	require.NoError(t, err)
	assert.Greater(t, len(releases), 0)

	releaseInfo, err := ds.SearchDependencyUpdate(dep)
	require.NoError(t, err)
	require.NotNil(t, releaseInfo)

	// Verify that the releases list contains v0.6.4 and v0.15.0
	hasVersion := func(v string) bool {
		for _, r := range releases {
			if r.VersionString == v {
				return true
			}
		}
		return false
	}
	assert.True(t, hasVersion("v0.6.4"), "v0.6.4 not found in releases")
	assert.True(t, hasVersion("v0.15.0"), "v0.15.0 not found in releases")
}

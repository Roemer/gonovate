package datasources

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
	"github.com/stretchr/testify/assert"
)

func TestBrowserVersionChrome(t *testing.T) {
	t.Skip("This test is for local debugging only")

	assert := assert.New(t)

	// Load the defaults
	cfg, err := config.Load("preset:defaults")
	assert.NoError(err)
	assert.NotNil(cfg)

	// Create the datasource
	ds := NewBrowserVersionDatasource(slog.Default(), cfg)

	// Create the dependency and enrich it with rules from the config
	chrome := &shared.Dependency{Name: "chrome", Datasource: shared.DATASOURCE_TYPE_BROWSERVERSION, Version: "126.0.0.0"}
	cfg.EnrichDependencyFromRules(chrome)

	// Get the releases from the datasource
	releases, err := ds.getReleases(chrome)
	assert.NoError(err)
	assert.NotNil(releases)
	fmt.Println("Found releases:")
	for _, release := range releases {
		fmt.Println(release.VersionString)
	}

	// Search for an update
	ri, ver, err := ds.SearchDependencyUpdate(chrome)
	assert.NoError(err)
	assert.NotNil(ri)
	assert.NotNil(ver)
	fmt.Println("Update found to version:")
	fmt.Println(ri.VersionString)
}

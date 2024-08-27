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
	dep := &shared.Dependency{Name: "chrome", Datasource: shared.DATASOURCE_TYPE_BROWSERVERSION, Version: "126.0.0.0"}
	cfg.EnrichDependencyFromRules(dep)

	// Get the releases from the datasource
	releases, err := ds.getReleases(dep)
	assert.NoError(err)
	assert.NotNil(releases)
	fmt.Println("Found releases:")
	for _, release := range releases {
		fmt.Println(release.VersionString)
	}

	// Search for an update
	ri, ver, err := ds.SearchDependencyUpdate(dep)
	assert.NoError(err)
	assert.NotNil(ri)
	assert.NotNil(ver)
	fmt.Println("Update found to version:")
	fmt.Println(ri.VersionString)
}

func TestDockerDigest(t *testing.T) {
	t.Skip("This test is for local debugging only")

	assert := assert.New(t)

	// Load the defaults
	cfg, err := config.Load("preset:defaults")
	assert.NoError(err)
	assert.NotNil(cfg)

	// Create the datasource
	ds := NewDockerDatasource(slog.Default(), cfg)

	fmt.Println("Vaultwarden")
	{
		dep := &shared.Dependency{Name: "vaultwarden/server", Datasource: shared.DATASOURCE_TYPE_DOCKER, Version: "1.30.3", IgnoreNonMatching: shared.TruePtr}
		cfg.EnrichDependencyFromRules(dep)

		digest, err := ds.(*DockerDatasource).getAdditionalData(dep, &shared.ReleaseInfo{VersionString: "1.30.3"}, "digest")
		assert.NoError(err)
		fmt.Println(digest)
	}

	fmt.Println("ut99")
	{
		dep := &shared.Dependency{Name: "roemer/ut99-server", Datasource: shared.DATASOURCE_TYPE_DOCKER, Version: "latest", IgnoreNonMatching: shared.TruePtr}
		cfg.EnrichDependencyFromRules(dep)

		digest, err := ds.(*DockerDatasource).getAdditionalData(dep, &shared.ReleaseInfo{VersionString: "latest"}, "digest")
		assert.NoError(err)
		fmt.Println(digest)
	}
}

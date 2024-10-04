package integration_tests

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/roemer/gonovate/pkg/config"
	"github.com/roemer/gonovate/pkg/datasources"
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
	settings := &common.DatasourceSettings{
		Logger:    slog.Default(),
		HostRules: cfg.HostRules,
	}
	ds := datasources.NewBrowserVersionDatasource(settings)

	// Create the dependency and enrich it with rules from the config
	dep := &common.Dependency{Name: "chrome", Datasource: common.DATASOURCE_TYPE_BROWSERVERSION, Version: "126.0.0.0"}
	err = cfg.ApplyToDependency(dep)
	assert.NoError(err)

	// Get the releases from the datasource
	releases, err := ds.GetReleases(dep)
	assert.NoError(err)
	assert.NotNil(releases)
	fmt.Println("Found releases:")
	for _, release := range releases {
		fmt.Println(release.VersionString)
	}

	// Search for an update
	ri, err := ds.SearchDependencyUpdate(dep)
	assert.NoError(err)
	assert.NotNil(ri)
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
	settings := &common.DatasourceSettings{
		Logger:    slog.Default(),
		HostRules: cfg.HostRules,
	}
	ds := datasources.NewDockerDatasource(settings)

	fmt.Println("Vaultwarden")
	{
		dep := &common.Dependency{Name: "vaultwarden/server", Datasource: common.DATASOURCE_TYPE_DOCKER, Version: "1.30.3", IgnoreNonMatching: common.TruePtr}
		err = cfg.ApplyToDependency(dep)
		assert.NoError(err)

		digest, err := ds.(*datasources.DockerDatasource).GetDigest(dep, "1.30.3")
		assert.NoError(err)
		fmt.Println(digest)
	}

	fmt.Println("ut99")
	{
		dep := &common.Dependency{Name: "roemer/ut99-server", Datasource: common.DATASOURCE_TYPE_DOCKER, Version: "latest", IgnoreNonMatching: common.TruePtr}
		err = cfg.ApplyToDependency(dep)
		assert.NoError(err)

		digest, err := ds.(*datasources.DockerDatasource).GetDigest(dep, "latest")
		assert.NoError(err)
		fmt.Println(digest)
	}
}

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
	configLoader := config.NewConfigLoader(slog.Default())
	cfg, err := configLoader.Load("preset:defaults")
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
	configLoader := config.NewConfigLoader(slog.Default())
	cfg, err := configLoader.Load("preset:defaults")
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

func TestGitLabPackages(t *testing.T) {
	t.Skip("This test is for local debugging only")

	assert := assert.New(t)

	// Load the defaults
	configLoader := config.NewConfigLoader(slog.Default())
	cfg, err := configLoader.Load("preset:defaults")
	assert.NoError(err)
	assert.NotNil(cfg)

	// Create the datasource
	settings := &common.DatasourceSettings{
		Logger:    slog.Default(),
		HostRules: cfg.HostRules,
	}
	ds := datasources.NewGitLabPackagesDatasource(settings)

	// Create the dependency and enrich it with rules from the config
	dep := &common.Dependency{Name: "gitlab-org/release-cli:release-cli", Datasource: common.DATASOURCE_TYPE_GITLAB_PACKAGES, Version: "0.18.0"}
	//dep.ExtractVersion
	dep.IgnoreNonMatching = common.TruePtr
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

func TestGoVersion(t *testing.T) {
	t.Skip("This test is for local debugging only")

	assert := assert.New(t)

	// Load the defaults
	configLoader := config.NewConfigLoader(slog.Default())
	cfg, err := configLoader.Load("preset:defaults")
	assert.NoError(err)
	assert.NotNil(cfg)

	// Create the datasource
	settings := &common.DatasourceSettings{
		Logger:    slog.Default(),
		HostRules: cfg.HostRules,
	}
	ds := datasources.NewGoVersionDatasource(settings)

	// Create the dependency and enrich it with rules from the config
	dep := &common.Dependency{Name: "go-stable", Datasource: common.DATASOURCE_TYPE_GOVERSION, Version: "1.23.0"}
	dep.MaxUpdateType = common.UPDATE_TYPE_PATCH
	//dep.ExtractVersion
	dep.IgnoreNonMatching = common.TruePtr
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

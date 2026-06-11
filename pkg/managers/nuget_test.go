package managers

import (
	"log/slog"
	"testing"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestNugetManagerExtract(t *testing.T) {
	assert := assert.New(t)

	manager := NewNugetManager("manager", &common.ManagerSettings{
		Logger: slog.Default(),
	})
	dependencies, err := manager.ExtractDependencies(`../../testdata/nuget/a/sample.csproj`)
	assert.NoError(err)
	assert.NotNil(dependencies)
	assert.Len(dependencies, 2)

	assert.Equal("Newtonsoft.Json", dependencies[0].Name)
	assert.Equal("13.0.3", dependencies[0].Version)
	assert.Equal(common.DATASOURCE_TYPE_NUGET, dependencies[0].Datasource)

	assert.Equal("Serilog", dependencies[1].Name)
	assert.Equal("3.1.1", dependencies[1].Version)
	assert.Equal(common.DATASOURCE_TYPE_NUGET, dependencies[1].Datasource)
}

func TestNugetManagerExtractVersionElement(t *testing.T) {
	assert := assert.New(t)

	nugetManager := NewNugetManager("manager", &common.ManagerSettings{
		Logger: slog.Default(),
	}).(*NugetManager)

	fileContent := `<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json">
      <Version>13.0.3</Version>
    </PackageReference>
  </ItemGroup>
</Project>`

	dependencies, err := nugetManager.extractDependenciesFromString(fileContent, "test.csproj")
	assert.NoError(err)
	assert.NotNil(dependencies)
	assert.Len(dependencies, 1)

	assert.Equal("Newtonsoft.Json", dependencies[0].Name)
	assert.Equal("13.0.3", dependencies[0].Version)
	assert.Equal(common.DATASOURCE_TYPE_NUGET, dependencies[0].Datasource)
}

func TestNugetManagerExtractSkipsMissingVersion(t *testing.T) {
	assert := assert.New(t)

	nugetManager := NewNugetManager("manager", &common.ManagerSettings{
		Logger: slog.Default(),
	}).(*NugetManager)

	fileContent := `<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="13.0.3" />
    <PackageReference Include="NoVersionPackage" />
  </ItemGroup>
</Project>`

	dependencies, err := nugetManager.extractDependenciesFromString(fileContent, "test.csproj")
	assert.NoError(err)
	assert.NotNil(dependencies)
	assert.Len(dependencies, 1)
	assert.Equal("Newtonsoft.Json", dependencies[0].Name)
}

package managers

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/roemer/gonovate/pkg/common"
)

type NugetManager struct {
	*managerBase
}

func NewNugetManager(id string, settings *common.ManagerSettings) common.IManager {
	manager := &NugetManager{
		managerBase: newManagerBase(id, common.MANAGER_TYPE_NUGET, settings),
	}
	manager.impl = manager
	return manager
}

func (manager *NugetManager) ExtractDependencies(filePath string) ([]*common.Dependency, error) {
	// Read the entire file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileContent := string(fileContentBytes)

	// Extract the dependencies from the string
	return manager.extractDependenciesFromString(fileContent, filePath)
}

func (manager *NugetManager) ApplyDependencyUpdate(dependency *common.Dependency, newRelease *common.ReleaseInfo) error {
	return replaceDependencyVersionInFileWithCheck(dependency, newRelease, func(dependency *common.Dependency, newFileContent string) (*common.Dependency, error) {
		newDeps, err := manager.extractDependenciesFromString(newFileContent, dependency.FilePath)
		if err != nil {
			return nil, err
		}
		newDep, err := manager.getSingleDependency(dependency.Name, newDeps)
		if err != nil {
			return nil, err
		}
		return newDep, nil
	})
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

func (manager *NugetManager) extractDependenciesFromString(fileContent string, filePath string) ([]*common.Dependency, error) {
	// A slice to collect all found dependencies
	foundDependencies := []*common.Dependency{}

	// Parse the XML content
	var project nugetProject
	if err := xml.Unmarshal([]byte(fileContent), &project); err != nil {
		return nil, fmt.Errorf("failed to parse XML in file %s: %w", filePath, err)
	}

	// Extract all package references from all item groups
	for _, itemGroup := range project.ItemGroups {
		for _, pkg := range itemGroup.PackageReferences {
			if pkg.Include == "" {
				continue
			}
			// Version can be either an attribute or a child element
			version := pkg.VersionAttr
			if version == "" {
				version = pkg.VersionElement
			}
			if version == "" {
				continue
			}
			newDependency := manager.newDependency(pkg.Include, common.DATASOURCE_TYPE_NUGET, version, filePath)
			foundDependencies = append(foundDependencies, newDependency)
		}
	}

	return foundDependencies, nil
}

// XML structures for .NET project files (.csproj, .vbproj, .fsproj)
type nugetProject struct {
	ItemGroups []nugetItemGroup `xml:"ItemGroup"`
}

type nugetItemGroup struct {
	PackageReferences []nugetPackageReference `xml:"PackageReference"`
}

type nugetPackageReference struct {
	Include        string `xml:"Include,attr"`
	VersionAttr    string `xml:"Version,attr"`
	VersionElement string `xml:"Version"`
}

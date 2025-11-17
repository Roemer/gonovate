package managers

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/roemer/gonovate/pkg/common"
)

type HelmManager struct {
	*managerBase
}

func NewHelmManager(id string, settings *common.ManagerSettings) common.IManager {
	manager := &HelmManager{
		managerBase: newManagerBase(id, common.MANAGER_TYPE_HELM, settings),
	}
	manager.impl = manager
	return manager
}

func (manager *HelmManager) ExtractDependencies(filePath string) ([]*common.Dependency, error) {
	// Read the entire file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileContent := string(fileContentBytes)

	// Extract the dependencies from the string
	return manager.extractDependenciesFromString(fileContent, filePath)
}

func (manager *HelmManager) ApplyDependencyUpdate(dependency *common.Dependency) error {
	return replaceDependencyVersionInFileWithCheck(dependency, func(dependency *common.Dependency, newFileContent string) (*common.Dependency, error) {
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

func (manager *HelmManager) extractDependenciesFromString(fileContent string, filePath string) ([]*common.Dependency, error) {
	// A slice to collect all found dependencies
	foundDependencies := []*common.Dependency{}

	// Decode the file
	dec := yaml.NewDecoder(strings.NewReader(fileContent))
	for {
		var yamlObject helmFile
		if err := dec.Decode(&yamlObject); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decoding error: %s", err)
		}
		// Process it
		for _, dependency := range yamlObject.Dependencies {
			newDependency := manager.newDependency(dependency.Name, common.DATASOURCE_TYPE_HELM, dependency.Version, filePath)
			newDependency.RegistryUrls = []string{dependency.Repository}
			foundDependencies = append(foundDependencies, newDependency)
		}
	}

	// Return the found dependencies
	return foundDependencies, nil
}

type helmFile struct {
	Dependencies []struct {
		Name       string `yaml:"name"`
		Version    string `yaml:"version"`
		Repository string `yaml:"repository"`
	} `yaml:"dependencies"`
}

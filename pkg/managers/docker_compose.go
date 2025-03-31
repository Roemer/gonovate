package managers

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/roemer/gonovate/pkg/common"
)

type DockerComposeManager struct {
	*managerBase
}

func NewDockerComposeManager(id string, settings *common.ManagerSettings) common.IManager {
	manager := &DockerComposeManager{
		managerBase: newManagerBase(id, common.MANAGER_TYPE_DOCKER_COMPOSE, settings),
	}
	manager.impl = manager
	return manager
}

func (manager *DockerComposeManager) ExtractDependencies(filePath string) ([]*common.Dependency, error) {
	// Read the entire file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileContent := string(fileContentBytes)

	// Extract the dependencies from the string
	return manager.extractDependenciesFromString(fileContent, filePath)
}

func (manager *DockerComposeManager) ApplyDependencyUpdate(dependency *common.Dependency) error {
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

func (manager *DockerComposeManager) extractDependenciesFromString(fileContent string, filePath string) ([]*common.Dependency, error) {
	// A slice to collect all found dependencies
	foundDependencies := []*common.Dependency{}

	// Decode the file
	dockerComposeFile := &dockerComposeFile{}
	if err := yaml.Unmarshal([]byte(fileContent), dockerComposeFile); err != nil {
		return nil, fmt.Errorf("failed parsing file '%s': %w", filePath, err)
	}
	for _, service := range dockerComposeFile.Services {
		// Only if an image is defined in the service
		if len(service.Image) > 0 {
			name, tag, digest := splitDockerDependency(service.Image)
			newDependency := manager.newDependency(name, common.DATASOURCE_TYPE_DOCKER, tag, filePath)
			if digest != "" {
				newDependency.Digest = digest
				setSkipVersionCheckIfVersionMatchesKeyword(newDependency, "latest")
			} else {
				// Without digest and with latest only, we cannot update
				setSkipIfVersionMatchesKeyword(newDependency, "latest")
			}
			foundDependencies = append(foundDependencies, newDependency)
		}
	}

	// Return the found dependencies
	return foundDependencies, nil
}

type dockerComposeFile struct {
	Services map[string]struct {
		Image string `yaml:"image"`
	} `yaml:"services"`
}

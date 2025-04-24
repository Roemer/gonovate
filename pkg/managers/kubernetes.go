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

type KubernetesManager struct {
	*managerBase
}

func NewKubernetesManager(id string, settings *common.ManagerSettings) common.IManager {
	manager := &KubernetesManager{
		managerBase: newManagerBase(id, common.MANAGER_TYPE_KUBERNETES, settings),
	}
	manager.impl = manager
	return manager
}

func (manager *KubernetesManager) ExtractDependencies(filePath string) ([]*common.Dependency, error) {
	// Read the entire file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileContent := string(fileContentBytes)

	// Extract the dependencies from the string
	return manager.extractDependenciesFromString(fileContent, filePath)
}

func (manager *KubernetesManager) ApplyDependencyUpdate(dependency *common.Dependency) error {
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

func (manager *KubernetesManager) extractDependenciesFromString(fileContent string, filePath string) ([]*common.Dependency, error) {
	// A slice to collect all found dependencies
	foundDependencies := []*common.Dependency{}

	// Decode the file
	dec := yaml.NewDecoder(strings.NewReader(fileContent))
	for {
		var yamlObject kubernetesFile
		if err := dec.Decode(&yamlObject); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decoding error: %s", err)
		}
		// Process it
		for _, container := range yamlObject.Spec.Template.Spec.Containers {
			name, tag, digest := splitDockerDependency(container.Image)
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

type kubernetesFile struct {
	Spec struct {
		Template struct {
			Spec struct {
				Containers []struct {
					Image string `yaml:"image"`
				} `yaml:"containers"`
			} `yaml:"spec"`
		} `yaml:"template"`
	} `yaml:"spec"`
}

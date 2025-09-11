package managers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/roemer/goext"
	"github.com/roemer/gonovate/pkg/common"
)

type NpmManager struct {
	*managerBase
}

func NewNpmManager(id string, settings *common.ManagerSettings) common.IManager {
	manager := &NpmManager{
		managerBase: newManagerBase(id, common.MANAGER_TYPE_NPM, settings),
	}
	manager.impl = manager
	return manager
}

func (manager *NpmManager) ExtractDependencies(filePath string) ([]*common.Dependency, error) {
	// Read the entire file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileContent := string(fileContentBytes)

	// Extract the dependencies from the string
	return manager.extractDependenciesFromString(fileContent, filePath)
}

func (manager *NpmManager) ApplyDependencyUpdate(dependency *common.Dependency) error {
	workingDir := filepath.Dir(dependency.FilePath)
	// Prepare the command
	params := []string{"install", "--save-exact"}
	if dependency.Type == npm_dependency_type_dev {
		params = append(params, "--save-dev")
	}
	// Add the package
	params = append(params, fmt.Sprintf("%s@%s", dependency.Name, dependency.NewRelease.VersionString))

	// Execute the command
	outStr, errStr, err := goext.CmdRunners.Default.WithWorkingDirectory(workingDir).RunGetOutput("npm", params...)
	if err != nil {
		return fmt.Errorf("npm command failed: error: %w, stdout: %s, stderr: %s", err, outStr, errStr)
	}

	return nil
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

const (
	npm_dependency_type_direct = "direct"
	npm_dependency_type_dev    = "dev"
)

func (manager *NpmManager) extractDependenciesFromString(fileContent string, filePath string) ([]*common.Dependency, error) {
	// A slice to collect all found dependencies
	foundDependencies := []*common.Dependency{}

	// Decode the file content as JSON
	var jsonData struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal([]byte(fileContent), &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON in file %s: %w", filePath, err)
	}

	// Collect the dependencies
	for dep, version := range jsonData.Dependencies {
		newDependency := manager.newDependency(dep, common.DATASOURCE_TYPE_NPM, version, filePath)
		newDependency.Type = npm_dependency_type_direct
		foundDependencies = append(foundDependencies, newDependency)
	}
	for dep, version := range jsonData.DevDependencies {
		newDependency := manager.newDependency(dep, common.DATASOURCE_TYPE_NPM, version, filePath)
		newDependency.Type = npm_dependency_type_dev
		foundDependencies = append(foundDependencies, newDependency)
	}

	// Return the found dependencies
	return foundDependencies, nil
}

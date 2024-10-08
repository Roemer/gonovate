package managers

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/roemer/gonovate/pkg/common"
)

type GoModManager struct {
	*managerBase
}

func NewGoModManager(id string, settings *common.ManagerSettings) common.IManager {
	manager := &GoModManager{
		managerBase: newManagerBase(id, common.MANAGER_TYPE_GOMOD, settings),
	}
	manager.impl = manager
	return manager
}

func (manager *GoModManager) ExtractDependencies(filePath string) ([]*common.Dependency, error) {
	// Setup
	goVersionRegex := regexp.MustCompile(`^\s*go\s+([^s]+)\s*$`)
	moduleRegex := regexp.MustCompile(`^(?:require)?\s+(?P<module>[^\s]+\/[^\s]+)\s+(?P<version>[^\s]+)(?:\s*\/\/\s*(?P<comment>[^\s]+)\s*)?$`)

	// Read the file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// A slice to collect all found dependencies
	foundDependencies := []*common.Dependency{}
	// Scan the content line by line
	scanner := bufio.NewScanner(bytes.NewReader(fileContentBytes))
	for scanner.Scan() {
		line := scanner.Text()
		// Match the golang version
		if match := goVersionRegex.FindStringSubmatch(line); match != nil {
			newDependency := manager.newDependency("go-stable", common.DATASOURCE_TYPE_GOVERSION, match[1], filePath)
			newDependency.Type = "golang"
			foundDependencies = append(foundDependencies, newDependency)
			continue
		}
		// Match a module
		if match := common.FindNamedMatchesWithIndex(moduleRegex, line, false); match != nil {
			version := match["version"][0].Value
			version = strings.TrimSuffix(version, "+incompatible")

			newDependency := manager.newDependency(match["module"][0].Value, common.DATASOURCE_TYPE_GOMOD, version, filePath)
			newDependency.Type = "direct"
			if v, ok := match["comment"]; ok && v[0].Value == "indirect" {
				newDependency.Type = "indirect"
				// Indirect dependencies are skipped
				continue
			}
			foundDependencies = append(foundDependencies, newDependency)
		}
	}

	// Check if there was an error while scanning
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed processing the line scanner")
	}

	// Return the found dependencies
	return foundDependencies, nil
}

func (manager *GoModManager) ApplyDependencyUpdate(dependency *common.Dependency) error {
	dependencyName := dependency.Name
	if dependency.Type == "golang" {
		dependencyName = "go"
	}
	outStr, errStr, err := common.Execute.RunGetOutput(false, "go", "get", fmt.Sprintf("%s@%s", dependencyName, dependency.NewRelease.VersionString))
	if err != nil {
		return fmt.Errorf("go command failed: error: %w, stdout: %s, stderr: %s", err, outStr, errStr)
	}
	outStr, errStr, err = common.Execute.RunGetOutput(false, "go", "mod", "tidy")
	if err != nil {
		return fmt.Errorf("go command failed: error: %w, stdout: %s, stderr: %s", err, outStr, errStr)
	}
	return nil
}

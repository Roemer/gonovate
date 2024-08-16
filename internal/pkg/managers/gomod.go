package managers

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type GoModManager struct {
	managerBase
}

func NewGoModManager(logger *slog.Logger, id string, rootConfig *config.RootConfig, managerSettings *config.ManagerSettings) IManager {
	manager := &GoModManager{
		managerBase: managerBase{
			logger:     logger.With(slog.String("handlerId", id)),
			id:         id,
			rootConfig: rootConfig,
			settings:   managerSettings,
		},
	}
	manager.impl = manager
	return manager
}

func (manager *GoModManager) ExtractDependencies(filePath string) ([]*shared.Dependency, error) {
	// Setup
	goVersionRegex := regexp.MustCompile(`^\s*go\s+([^s]+)\s*$`)
	moduleRegex := regexp.MustCompile(`^(?:require)?\s+(?P<module>[^\s]+\/[^\s]+)\s+(?P<version>[^\s]+)(?:\s*\/\/\s*(?P<comment>[^\s]+)\s*)?$`)

	// Read the file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// A slice to collect all found dependencies
	foundDependencies := []*shared.Dependency{}
	// Scan the content line by line
	scanner := bufio.NewScanner(bytes.NewReader(fileContentBytes))
	for scanner.Scan() {
		line := scanner.Text()
		// Match the golang version
		if match := goVersionRegex.FindStringSubmatch(line); match != nil {
			newDepencency := &shared.Dependency{
				Name:       "go-stable",
				Datasource: shared.DATASOURCE_TYPE_GOVERSION,
				Type:       "golang",
				Version:    match[1],
			}
			foundDependencies = append(foundDependencies, newDepencency)
			continue
		}
		// Match a module
		if match := shared.FindNamedMatchesWithIndex(moduleRegex, line, false); match != nil {
			newDepencency := &shared.Dependency{
				Name:       match["module"][0].Value,
				Datasource: shared.DATASOURCE_TYPE_GOMOD,
				Type:       "direct",
				Version:    match["version"][0].Value,
			}
			if v, ok := match["comment"]; ok && v[0].Value == "indirect" {
				newDepencency.Type = "indirect"
				// Indirect dependencies are skipped
				continue
			}
			foundDependencies = append(foundDependencies, newDepencency)
		}
	}

	// Check if there was an error while scanning
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed processing the line scanner")
	}

	// Return the found dependencies
	return foundDependencies, nil
}

func (manager *GoModManager) ApplyDependencyUpdate(dependency *shared.Dependency) error {
	dependencyName := dependency.Name
	if dependency.Type == "golang" {
		dependencyName = "go"
	}
	outStr, errStr, err := shared.Execute.RunGetOutput(false, "go", "get", fmt.Sprintf("%s@%s", dependencyName, dependency.NewRelease.VersionString))
	if err != nil {
		return fmt.Errorf("go command failed: error: %w, stdout: %s, stderr: %s", err, outStr, errStr)
	}
	outStr, errStr, err = shared.Execute.RunGetOutput(false, "go", "mod", "tidy")
	if err != nil {
		return fmt.Errorf("go command failed: error: %w, stdout: %s, stderr: %s", err, outStr, errStr)
	}
	return nil
}

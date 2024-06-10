package managers

import (
	"bufio"
	"bytes"
	"fmt"
	"gonovate/core"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
)

type GoModManager struct {
	managerBase2
}

func NewGoModManager(logger *slog.Logger, managerConfig *core.Manager) IManager2 {
	manager := &GoModManager{
		managerBase2: managerBase2{
			logger:        logger.With(slog.String("handlerId", managerConfig.Id)),
			ManagerConfig: managerConfig,
		},
	}
	manager.impl = manager
	return manager
}

func (manager *GoModManager) ExtractDependency(dependencyName string) (*Dependency, error) {
	panic("not implemented") // TODO: Implement
}

func (manager *GoModManager) ExtractDependencies(filePath string) ([]*Dependency, error) {
	// Setup
	goVersionRegex := regexp.MustCompile(`^\s*go\s+([^s]+)\s*$`)
	moduleRegex := regexp.MustCompile(`^(?:require)?\s+(?P<module>[^\s]+\/[^\s]+)\s+(?P<version>[^\s]+)(?:\s*\/\/\s*(?P<comment>[^\s]+)\s*)?$`)

	// Read the file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// A slice to collect all found dependencies
	foundDependencies := []*Dependency{}
	// Scan the content line by line
	scanner := bufio.NewScanner(bytes.NewReader(fileContentBytes))
	for scanner.Scan() {
		line := scanner.Text()
		// Match the golang version
		if match := goVersionRegex.FindStringSubmatch(line); match != nil {
			newDepencency := &Dependency{
				Name:       "go",
				Type:       "golang",
				Version:    match[1],
				Datasource: core.DATASOURCE_TYPE_GOVERSION,
			}
			foundDependencies = append(foundDependencies, newDepencency)
			continue
		}
		// Match a module
		if match := findNamedMatchesWithIndex(moduleRegex, line, false); match != nil {
			newDepencency := &Dependency{
				Name:       match["module"][0].Value,
				Type:       "direct",
				Version:    match["version"][0].Value,
				Datasource: core.DATASOURCE_TYPE_GOMOD,
			}
			if v, ok := match["comment"]; ok && v[0].Value == "indirect" {
				newDepencency.Type = "indirect"
				// Indirect dependencies are also skipped
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

func (manager *GoModManager) ApplyDependencyUpdate(dependencyUpdate *DependencyUpdate) error {
	command := exec.Command("go", "get", fmt.Sprintf("%s@%s", dependencyUpdate.Dependency, dependencyUpdate.NewVersion))
	err := command.Run()
	return err
}

package managers

import (
	"bufio"
	"fmt"
	"gonovate/core"
	"os/exec"
	"regexp"
	"strings"
)

type GoModManager struct {
}

func (manager *GoModManager) ExtractDependency(dependencyName string) (*Dependency, error) {
	panic("not implemented") // TODO: Implement
}

func (manager *GoModManager) ExtractDependencies(content string) ([]*Dependency, error) {
	// Setup
	goVersionRegex := regexp.MustCompile(`^\s*go\s+([^s]+)\s*$`)
	moduleRegex := regexp.MustCompile(`^(?:require)?\s+(?P<module>[^\s]+\/[^\s]+)\s+(?P<version>[^\s]+)(?:\s*\/\/\s*(?P<comment>[^\s]+)\s*)?$`)

	foundDependencies := []*Dependency{}
	// Scan the content line by line
	scanner := bufio.NewScanner(strings.NewReader(content))
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

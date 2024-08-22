package managers

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type DockerfileManager struct {
	managerBase
}

func NewDockerfileManager(logger *slog.Logger, id string, rootConfig *config.RootConfig, managerSettings *config.ManagerSettings) IManager {
	manager := &DockerfileManager{
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

func (manager *DockerfileManager) ExtractDependencies(filePath string) ([]*shared.Dependency, error) {
	// Setup
	dockerFromRegex := regexp.MustCompile(`^FROM(?:\s+--platform=.*?)?\s+(.+?)(?:\s+(?:as|AS)\s+.+)?\s*$`)

	// Read the file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// A slice to collect all found dependencies
	foundDependencies := []*shared.Dependency{}
	// Scan the content line by line
	scanner := bufio.NewScanner(bytes.NewReader(fileContentBytes))
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		// Search for the marker
		if match := dockerFromRegex.FindStringSubmatch(line); match != nil {
			name, version := splitDockerDependency(match[1])
			newDepencency := &shared.Dependency{
				Name:        name,
				Datasource:  shared.DATASOURCE_TYPE_DOCKER,
				Version:     version,
				ManagerData: &dockerfileData{lineNumber: lineCount},
			}
			disableIfVersionMatches(newDepencency, "latest")
			foundDependencies = append(foundDependencies, newDepencency)
			break
		}
		lineCount++
	}

	// Check if there was an error while scanning
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed processing the line scanner")
	}

	// Return the found dependencies
	return foundDependencies, nil
}

func (manager *DockerfileManager) ApplyDependencyUpdate(dependency *shared.Dependency) error {
	data := dependency.ManagerData.(*dockerfileData)
	return replaceVersionInFileLine(dependency.FilePath, dependency.Version, dependency.NewRelease.VersionString, data.lineNumber)
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

type dockerfileData struct {
	lineNumber int
}

func replaceVersionInFileLine(filePath string, oldVersion string, newVersion string, line int) error {
	// Read the file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	fileContent := string(fileContentBytes)
	lines := strings.Split(fileContent, "\n")
	if line >= len(lines) {
		return fmt.Errorf("the file does not have enough lines")
	}
	lines[line] = strings.Replace(lines[line], oldVersion, newVersion, 1)

	// Write the file with the changes
	if err := os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), os.ModePerm); err != nil {
		return err
	}
	return nil
}

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

type DockerfileManager struct {
	*managerBase
}

func NewDockerfileManager(settings *common.ManagerSettings) common.IManager {
	manager := &DockerfileManager{
		managerBase: newManagerBase(settings),
	}
	manager.impl = manager
	return manager
}

func (manager *DockerfileManager) ExtractDependencies(filePath string) ([]*common.Dependency, error) {
	// Setup
	dockerFromRegex := regexp.MustCompile(`^FROM(?:\s+--platform=(.*)?)?\s+(.+?)(?:\s+(?:as|AS)\s+.+)?\s*$`)

	// Read the file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// A slice to collect all found dependencies
	foundDependencies := []*common.Dependency{}
	// Scan the content line by line
	scanner := bufio.NewScanner(bytes.NewReader(fileContentBytes))
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		// Search for the marker
		if match := dockerFromRegex.FindStringSubmatch(line); match != nil {
			platform := match[1]
			image := match[2]
			name, tag, digest := splitDockerDependency(image)
			newDependency := manager.newDependency(name, common.DATASOURCE_TYPE_DOCKER, tag, filePath)
			newDependency.ManagerInfo.ManagerData = &dockerfileData{lineNumber: lineCount}
			if platform != "" {
				newDependency.AdditionalData["platform"] = platform
			}
			if digest != "" {
				newDependency.Digest = digest
				setSkipVersionCheckIfVersionMatchesKeyword(newDependency, "latest")
			} else {
				// Without digest and with latest only, we cannot update
				setSkipIfVersionMatchesKeyword(newDependency, "latest")
			}
			foundDependencies = append(foundDependencies, newDependency)
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

func (manager *DockerfileManager) ApplyDependencyUpdate(dependency *common.Dependency) error {
	data := dependency.ManagerInfo.ManagerData.(*dockerfileData)
	oldFullVersion, newFullVersion := getDockerCurrentAndNewFullVersion(dependency)
	return replaceVersionInFileLine(dependency.FilePath, oldFullVersion, newFullVersion, data.lineNumber)
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

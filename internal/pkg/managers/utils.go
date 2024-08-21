package managers

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/roemer/gonovate/internal/pkg/shared"
	"github.com/samber/lo"
)

// Splits a Docker image into separate name and tag. Uses "latest" if no tag is present.
func splitDockerDependency(dependencyString string) (string, string) {
	parts := strings.SplitN(dependencyString, ":", 2)
	name := parts[0]
	version := lo.Ternary(len(parts) > 1, parts[1], "latest")
	return name, version
}

func replaceDependencyVersionInFileWithCheck(dependency *shared.Dependency, refetchDependencyFunc func(dependency *shared.Dependency, newFileContent string) (*shared.Dependency, error)) error {
	// Read the file
	fileContentBytes, err := os.ReadFile(dependency.FilePath)
	if err != nil {
		return err
	}
	fileContent := string(fileContentBytes)

	// Search for all places where the exact version string is present
	regVersion := regexp.MustCompile(regexp.QuoteMeta(dependency.Version))
	matches := regVersion.FindAllStringIndex(fileContent, -1)
	dependencyUpdated := false
	// Loop thru all the matches, replace one after another and re-check if the correct dependency iks updated
	for _, match := range matches {
		matchStart := match[0]
		matchEnd := match[1]
		// Create a new content with the replaced version
		tempContent := fileContent[:matchStart] + dependency.NewRelease.VersionString + fileContent[matchEnd:]

		// Check if the dependency is now updated
		newDependency, err := refetchDependencyFunc(dependency, tempContent)
		if err != nil {
			return err
		}
		if newDependency.Version == dependency.NewRelease.VersionString {
			// If so, set the new content and break out of the loop
			fileContent = tempContent
			dependencyUpdated = true
			break
		}
		// Otherwise continue with the loop and try the next match
	}

	// Throw an error if the dependency could not be updated
	if !dependencyUpdated {
		return fmt.Errorf("failed to update dependency: %s", dependency.Name)
	}

	// Write the file with the changes
	if err := os.WriteFile(dependency.FilePath, []byte(fileContent), os.ModePerm); err != nil {
		return err
	}
	return nil
}

package managers

import (
	"fmt"
	"os"
	"regexp"
	"slices"

	"github.com/roemer/gonovate/internal/pkg/shared"
)

// Splits a Docker image into separate name, tag and digest. Uses "latest" if no tag is present.
func splitDockerDependency(dependencyString string) (string, string, string) {
	dockerImageRegex := regexp.MustCompile("^(?P<image>[^:@]+?)(?::(?P<tag>[^@]+?))?(?:@(?P<digest>sha256:.+))?$")
	match := dockerImageRegex.FindStringSubmatch(dependencyString)
	if match == nil {
		return dependencyString, "", ""
	}

	image := match[1]
	tag := match[2]
	digest := match[3]

	if tag == "" && digest == "" {
		tag = "latest"
	}
	return image, tag, digest
}

// This method returns the full Docker version, including a digest if any is given
func getDockerCurrentAndNewFullVersion(dependency *shared.Dependency) (string, string) {
	oldVersion := dependency.Version
	newVersion := dependency.NewRelease.VersionString
	// Enrich with digest (if any)
	if dependency.HasDigest() {
		oldVersion += "@" + dependency.Digest
		newVersion += "@" + dependency.NewRelease.Digest
	}
	return oldVersion, newVersion
}

// Skips the dependency if the version matches one of the given keywords.
func skipIfVersionMatches(dependency *shared.Dependency, skipValues ...string) {
	if dependency.Skip == nil || !*dependency.Skip {
		if slices.Contains(skipValues, dependency.Version) {
			dependency.Skip = shared.TruePtr
			dependency.SkipReason = fmt.Sprintf("Version is set to '%s'", dependency.Version)
		}
	}
}

// Skips the version check for the given dependency version machtes one of the keywords.
func skipVersionCheckIfVersionMatches(dependency *shared.Dependency, skipValues ...string) {
	if dependency.SkipVersionCheck == nil || !*dependency.SkipVersionCheck {
		if slices.Contains(skipValues, dependency.Version) {
			dependency.SkipVersionCheck = shared.TruePtr
		}
	}
}

func replaceDependencyVersionInFileWithCheck(dependency *shared.Dependency, refetchDependencyFunc func(dependency *shared.Dependency, newFileContent string) (*shared.Dependency, error)) error {
	// Read the file
	fileContentBytes, err := os.ReadFile(dependency.FilePath)
	if err != nil {
		return err
	}
	fileContent := string(fileContentBytes)

	// Set the strings to search and replace
	oldString := dependency.Version
	newString := dependency.NewRelease.VersionString

	// Handle docker dependency which can have a digest
	if dependency.Datasource == shared.DATASOURCE_TYPE_DOCKER {
		oldFullVersion, newFullVersion := getDockerCurrentAndNewFullVersion(dependency)
		oldString = oldFullVersion
		newString = newFullVersion
	}

	// Search for all places where the exact version string is present
	regVersion := regexp.MustCompile(regexp.QuoteMeta(oldString))
	matches := regVersion.FindAllStringIndex(fileContent, -1)
	dependencyUpdated := false
	// Loop thru all the matches, replace one after another and re-check if the correct dependency iks updated
	for _, match := range matches {
		matchStart := match[0]
		matchEnd := match[1]
		// Create a new content with the replaced version
		tempContent := fileContent[:matchStart] + newString + fileContent[matchEnd:]

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

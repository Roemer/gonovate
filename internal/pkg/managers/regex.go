package managers

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type RegexManager struct {
	managerBase
}

func NewRegexManager(logger *slog.Logger, config *config.RootConfig, managerConfig *config.ManagerConfig) IManager {
	manager := &RegexManager{
		managerBase: managerBase{
			logger:        logger.With(slog.String("handlerId", managerConfig.Id)),
			Config:        config,
			ManagerConfig: managerConfig,
		},
	}
	manager.impl = manager
	return manager
}

func (manager *RegexManager) ExtractDependencies(filePath string) ([]*shared.Dependency, error) {
	// Read the entire file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileContent := string(fileContentBytes)

	// Extract the dependencies from the string
	return manager.extractDependenciesFromString(fileContent)
}

func (manager *RegexManager) ApplyDependencyUpdate(dependency *shared.Dependency) error {
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
		newDeps, err := manager.extractDependenciesFromString(tempContent)
		if err != nil {
			return err
		}
		newDep, err := manager.getSingleDependency(dependency.Name, newDeps)
		if err != nil {
			return err
		}
		if newDep.Version == dependency.NewRelease.VersionString {
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

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

// Extract dependencies from a given string by searching for the regexes.
func (manager *RegexManager) extractDependenciesFromString(fileContent string) ([]*shared.Dependency, error) {
	// Precompile the regexes from the manager settings
	precompiledRegexList := []*regexp.Regexp{}
	for _, regStr := range manager.ManagerConfig.ManagerSettings.MatchStrings {
		resolvedMatchString, err := manager.Config.ResolveMatchString(regStr)
		if err != nil {
			return nil, err
		}
		regex, err := regexp.Compile(resolvedMatchString)
		if err != nil {
			return nil, err
		}
		precompiledRegexList = append(precompiledRegexList, regex)
	}
	manager.logger.Debug(fmt.Sprintf("Found %d match pattern(s) to process", len(precompiledRegexList)))

	// Prepare a slice to collect all found dependencies
	foundDependencies := []*shared.Dependency{}

	// Loop thru all regex patterns
	for _, regex := range precompiledRegexList {
		matchList := shared.FindAllNamedMatchesWithIndex(regex, fileContent, false, -1)
		if matchList == nil {
			// The regex was not matched, go to the next
			continue
		}
		for _, match := range matchList {
			// The version must be found with the regexp on the line
			versionObject, versionOk := match["version"]
			if !versionOk {
				// The version field is mandatory
				return nil, fmt.Errorf("the field 'version' did not match")
			}
			//  Optional fields
			datasourceObject, datasourceOk := match["datasource"]
			dependencyObject, dependencyOk := match["dependencyName"]
			versioningObject, versioningOk := match["versioning"]
			maxUpdateTypeObject, maxUpdateTypeOk := match["maxUpdateType"]
			extractVersionObject, extractVersionOk := match["extractVersion"]

			// Build the dependency object
			newDepencency := &shared.Dependency{
				Version: versionObject[0].Value,
			}
			if datasourceOk {
				newDepencency.Datasource = shared.DatasourceType(datasourceObject[0].Value)
			}
			if dependencyOk {
				newDepencency.Name = dependencyObject[0].Value
			}
			if versioningOk {
				newDepencency.Versioning = versioningObject[0].Value
			}
			if maxUpdateTypeOk {
				newDepencency.MaxUpdateType = shared.UpdateType(maxUpdateTypeObject[0].Value)
			}
			if extractVersionOk {
				newDepencency.ExtractVersion = extractVersionObject[0].Value
			}

			// Add the dependency
			foundDependencies = append(foundDependencies, newDepencency)
		}
	}

	// Return the found dependencies
	return foundDependencies, nil
}

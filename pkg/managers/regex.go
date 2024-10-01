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

func NewRegexManager(logger *slog.Logger, id string, rootConfig *config.RootConfig, managerSettings *config.ManagerSettings) IManager {
	manager := &RegexManager{
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
	return replaceDependencyVersionInFileWithCheck(dependency, func(dependency *shared.Dependency, newFileContent string) (*shared.Dependency, error) {
		newDeps, err := manager.extractDependenciesFromString(newFileContent)
		if err != nil {
			return nil, err
		}
		newDep, err := manager.getSingleDependency(dependency.Name, newDeps)
		if err != nil {
			return nil, err
		}
		return newDep, nil
	})
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

// Extract dependencies from a given string by searching for the regexes.
func (manager *RegexManager) extractDependenciesFromString(fileContent string) ([]*shared.Dependency, error) {
	// Precompile the regexes from the manager settings
	precompiledRegexList := []*regexp.Regexp{}
	for _, regStr := range manager.settings.MatchStrings {
		resolvedMatchString, err := manager.rootConfig.ResolveMatchString(regStr)
		if err != nil {
			return nil, err
		}
		regex, err := regexp.Compile(resolvedMatchString)
		if err != nil {
			return nil, err
		}
		precompiledRegexList = append(precompiledRegexList, regex)
	}
	manager.logger.Debug(fmt.Sprintf("Found %s to process", shared.GetSingularPluralStringSimple(precompiledRegexList, "match pattern")))

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

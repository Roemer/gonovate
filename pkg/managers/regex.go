package managers

import (
	"fmt"
	"os"
	"regexp"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/roemer/gonovate/pkg/presets"
)

type RegexManager struct {
	*managerBase
}

func NewRegexManager(id string, settings *common.ManagerSettings) common.IManager {
	manager := &RegexManager{
		managerBase: newManagerBase(id, common.MANAGER_TYPE_INLINE, settings),
	}
	manager.impl = manager
	return manager
}

func (manager *RegexManager) ExtractDependencies(filePath string) ([]*common.Dependency, error) {
	// Read the entire file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileContent := string(fileContentBytes)

	// Extract the dependencies from the string
	return manager.extractDependenciesFromString(fileContent, filePath)
}

func (manager *RegexManager) ApplyDependencyUpdate(dependency *common.Dependency) error {
	return replaceDependencyVersionInFileWithCheck(dependency, func(dependency *common.Dependency, newFileContent string) (*common.Dependency, error) {
		newDeps, err := manager.extractDependenciesFromString(newFileContent, dependency.FilePath)
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
func (manager *RegexManager) extractDependenciesFromString(fileContent string, filePath string) ([]*common.Dependency, error) {
	// Precompile the regexes from the manager settings
	precompiledRegexList := []*regexp.Regexp{}
	for _, regStr := range manager.settings.RegexManagerSettings.MatchStrings {
		resolvedMatchString, err := presets.ResolveMatchString(regStr, manager.settings.RegexManagerSettings.MatchStringPresets)
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
	foundDependencies := []*common.Dependency{}

	// Loop thru all regex patterns
	for _, regex := range precompiledRegexList {
		matchList := common.FindAllNamedMatchesWithIndex(regex, fileContent, false, -1)
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
			// Optional fields
			datasourceObject, datasourceOk := match["datasource"]
			dependencyObject, dependencyOk := match["dependencyName"]
			versioningObject, versioningOk := match["versioning"]
			maxUpdateTypeObject, maxUpdateTypeOk := match["maxUpdateType"]
			extractVersionObject, extractVersionOk := match["extractVersion"]

			// Build the dependency object
			newDependency := manager.newDependency("", "", versionObject[0].Value, filePath)
			if datasourceOk {
				newDependency.Datasource = common.DatasourceType(datasourceObject[0].Value)
			}
			if dependencyOk {
				newDependency.Name = dependencyObject[0].Value
			}
			if versioningOk {
				newDependency.Versioning = versioningObject[0].Value
			}
			if maxUpdateTypeOk {
				newDependency.MaxUpdateType = common.UpdateType(maxUpdateTypeObject[0].Value)
			}
			if extractVersionOk {
				newDependency.ExtractVersion = extractVersionObject[0].Value
			}

			// Add the dependency
			foundDependencies = append(foundDependencies, newDependency)
		}
	}

	// Return the found dependencies
	return foundDependencies, nil
}

package managers

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/roemer/gonovate/pkg/presets"
)

type InlineManager struct {
	*managerBase
}

func NewInlineManager(settings *common.ManagerSettings) common.IManager {
	manager := &InlineManager{
		managerBase: newManagerBase(settings),
	}
	manager.impl = manager
	return manager
}

func (manager *InlineManager) ExtractDependencies(filePath string) ([]*common.Dependency, error) {
	// Read the entire file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileContent := string(fileContentBytes)

	// Extract the dependencies from the string
	return manager.extractDependenciesFromString(fileContent, filePath)
}

func (manager *InlineManager) ApplyDependencyUpdate(dependency *common.Dependency) error {
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

// Extract dependencies from a given string by searching for markers
func (manager *InlineManager) extractDependenciesFromString(fileContent string, filePath string) ([]*common.Dependency, error) {
	// Prepare the marker regex which searches the file for the inline markers
	markerRegex := regexp.MustCompile("(?m)^[[:blank:]]*[/#*`'<!-{]+ gonovate: (.+)\\s*$")

	// Prepare a slice to collect all found dependencies
	foundDependencies := []*common.Dependency{}

	// Search for the markers
	markerMatches := markerRegex.FindAllStringSubmatchIndex(fileContent, -1)
	indexOffset := 0
	for _, markerMatch := range markerMatches {
		markerStart := markerMatch[2] + indexOffset
		markerEnd := markerMatch[3] + indexOffset
		configStr := fileContent[markerStart:markerEnd]

		// Get the config for the marker
		inlineConfig := &inlineManagerConfig{}
		if err := json.Unmarshal([]byte(configStr), inlineConfig); err != nil {
			return nil, fmt.Errorf("failed parsing marker config at position %d: %w", markerStart, err)
		}

		// Build the regex that was defined in the marker
		resolvedMatchString, err := presets.ResolveMatchString(inlineConfig.MatchString, manager.settings.RegexManagerSettings.MatchStringPresets)
		if err != nil {
			return nil, err
		}
		newReg := regexp.MustCompile(resolvedMatchString)
		// Search the remaining file content with this new regex and process the first match only
		contentSearchStart := markerEnd + 1
		matchList := common.FindAllNamedMatchesWithIndex(newReg, fileContent[contentSearchStart:], false, 1)
		if matchList == nil || len(matchList) < 1 {
			return nil, fmt.Errorf("regex defined in marker at position %d did not match anything", markerStart)
		}
		// We are only interested in the first match
		match := matchList[0]
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
		newDependency := manager.newDependency("", "", versionObject[0].Value, filePath)
		if datasourceOk {
			newDependency.Datasource = common.DatasourceType(datasourceObject[0].Value)
		} else if inlineConfig.Datasource != "" {
			newDependency.Datasource = inlineConfig.Datasource
		}
		if dependencyOk {
			newDependency.Name = dependencyObject[0].Value
		} else if inlineConfig.DependencyName != "" {
			newDependency.Name = inlineConfig.DependencyName
		}
		if versioningOk {
			newDependency.Versioning = versioningObject[0].Value
		} else if inlineConfig.Versioning != "" {
			newDependency.Versioning = inlineConfig.Versioning
		}
		if maxUpdateTypeOk {
			newDependency.MaxUpdateType = common.UpdateType(maxUpdateTypeObject[0].Value)
		} else if inlineConfig.MaxUpdateType != "" {
			newDependency.MaxUpdateType = inlineConfig.MaxUpdateType
		}
		if extractVersionOk {
			newDependency.ExtractVersion = extractVersionObject[0].Value
		} else if inlineConfig.ExtractVersion != "" {
			newDependency.ExtractVersion = inlineConfig.ExtractVersion
		}

		// Add the dependency
		foundDependencies = append(foundDependencies, newDependency)
	}

	// Return the found dependencies
	return foundDependencies, nil
}

type inlineManagerConfig struct {
	DependencyName string                `json:"dependencyName"`
	Datasource     common.DatasourceType `json:"datasource"`
	MatchString    string                `json:"matchString"`
	Versioning     string                `json:"versioning"`
	MaxUpdateType  common.UpdateType     `json:"maxUpdateType"`
	ExtractVersion string                `json:"extractVersion"`
}

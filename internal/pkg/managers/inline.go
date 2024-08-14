package managers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type InlineManager struct {
	managerBase
}

func NewInlineManager(logger *slog.Logger, id string, rootConfig *config.RootConfig, managerSettings *config.ManagerSettings) IManager {
	manager := &InlineManager{
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

func (manager *InlineManager) ExtractDependencies(filePath string) ([]*shared.Dependency, error) {
	// Read the entire file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileContent := string(fileContentBytes)

	// Extract the dependencies from the string
	return manager.extractDependenciesFromString(fileContent)
}

func (manager *InlineManager) ApplyDependencyUpdate(dependency *shared.Dependency) error {
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

// Extract dependencies from a given string by searching for markers
func (manager *InlineManager) extractDependenciesFromString(fileContent string) ([]*shared.Dependency, error) {
	// Prepare the marker regex which searches the file for the inline markers
	markerRegex := regexp.MustCompile("(?m)^[[:blank:]]*[/#*`'<!-{]+ gonovate: (.+)\\s*$")

	// Prepare a slice to collect all found dependencies
	foundDependencies := []*shared.Dependency{}

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
		resolvedMatchString, err := manager.rootConfig.ResolveMatchString(inlineConfig.MatchString)
		if err != nil {
			return nil, err
		}
		newReg := regexp.MustCompile(resolvedMatchString)
		// Search the remaining file content with this new regex and process the first match only
		contentSearchStart := markerEnd + 1
		matchList := shared.FindAllNamedMatchesWithIndex(newReg, fileContent[contentSearchStart:], false, 1)
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
		newDepencency := &shared.Dependency{
			Version: versionObject[0].Value,
		}
		if datasourceOk {
			newDepencency.Datasource = shared.DatasourceType(datasourceObject[0].Value)
		} else if inlineConfig.Datasource != "" {
			newDepencency.Datasource = inlineConfig.Datasource
		}
		if dependencyOk {
			newDepencency.Name = dependencyObject[0].Value
		} else if inlineConfig.DependencyName != "" {
			newDepencency.Name = inlineConfig.DependencyName
		}
		if versioningOk {
			newDepencency.Versioning = versioningObject[0].Value
		} else if inlineConfig.Versioning != "" {
			newDepencency.Versioning = inlineConfig.Versioning
		}
		if maxUpdateTypeOk {
			newDepencency.MaxUpdateType = shared.UpdateType(maxUpdateTypeObject[0].Value)
		} else if inlineConfig.MaxUpdateType != "" {
			newDepencency.MaxUpdateType = inlineConfig.MaxUpdateType
		}
		if extractVersionOk {
			newDepencency.ExtractVersion = extractVersionObject[0].Value
		} else if inlineConfig.ExtractVersion != "" {
			newDepencency.ExtractVersion = inlineConfig.ExtractVersion
		}

		// Add the dependency
		foundDependencies = append(foundDependencies, newDepencency)
	}

	// Return the found dependencies
	return foundDependencies, nil
}

type inlineManagerConfig struct {
	DependencyName string                `json:"dependencyName"`
	Datasource     shared.DatasourceType `json:"datasource"`
	MatchString    string                `json:"matchString"`
	Versioning     string                `json:"versioning"`
	MaxUpdateType  shared.UpdateType     `json:"maxUpdateType"`
	ExtractVersion string                `json:"extractVersion"`
}
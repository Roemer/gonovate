package managers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/roemer/gonovate/internal/config"
	"github.com/roemer/gonovate/internal/core"
)

type Inline2Manager struct {
	managerBase2
}

func NewInline2Manager(logger *slog.Logger, config *config.Config, managerConfig *config.Manager) IManager2 {
	manager := &Inline2Manager{
		managerBase2: managerBase2{
			logger:        logger.With(slog.String("handlerId", managerConfig.Id)),
			Config:        config,
			ManagerConfig: managerConfig,
		},
	}
	manager.impl = manager
	return manager
}

func (manager *Inline2Manager) ExtractDependency(dependencyName string) (*core.Dependency, error) {
	panic("not implemented") // TODO: Implement
}

func (manager *Inline2Manager) ExtractDependencies(filePath string) ([]*core.Dependency, error) {
	// Prepare the marker regex which searches the file for the inline markers
	markerRegex := regexp.MustCompile("(?m)^[[:blank:]]*[/#*`]+ gonovate: (.+)\\s*$")

	// Read the entire file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileContent := string(fileContentBytes)

	// A slice to collect all found dependencies
	foundDependencies := []*core.Dependency{}

	// Search for the markers
	matchesIndex := markerRegex.FindAllStringSubmatchIndex(fileContent, -1)
	indexOffset := 0
	for _, match := range matchesIndex {
		start := match[2] + indexOffset
		end := match[3] + indexOffset
		configStr := fileContent[start:end]

		// Get the config for the marker
		inlineConfig := &inlineManagerConfig{}
		if err = json.Unmarshal([]byte(configStr), inlineConfig); err != nil {
			return nil, fmt.Errorf("failed parsing marker config at position %d: %w", start, err)
		}

		// Build the regex that was defined in the marker
		resolvedMatchString, err := manager.Config.ResolveMatchString(inlineConfig.MatchString)
		if err != nil {
			return nil, err
		}
		newReg := regexp.MustCompile(resolvedMatchString)
		// Search the remaining file content with this new regex and process the first match only
		contentSearchStart := end + 1
		matchList := findAllNamedMatchesWithIndex(newReg, fileContent[contentSearchStart:], false, 1)
		if matchList == nil || len(matchList) < 1 {
			return nil, fmt.Errorf("regex defined in marker at position %d did not match anything", start)
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
		packageObject, packageOk := match["packageName"]
		versioningObject, versioningOk := match["versioning"]
		maxUpdateTypeObject, maxUpdateTypeOk := match["maxUpdateType"]
		extractVersionObject, extractVersionOk := match["extractVersion"]

		// Build the dependency object
		newDepencency := &core.Dependency{
			Version: versionObject[0].Value,
		}
		if datasourceOk {
			newDepencency.Datasource = core.DatasourceType(datasourceObject[0].Value)
		} else if inlineConfig.Datasource != "" {
			newDepencency.Datasource = inlineConfig.Datasource
		}
		if packageOk {
			newDepencency.Name = packageObject[0].Value
		} else if inlineConfig.PackageName != "" {
			newDepencency.Name = inlineConfig.PackageName
		}
		if versioningOk {
			newDepencency.Versioning = versioningObject[0].Value
		} else if inlineConfig.Versioning != "" {
			newDepencency.Versioning = inlineConfig.Versioning
		}
		if maxUpdateTypeOk {
			newDepencency.MaxUpdateType = core.UpdateType(maxUpdateTypeObject[0].Value)
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

func (manager *Inline2Manager) ApplyDependencyUpdate(dependencyUpdate *core.DependencyUpdate) error {
	panic("not implemented") // TODO: Implement
}

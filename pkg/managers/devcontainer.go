package managers

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/adhocore/jsonc"
	"github.com/roemer/gonovate/pkg/common"
)

type DevcontainerManager struct {
	*managerBase
}

func NewDevcontainerManager(id string, settings *common.ManagerSettings) common.IManager {
	manager := &DevcontainerManager{
		managerBase: newManagerBase(id, common.MANAGER_TYPE_DEVCONTAINER, settings),
	}
	manager.impl = manager
	return manager
}

func (manager *DevcontainerManager) ExtractDependencies(filePath string) ([]*common.Dependency, error) {
	// Read the entire file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	// Extract the dependencies from the string
	return manager.extractDependenciesFromString(string(fileContentBytes), filePath)
}

func (manager *DevcontainerManager) ApplyDependencyUpdate(dependency *common.Dependency) error {
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

func (manager *DevcontainerManager) extractDependenciesFromString(jsonContent string, filePath string) ([]*common.Dependency, error) {
	// devcontainer.json files are jsonc/json5, so convert it to json first
	j := jsonc.New()
	strippedJson := j.StripS(jsonContent)

	// Parse the file
	inlineConfig := &devcontainerData{}
	if err := json.Unmarshal([]byte(strippedJson), inlineConfig); err != nil {
		return nil, fmt.Errorf("failed parsing devcontainer config: %w", err)
	}

	// A slice to collect all found dependencies
	foundDependencies := []*common.Dependency{}

	// Search for an image dependency
	if inlineConfig.Image != "" {
		name, tag, digest := splitDockerDependency(inlineConfig.Image)
		newDependency := manager.newDependency(name, common.DATASOURCE_TYPE_DOCKER, tag, filePath)
		newDependency.Type = "image"
		if digest != "" {
			newDependency.Digest = digest
			setSkipVersionCheckIfVersionMatchesKeyword(newDependency, "latest")
		} else {
			// Without digest and with latest only, we cannot update
			setSkipIfVersionMatchesKeyword(newDependency, "latest")
		}
		foundDependencies = append(foundDependencies, newDependency)
	}

	// Search for features and dependencies inside features
	for feature, dependenciesInsideFeature := range inlineConfig.Features {
		// Create the feature dependency
		name, tag, digest := splitDockerDependency(feature)
		featureDependency := manager.newDependency(name, common.DATASOURCE_TYPE_DOCKER, tag, filePath)
		featureDependency.Type = "feature"
		// Features currently cannot contain digests, but might in the future
		if digest != "" {
			featureDependency.Digest = digest
			setSkipVersionCheckIfVersionMatchesKeyword(featureDependency, "latest")
		} else {
			// Without digest and with latest only, we cannot update
			setSkipIfVersionMatchesKeyword(featureDependency, "latest")
		}
		foundDependencies = append(foundDependencies, featureDependency)

		if len(dependenciesInsideFeature) == 0 {
			// This feature has no dependencies inside, so skip further processing
			continue
		}

		// Search for the feature and property in the settings
		if manager.settings.DevcontainerManagerSettings == nil {
			manager.logger.Debug("Manager has no settings for features, skipping dependencies")
			continue
		}
		featureSettings, ok := manager.settings.DevcontainerManagerSettings.FeatureDependencies[name]
		if !ok {
			manager.logger.Debug(fmt.Sprintf("Feature '%s' has no settings, skipping dependencies", name))
			continue
		}

		// Process the dependencies inside the feature
		for property, propertyValue := range dependenciesInsideFeature {
			// Filter out properties with values that are not strings
			if propertyString, ok := propertyValue.(string); !ok {
				continue
			} else {
				// Search for a setting with the same property name
				idx := slices.IndexFunc(featureSettings, func(d *common.DevcontainerManagerFeatureDependency) bool { return d.Property == property })
				if idx < 0 {
					manager.logger.Debug(fmt.Sprintf("Dependency in feature for property '%s' has no settings", property))
					continue
				}
				featureDependency := featureSettings[idx]
				newDependencyInsideFeature := manager.newDependency(featureDependency.DependencyName, featureDependency.Datasource, propertyString, filePath)
				newDependencyInsideFeature.Type = "dependency"
				setSkipIfVersionMatchesKeyword(newDependencyInsideFeature, "latest", "none")
				foundDependencies = append(foundDependencies, newDependencyInsideFeature)
			}
		}
	}

	return foundDependencies, nil
}

type devcontainerData struct {
	Image    string                            `json:"image"`
	Features map[string]map[string]interface{} `json:"features"`
}

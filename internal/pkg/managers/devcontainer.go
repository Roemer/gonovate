package managers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/adhocore/jsonc"
	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

type DevcontainerManager struct {
	managerBase
}

func NewDevcontainerManager(logger *slog.Logger, id string, rootConfig *config.RootConfig, managerSettings *config.ManagerSettings) IManager {
	manager := &DevcontainerManager{
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

func (manager *DevcontainerManager) ExtractDependencies(filePath string) ([]*shared.Dependency, error) {
	// Read the entire file
	fileContentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	// Extract the dependencies from the string
	return manager.extractDependenciesFromString(string(fileContentBytes))
}

func (manager *DevcontainerManager) ApplyDependencyUpdate(dependency *shared.Dependency) error {
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

func (manager *DevcontainerManager) extractDependenciesFromString(jsonContent string) ([]*shared.Dependency, error) {
	// devcontainer.json files are jsonc/json5, so convert it to json first
	j := jsonc.New()
	strippedJson := j.StripS(jsonContent)

	// Parse the file
	inlineConfig := &devcontainerData{}
	if err := json.Unmarshal([]byte(strippedJson), inlineConfig); err != nil {
		return nil, fmt.Errorf("failed parsing devcontainer config: %w", err)
	}

	// A slice to collect all found dependencies
	foundDependencies := []*shared.Dependency{}

	// Search for an image dependency
	if inlineConfig.Image != "" {
		name, version := splitDockerDependency(inlineConfig.Image)
		newDepencency := &shared.Dependency{
			Name:       name,
			Datasource: shared.DATASOURCE_TYPE_DOCKER,
			Version:    version,
			Type:       "image",
		}
		disableDockerIfLatest(newDepencency)
		foundDependencies = append(foundDependencies, newDepencency)
	}

	// Search for features and dependencies inside features
	for feature, dependenciesInsideFeature := range inlineConfig.Features {
		// Create the feature dependency
		name, version := splitDockerDependency(feature)
		featureDependency := &shared.Dependency{
			Name:       name,
			Datasource: shared.DATASOURCE_TYPE_DOCKER,
			Version:    version,
			Type:       "feature",
		}
		disableDockerIfLatest(featureDependency)
		foundDependencies = append(foundDependencies, featureDependency)

		if len(dependenciesInsideFeature) == 0 {
			// This feature has no dependencies inside, so skip further processing
			continue
		}

		// Search for the feature and property in the settings
		featureSettings, ok := manager.settings.DevcontainerSettings[name]
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
				idx := slices.IndexFunc(featureSettings, func(d *config.DevcontainerFeatureDependency) bool { return d.Property == property })
				if idx < 0 {
					manager.logger.Debug(fmt.Sprintf("Dependency in feature for property '%s' has no settings", property))
					continue
				}
				featureDependency := featureSettings[idx]
				newDependencyInsideFeature := &shared.Dependency{
					Name:       featureDependency.DependencyName,
					Datasource: featureDependency.Datasource,
					Version:    propertyString,
				}
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

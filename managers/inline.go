package managers

import (
	"encoding/json"
	"fmt"
	"gonovate/core"
	"gonovate/platforms"
	"log/slog"
	"os"
	"regexp"
)

type InlineManager struct {
	managerBase
}

func NewInlineManager(logger *slog.Logger, globalConfig *core.Config, managerConfig *core.Manager) IManager {
	manager := &InlineManager{
		managerBase: managerBase{
			logger:       logger.With(slog.String("handlerId", managerConfig.Id)),
			GlobalConfig: globalConfig,
			Config:       managerConfig,
		},
	}
	manager.impl = manager
	return manager
}

func (manager *InlineManager) process(platform platforms.IPlatform) error {
	manager.logger.Info(fmt.Sprintf("Starting InlineManager with Id %s", manager.Config.Id))

	// Process all rules to apply the ones relevant for the manager and store the ones relevant for packages.
	managerSettings, possiblePackageRules := manager.GlobalConfig.FilterForManager(manager.Config)

	// Skip if it is disabled
	if managerSettings.Disabled != nil && *managerSettings.Disabled {
		manager.logger.Info(fmt.Sprintf("Skipping Manager '%s' (%s) as it is disabled", manager.Config.Id, manager.Config.Type))
		return nil
	}

	// Search file candidates
	manager.logger.Debug(fmt.Sprintf("Searching files with %d pattern(s)", len(managerSettings.FilePatterns)))
	candidates, err := core.SearchFiles(".", managerSettings.FilePatterns, manager.GlobalConfig.IgnorePatterns)
	manager.logger.Debug(fmt.Sprintf("Found %d matching file(s)", len(candidates)))
	if err != nil {
		return err
	}

	// Prepare the marker regex which searches the file for the inline markers
	markerRegex := regexp.MustCompile("(?m)^[[:blank:]]*[/#*`]+ gonovate: (.+)\\s*$")

	// Process all candidates
	for _, candidate := range candidates {
		fileLogger := manager.logger.With(slog.String("file", candidate))
		fileLogger.Debug(fmt.Sprintf("Processing file '%s'", candidate))

		// Read the entire file
		fileContentBytes, err := os.ReadFile(candidate)
		if err != nil {
			return err
		}
		fileContent := string(fileContentBytes)

		// Search for the markers
		matchesIndex := markerRegex.FindAllStringSubmatchIndex(fileContent, -1)
		indexOffset := 0
		for _, match := range matchesIndex {
			start := match[2] + indexOffset
			end := match[3] + indexOffset
			configStr := fileContent[start:end]

			// Get the config for the marker
			config := &inlineManagerConfig{}
			if err = json.Unmarshal([]byte(configStr), config); err != nil {
				return fmt.Errorf("failed parsing marker config at position %d: %w", start, err)
			}

			// Build the regex that was defined in the marker
			newReg := regexp.MustCompile(config.Matchstring)
			// Search the remaining file content with this new regex and process the first match only
			contentSearchStart := end + 1
			matchList := findAllNamedMatchesWithIndex(newReg, fileContent[contentSearchStart:], false, 1)
			if matchList == nil || len(matchList) < 1 {
				return fmt.Errorf("regex defined in marker at position %d did not match anything", start)
			}
			// We are only interested in the first match
			match := matchList[0]
			// The version must be found with the regexp on the line
			versionObject, versionOk := match["version"]
			if !versionOk {
				// The version field is mandatory
				return fmt.Errorf("the field 'version' did not match")
			}
			//  Optional fields
			datasourceObject, datasourceOk := match["datasource"]
			packageObject, packageOk := match["packageName"]
			versioningObject, versioningOk := match["versioning"]

			// Build a package settings from the marker and the following direct match. This rule always has the highest priority
			priorityPackageSettings := &core.PackageSettings{}
			if config.PackageName != "" {
				priorityPackageSettings.PackageName = config.PackageName
			}
			if config.Datasource != "" {
				priorityPackageSettings.Datasource = config.Datasource
			}
			if config.Versioning != "" {
				priorityPackageSettings.Versioning = config.Versioning
			}
			if packageOk {
				priorityPackageSettings.PackageName = packageObject[0].Value
			}
			if datasourceOk {
				priorityPackageSettings.Datasource = datasourceObject[0].Value
			}
			if versioningOk {
				priorityPackageSettings.Versioning = versioningObject[0].Value
			}
			// Build the merge package settings
			packageSettings, err := buildMergedPackageSettings(manager.Config.PackageSettings, priorityPackageSettings, possiblePackageRules, candidate)
			if err != nil {
				return err
			}

			// Search for a new version for the package
			currentVersion := manager.sanitizeString(versionObject[0].Value)
			newReleaseInfo, err := manager.searchPackageUpdate(currentVersion, packageSettings, manager.GlobalConfig.HostRules)
			if err != nil {
				return err
			}
			if newReleaseInfo != nil {
				if err := platform.PrepareForChanges(packageSettings.PackageName, currentVersion, newReleaseInfo.Version.Raw); err != nil {
					return err
				}

				// Prepare the changes
				replaceStart := versionObject[0].StartIndex + contentSearchStart
				replaceEnd := versionObject[0].EndIndex + contentSearchStart
				fileContent = fileContent[:replaceStart] + newReleaseInfo.Version.Raw + fileContent[replaceEnd:]
				oldLength := len(versionObject[0].Value)
				newLength := len(newReleaseInfo.Version.Raw)
				indexOffset += newLength - oldLength
				// Write the file with the changes
				if err := os.WriteFile(candidate, []byte(fileContent), os.ModePerm); err != nil {
					return err
				}

				if err := platform.SubmitChanges(packageSettings.PackageName, currentVersion, newReleaseInfo.Version.Raw); err != nil {
					return err
				}
				if err := platform.PublishChanges(); err != nil {
					return err
				}
				if err := platform.ResetToBase(); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type inlineManagerConfig struct {
	PackageName string `json:"packageName"`
	Datasource  string `json:"datasource"`
	Matchstring string `json:"matchstring"`
	Versioning  string `json:"versioning"`
}

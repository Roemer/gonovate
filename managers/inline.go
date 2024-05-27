package managers

import (
	"cmp"
	"encoding/json"
	"fmt"
	"gonovate/core"
	"log/slog"
	"os"
	"regexp"
	"slices"

	"github.com/samber/lo"
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

func (manager *InlineManager) getChanges() ([]core.IChange, error) {
	// Process all rules to apply the ones relevant for the manager and store the ones relevant for packages.
	managerSettings, possiblePackageRules := manager.GlobalConfig.FilterForManager(manager.Config)

	// Skip if it is disabled
	if managerSettings.Disabled != nil && *managerSettings.Disabled {
		manager.logger.Info(fmt.Sprintf("Skipping Manager '%s' (%s) as it is disabled", manager.Config.Id, manager.Config.Type))
		return nil, nil
	}

	// Search file candidates
	manager.logger.Debug(fmt.Sprintf("Searching files with %d pattern(s)", len(managerSettings.FilePatterns)))
	candidates, err := core.SearchFiles(".", managerSettings.FilePatterns, manager.GlobalConfig.IgnorePatterns)
	manager.logger.Debug(fmt.Sprintf("Found %d matching file(s)", len(candidates)))
	if err != nil {
		return nil, err
	}

	// Prepare the marker regex which searches the file for the inline markers
	markerRegex := regexp.MustCompile("(?m)^[[:blank:]]*[/#*`]+ gonovate: (.+)\\s*$")

	// Process all candidates and collect the changes
	changes := []core.IChange{}
	for _, candidate := range candidates {
		fileLogger := manager.logger.With(slog.String("file", candidate))
		fileLogger.Debug(fmt.Sprintf("Processing file '%s'", candidate))

		// Read the entire file
		fileContentBytes, err := os.ReadFile(candidate)
		if err != nil {
			return nil, err
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
				return nil, fmt.Errorf("failed parsing marker config at position %d: %w", start, err)
			}

			// Build the regex that was defined in the marker
			newReg := regexp.MustCompile(config.Matchstring)
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
				return nil, err
			}

			// Search for a new version for the package
			currentVersionString, _ := manager.sanitizeString(versionObject[0].Value)
			newReleaseInfo, currentVersion, err := manager.searchPackageUpdate(currentVersionString, packageSettings, manager.GlobalConfig.HostRules)
			if err != nil {
				return nil, err
			}
			if newReleaseInfo != nil {
				// There is a new version, so build the change object
				change := &inlineManagerChange{
					ChangeMeta: &core.ChangeMeta{
						Datasource:     packageSettings.Datasource,
						PackageName:    packageSettings.PackageName,
						File:           candidate,
						CurrentVersion: currentVersion,
						NewRelease:     newReleaseInfo,
					},
					StartIndex: versionObject[0].StartIndex + contentSearchStart,
					EndIndex:   versionObject[0].EndIndex + contentSearchStart,
					Difference: len(newReleaseInfo.Version.Raw) - len(versionObject[0].Value),
				}
				// Add the change
				changes = append(changes, change)
			}
		}
	}

	// Return the changes
	return changes, nil
}

func (manager *InlineManager) applyChanges(changes []core.IChange) error {
	// Convert the changes to the manager specific change
	changesTyped := lo.Map(changes, func(x core.IChange, _ int) *inlineManagerChange { return x.(*inlineManagerChange) })
	// Group the changes by file
	changesGroupedByFile := lo.GroupBy(changesTyped, func(i *inlineManagerChange) string {
		return i.File
	})
	// Loop thru the changes by file
	for file, changesForFile := range changesGroupedByFile {
		// Sort the changes by startindex
		slices.SortFunc(changesForFile, func(a, b *inlineManagerChange) int {
			return cmp.Compare(a.StartIndex, b.StartIndex)
		})
		// Read the file
		fileContentBytes, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		fileContent := string(fileContentBytes)
		// Apply the changes
		offset := 0
		for _, change := range changesForFile {
			// Replace the version
			fileContent = fileContent[:change.StartIndex+offset] + change.NewRelease.Version.Raw + fileContent[change.EndIndex+offset:]
			// Adjust the offset in case the length of the versions is different
			offset += change.Difference
		}
		// Write the file with the changes
		if err := os.WriteFile(file, []byte(fileContent), os.ModePerm); err != nil {
			return err
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

// The manager-specific change object that contains everything needed to apply the change
type inlineManagerChange struct {
	*core.ChangeMeta
	StartIndex int
	EndIndex   int
	Difference int
}

func (change *inlineManagerChange) GetMeta() *core.ChangeMeta {
	return change.ChangeMeta
}

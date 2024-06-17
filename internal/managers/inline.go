package managers

import (
	"cmp"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"slices"

	"github.com/roemer/gonovate/internal/config"
	"github.com/roemer/gonovate/internal/core"
	"github.com/samber/lo"
)

type InlineManager struct {
	managerBase
}

func NewInlineManager(logger *slog.Logger, config *config.Config, managerConfig *config.Manager) IManager {
	manager := &InlineManager{
		managerBase: managerBase{
			logger:        logger.With(slog.String("handlerId", managerConfig.Id)),
			Config:        config,
			ManagerConfig: managerConfig,
		},
	}
	manager.impl = manager
	return manager
}

func (manager *InlineManager) getChanges(mergedManagerSettings *config.ManagerSettings, possiblePackageRules []*config.Rule) ([]core.IChange, error) {
	// Search file candidates
	manager.logger.Debug(fmt.Sprintf("Searching files with %d pattern(s)", len(mergedManagerSettings.FilePatterns)))
	candidates, err := core.SearchFiles(".", mergedManagerSettings.FilePatterns, manager.Config.IgnorePatterns)
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

			// Build a package settings from the marker and the following direct match. This rule always has the highest priority
			priorityPackageSettings := &config.PackageSettings{}
			if inlineConfig.PackageName != "" {
				priorityPackageSettings.PackageName = inlineConfig.PackageName
			}
			if inlineConfig.Datasource != "" {
				priorityPackageSettings.Datasource = inlineConfig.Datasource
			}
			if inlineConfig.Versioning != "" {
				priorityPackageSettings.Versioning = inlineConfig.Versioning
			}
			if inlineConfig.MaxUpdateType != "" {
				priorityPackageSettings.MaxUpdateType = inlineConfig.MaxUpdateType
			}
			if inlineConfig.ExtractVersion != "" {
				priorityPackageSettings.ExtractVersion = inlineConfig.ExtractVersion
			}
			// Now overwrite from direct matches
			if packageOk {
				priorityPackageSettings.PackageName = packageObject[0].Value
			}
			if datasourceOk {
				priorityPackageSettings.Datasource = core.DatasourceType(datasourceObject[0].Value)
			}
			if versioningOk {
				priorityPackageSettings.Versioning = versioningObject[0].Value
			}
			if maxUpdateTypeOk {
				priorityPackageSettings.MaxUpdateType = core.UpdateType(maxUpdateTypeObject[0].Value)
			}
			if extractVersionOk {
				priorityPackageSettings.ExtractVersion = extractVersionObject[0].Value
			}
			// Build the merge package settings
			packageSettings, err := buildMergedPackageSettings(nil, priorityPackageSettings, possiblePackageRules, candidate, manager.ManagerConfig.Id)
			if err != nil {
				return nil, err
			}

			// Search for a new version for the package
			currentVersionString, _ := manager.sanitizeString(versionObject[0].Value)
			newReleaseInfo, currentVersion, err := manager.searchPackageUpdate(currentVersionString, packageSettings)
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
	PackageName    string              `json:"packageName"`
	Datasource     core.DatasourceType `json:"datasource"`
	MatchString    string              `json:"matchString"`
	Versioning     string              `json:"versioning"`
	MaxUpdateType  core.UpdateType     `json:"maxUpdateType"`
	ExtractVersion string              `json:"extractVersion"`
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

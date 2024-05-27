package managers

import (
	"cmp"
	"fmt"
	"gonovate/core"
	"log/slog"
	"os"
	"regexp"
	"slices"

	"github.com/samber/lo"
)

type RegexManager struct {
	managerBase
}

func NewRegexManager(logger *slog.Logger, globalConfig *core.Config, managerConfig *core.Manager) IManager {
	manager := &RegexManager{
		managerBase: managerBase{
			logger:       logger.With(slog.String("handlerId", managerConfig.Id)),
			GlobalConfig: globalConfig,
			Config:       managerConfig,
		},
	}
	manager.impl = manager
	return manager
}

func (manager *RegexManager) getChanges() ([]core.IChange, error) {
	manager.logger.Info(fmt.Sprintf("Starting RegexManager with Id %s", manager.Config.Id))

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

	// Precompile the regexes
	precompiledRegexList := []*regexp.Regexp{}
	for _, regStr := range managerSettings.MatchStrings {
		regex, err := regexp.Compile(regStr)
		if err != nil {
			return nil, err
		}
		precompiledRegexList = append(precompiledRegexList, regex)
	}
	manager.logger.Debug(fmt.Sprintf("Found %d match pattern(s) to process", len(precompiledRegexList)))

	// Process all candidates
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

		// Loop thru all regex patterns
		for _, regex := range precompiledRegexList {
			matchList := findAllNamedMatchesWithIndex(regex, fileContent, false, -1)
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
				packageObject, packageOk := match["packageName"]
				versioningObject, versioningOk := match["versioning"]

				// Log
				fileLogger.Debug(fmt.Sprintf("Found a match for regex '%s'", regex.String()))

				// Build a package settings from the direct match. This rule always has the highest priority
				priorityPackageSettings := &core.PackageSettings{}
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
					change := &regexManagerChange{
						ChangeMeta: &core.ChangeMeta{
							Datasource:              packageSettings.Datasource,
							PackageName:             packageSettings.PackageName,
							File:                    candidate,
							CurrentVersion:          currentVersion,
							NewRelease:              newReleaseInfo,
							PostUpgradeReplacements: packageSettings.PostUpgradeReplacements,
						},
						StartIndex: versionObject[0].StartIndex,
						EndIndex:   versionObject[0].EndIndex,
						Difference: len(newReleaseInfo.Version.Raw) - len(versionObject[0].Value),
					}
					// Add the change
					changes = append(changes, change)
				}
			}
		}
	}

	// Return the changes
	return changes, nil
}

func (manager *RegexManager) applyChanges(changes []core.IChange) error {
	// Convert the changes to the manager specific change
	changesTyped := lo.Map(changes, func(x core.IChange, _ int) *regexManagerChange { return x.(*regexManagerChange) })
	// Group the changes by file
	changesGroupedByFile := lo.GroupBy(changesTyped, func(i *regexManagerChange) string {
		return i.File
	})
	// Loop thru the changes by file
	for file, changesForFile := range changesGroupedByFile {
		// Sort the changes by startindex
		slices.SortFunc(changesForFile, func(a, b *regexManagerChange) int {
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

	// Run Post-Upgrade replacements
	for file, changesForFile := range changesGroupedByFile {
		// Check if there are any post-upgrade replacements
		hasPostUpgradeReplacements := lo.ContainsBy(changesForFile, func(change *regexManagerChange) bool { return len(change.PostUpgradeReplacements) > 0 })
		if hasPostUpgradeReplacements {
			// Read the file
			fileContentBytes, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			fileContent := string(fileContentBytes)
			// Apply the replacements
			for _, change := range changesForFile {
				for _, reStr := range change.PostUpgradeReplacements {
					re := regexp.MustCompile(reStr)
					fileContent, _ = replaceMatchesInRegex(re, fileContent, map[string]string{
						"version": change.NewRelease.Version.Raw,
						"sha1":    change.NewRelease.Hashes["sha1"],
						"sha256":  change.NewRelease.Hashes["sha256"],
						"sha512":  change.NewRelease.Hashes["sha512"],
						"md5":     change.NewRelease.Hashes["md5"],
					})
				}
			}
			// Write the file with the changes
			if err := os.WriteFile(file, []byte(fileContent), os.ModePerm); err != nil {
				return err
			}
		}
	}
	return nil
}

func replaceMatchesInRegex(regex *regexp.Regexp, str string, replacementMap map[string]string) (string, int) {
	matchList := findAllNamedMatchesWithIndex(regex, str, true, -1)
	orderedCaptures := []*capturedGroup{}
	for _, match := range matchList {
		for _, value := range match {
			orderedCaptures = append(orderedCaptures, value...)
		}
	}
	// Make sure the sorting is correct (by startIndex)
	slices.SortFunc(orderedCaptures, func(a, b *capturedGroup) int {
		return cmp.Compare(a.StartIndex, b.StartIndex)
	})
	diff := 0
	for _, value := range orderedCaptures {
		str = str[:(value.StartIndex+diff)] + replacementMap[value.Key] + str[value.EndIndex+diff:]
		diff += len(replacementMap[value.Key]) - len(value.Value)
	}
	return str, diff
}

// The manager-specific change object that contains everything needed to apply the change
type regexManagerChange struct {
	*core.ChangeMeta
	StartIndex int
	EndIndex   int
	Difference int
}

func (change *regexManagerChange) GetMeta() *core.ChangeMeta {
	return change.ChangeMeta
}

package managers

import (
	"cmp"
	"fmt"
	"gonovate/core"
	"gonovate/platforms"
	"log/slog"
	"os"
	"regexp"
	"slices"
)

type RegexManager struct {
	managerBase
}

func NewRegexManager(logger *slog.Logger, globalConfig *core.Config, managerConfig *core.Manager, platform platforms.IPlatform) IManager {
	manager := &RegexManager{
		managerBase: managerBase{
			logger:       logger.With(slog.String("handlerId", managerConfig.Id)),
			GlobalConfig: globalConfig,
			Config:       managerConfig,
			Platform:     platform,
		},
	}
	manager.impl = manager
	return manager
}

func (manager *RegexManager) process() error {
	manager.logger.Info(fmt.Sprintf("Starting RegexManager with Id %s", manager.Config.Id))

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

	// Precompile the regexes
	precompiledRegexList := []*regexp.Regexp{}
	for _, regStr := range managerSettings.MatchStrings {
		regex, err := regexp.Compile(regStr)
		if err != nil {
			return err
		}
		precompiledRegexList = append(precompiledRegexList, regex)
	}
	manager.logger.Debug(fmt.Sprintf("Found %d match pattern(s) to process", len(precompiledRegexList)))

	// Precompile Post-Upgrade regexes
	precompiledPostUpgradeRegexList := []*regexp.Regexp{}
	for _, regStr := range managerSettings.PostUpgradeReplacements {
		regex, err := regexp.Compile(regStr)
		if err != nil {
			return err
		}
		precompiledPostUpgradeRegexList = append(precompiledPostUpgradeRegexList, regex)
	}
	manager.logger.Debug(fmt.Sprintf("Found %d post-upgrade-replacement pattern(s) to process", len(precompiledPostUpgradeRegexList)))

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

		// Loop thru all regex patterns
		for _, regex := range precompiledRegexList {
			matchList := findAllNamedMatchesWithIndex(regex, fileContent, false, -1)
			if matchList == nil {
				// The regex was not matched, go to the next
				continue
			}
			// Process the individual matches in reverse order so the indexes do not break when replacing text
			slices.Reverse(matchList)
			for _, match := range matchList {
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
					return err
				}

				// Search for a new version for the package
				currentVersionString, _ := manager.sanitizeString(versionObject[0].Value)
				newReleaseInfo, _, err := manager.searchPackageUpdate(currentVersionString, packageSettings, manager.GlobalConfig.HostRules)
				if err != nil {
					return err
				}
				if newReleaseInfo != nil {
					// Build the new content with the new version number
					fileContent = fileContent[:versionObject[0].StartIndex] + newReleaseInfo.Version.Raw + fileContent[versionObject[0].EndIndex:]

					// Run Post-Upgrade replacements
					if len(precompiledPostUpgradeRegexList) > 0 {
						for _, re := range precompiledPostUpgradeRegexList {
							fileContent = replaceMatchesInRegex(re, fileContent, map[string]string{
								"version": newReleaseInfo.Version.Raw,
								"sha256":  newReleaseInfo.Hashes["sha256"],
								"md5":     newReleaseInfo.Hashes["md5"],
							})
						}
					}
				}
			}
		}

		// Write the file back
		if err := os.WriteFile(candidate+"2", []byte(fileContent), os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func (manager *RegexManager) resetForNewGroup() error {
	return nil
}

func (manager *RegexManager) applyChanges(changes []core.IChange) error {
	return nil
}

func replaceMatchesInRegex(regex *regexp.Regexp, str string, replacementMap map[string]string) string {
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
	return str
}

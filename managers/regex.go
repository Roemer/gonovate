package managers

import (
	"cmp"
	"fmt"
	"gonovate/core"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"
)

type RegexManager struct {
	managerBase
}

func NewRegexManager(logger *slog.Logger, globalConfig *core.Config, managerConfig *core.Manager) *RegexManager {
	manager := &RegexManager{
		managerBase: managerBase{
			logger:       logger.With(slog.String("handlerId", managerConfig.Id)),
			GlobalConfig: globalConfig,
			Config:       managerConfig,
		},
	}
	return manager
}

func (manager *RegexManager) Run() error {
	err := manager.process()
	if err != nil {
		manager.logger.Error(fmt.Sprintf("Manager failed with error: %s", err.Error()))
	}
	return err
}

func (manager *RegexManager) process() error {
	manager.logger.Info(fmt.Sprintf("Starting RegexManager with Id %s", manager.Config.Id))

	// Process all rules to apply the ones relevant for the manager and store the ones relevant for packages.
	managerSettings, possiblePackageRules := manager.GlobalConfig.FilterForManager(manager.Config)

	// Skip if it is disabled
	if managerSettings.Disabled {
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
			matchList := findAllNamedMatchesWithIndex(regex, fileContent, false)
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
				packageObject, packageOk := match["package"]
				versioningObject, versioningOk := match["versioning"]

				// Log
				fileLogger.Debug(fmt.Sprintf("Found a match for regex '%s'", regex.String()))

				// Build the packageSettings with all relevant rules
				packageSettings := &core.PackageSettings{}
				// Initially add the package settings from the manager (if any)
				packageSettings.MergeWith(manager.Config.PackageSettings)
				// Initially set the fields that can be defined from matches from the regexp
				if packageOk {
					packageSettings.PackageName = packageObject[0].value
				}
				if datasourceOk {
					packageSettings.Datasource = datasourceObject[0].value
				}
				if versioningOk {
					packageSettings.Versioning = versioningObject[0].value
				}
				// Loop thru the rules and apply the ones that match
				for _, rule := range possiblePackageRules {
					isAnyMatch := rule.Matches.IsMatchAll()
					// Check if there is at least one condition that matches
					if !isAnyMatch && rule.Matches != nil {
						// Manager
						if !isAnyMatch && len(rule.Matches.Managers) > 0 {
							if slices.Contains(rule.Matches.Managers, core.MANAGER_TYPE_REGEX) {
								isAnyMatch = true
							}
						}
						// Files
						if !isAnyMatch && len(rule.Matches.Files) > 0 {
							isMatch, err := core.FilePathMatchesPattern(candidate, rule.Matches.Files...)
							if err != nil {
								return err
							}
							if isMatch {
								isAnyMatch = true
							}
						}
						// Package
						if !isAnyMatch && len(rule.Matches.Packages) > 0 {
							if packageSettings.PackageName != "" && slices.Contains(rule.Matches.Packages, packageSettings.PackageName) {
								isAnyMatch = true
							}
						}
						// Datasource
						if !isAnyMatch && len(rule.Matches.Datasources) > 0 {
							if packageSettings.Datasource != "" && slices.Contains(rule.Matches.Datasources, packageSettings.Datasource) {
								isAnyMatch = true
							}
						}
					}
					// The rule has at least one match, add it
					if isAnyMatch {
						packageSettings.MergeWith(rule.PackageSettings)
						// Make sure that the optional fields from the match are not overwritten
						if packageOk {
							packageSettings.PackageName = packageObject[0].value
						}
						if datasourceOk {
							packageSettings.Datasource = datasourceObject[0].value
						}
						if versioningOk {
							packageSettings.Versioning = versioningObject[0].value
						}
					}
				}

				// Search for a new version for the package
				newReleaseInfo, err := manager.searchPackageUpdate(versionObject[0].value, packageSettings, manager.GlobalConfig.HostRules)
				if err != nil {
					return err
				}
				if newReleaseInfo != nil {
					// Build the new content with the new version number
					fileContent = fileContent[:versionObject[0].startIndex] + newReleaseInfo.Version.Raw + fileContent[versionObject[0].endIndex:]

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

func replaceMatchesInRegex(regex *regexp.Regexp, str string, replacementMap map[string]string) string {
	matchList := findAllNamedMatchesWithIndex(regex, str, true)
	orderedCaptures := []*capturedGroup{}
	for _, match := range matchList {
		for _, value := range match {
			orderedCaptures = append(orderedCaptures, value...)
		}
	}
	// Make sure the sorting is correct (by startIndex)
	slices.SortFunc(orderedCaptures, func(a, b *capturedGroup) int {
		return cmp.Compare(a.startIndex, b.startIndex)
	})
	diff := 0
	for _, value := range orderedCaptures {
		str = str[:(value.startIndex+diff)] + replacementMap[value.key] + str[value.endIndex+diff:]
		diff += len(replacementMap[value.key]) - len(value.value)
	}
	return str
}

// Find all named matches in the given string, returning an list of objects with start/end-index and the value for each named match
func findAllNamedMatchesWithIndex(regex *regexp.Regexp, str string, includeNotMatchedOptional bool) []map[string][]*capturedGroup {
	matchIndexPairsList := regex.FindAllStringSubmatchIndex(str, -1)
	if matchIndexPairsList == nil {
		// No matches
		return nil
	}

	subexpNames := regex.SubexpNames()
	allResults := []map[string][]*capturedGroup{}
	for _, matchIndexPairs := range matchIndexPairsList {
		results := map[string][]*capturedGroup{}
		// Loop thru the subexp names (skipping the first empty one which is the full match)
		for i, name := range (subexpNames)[1:] {
			if name == "" {
				// No name, so skip it
				continue
			}
			startIndex := matchIndexPairs[(i+1)*2]
			endIndex := matchIndexPairs[(i+1)*2+1]
			if startIndex == -1 || endIndex == -1 {
				// No match found
				if includeNotMatchedOptional {
					// Add anyways
					results[name] = append(results[name], &capturedGroup{startIndex: -1, endIndex: -1, key: name, value: ""})
				}
				continue
			}
			// Assign the correct value
			results[name] = append(results[name], &capturedGroup{startIndex: startIndex, endIndex: endIndex, key: name, value: str[startIndex:endIndex]})
		}
		allResults = append(allResults, results)
	}

	return allResults
}

type capturedGroup struct {
	startIndex int
	endIndex   int
	key        string
	value      string
}

func (cg capturedGroup) String() string {
	return fmt.Sprintf("%d->%d:%s:%s", cg.startIndex, cg.endIndex, cg.key, strings.ReplaceAll(strings.ReplaceAll(cg.value, "\r", "\\r"), "\n", "\\n"))
}

package managers

import (
	"bufio"
	"fmt"
	"gonovate/core"
	"log/slog"
	"os"
	"regexp"
	"slices"
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
	manager.logger.Debug(fmt.Sprintf("Found %d line pattern(s) to process", len(precompiledRegexList)))

	// Process all candidates
	for _, candidate := range candidates {
		fileLogger := manager.logger.With(slog.String("file", candidate))
		fileLogger.Debug(fmt.Sprintf("Processing file '%s'", candidate))
		// Open the file
		f, err := os.OpenFile(candidate, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return err
		}
		defer f.Close()

		// Process the file line by line
		var outputLines []string
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			// Get the line
			line := sc.Text()
			// Loop thru the regexes
			for _, regex := range precompiledRegexList {
				// Execute the regex and get the matches with the needed info
				m := findNamedMatchesWithIndex(regex, line, false)
				if m == nil {
					// The line did not match the regex
					continue
				}

				// The version must be found with the regexp on the line
				versionObject, versionOk := m["version"]
				if !versionOk {
					// The version field is mandatory
					return fmt.Errorf("the field 'version' did not match on line '%s'", line)
				}
				//  Optional fields
				datasourceObject, datasourceOk := m["datasource"]
				packageObject, packageOk := m["package"]
				versioningObject, versioningOk := m["versioning"]

				// Log
				fileLogger.Debug(fmt.Sprintf("Found a match for regex '%s'", regex.String()))

				// Build the packageSettings with all relevant rules
				packageSettings := &core.PackageSettings{}
				// Initially add the package settings from the manager (if any)
				packageSettings.MergeWith(manager.Config.PackageSettings)
				// Initially set the fields that can be defined from matches from the regexp
				if packageOk {
					packageSettings.PackageName = packageObject.value
				}
				if datasourceOk {
					packageSettings.Datasource = datasourceObject.value
				}
				if versioningOk {
					packageSettings.Versioning = versioningObject.value
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
							packageSettings.PackageName = packageObject.value
						}
						if datasourceOk {
							packageSettings.Datasource = datasourceObject.value
						}
						if versioningOk {
							packageSettings.Versioning = versioningObject.value
						}
					}
				}

				// Search for a new version for the package
				newVersion, hasUpdate, err := manager.searchPackageUpdate(versionObject.value, packageSettings, manager.GlobalConfig.HostRules)
				if err != nil {
					return err
				}
				if hasUpdate {
					// Build the new line with the new version number
					line = line[:versionObject.startIndex] + newVersion + line[versionObject.endIndex:]
				}
			}
			// Add the original or modified line back to the output
			outputLines = append(outputLines, line)
		}
		if err := sc.Err(); err != nil {
			return err
		}

		// Write the file back
		file, err := os.Create(candidate + "2")
		if err != nil {
			return err
		}
		defer file.Close()

		w := bufio.NewWriter(file)
		for _, line := range outputLines {
			fmt.Fprintln(w, line)
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}

	return nil
}

// Find all named matches in the given index, returning an object with start/end-index and the value for each named match
func findNamedMatchesWithIndex(regex *regexp.Regexp, str string, includeNotMatchedOptional bool) map[string]*capturedGroup {
	matchIndexPairs := regex.FindStringSubmatchIndex(str)
	if matchIndexPairs == nil {
		// No matches
		return nil
	}
	subexpNames := regex.SubexpNames()
	results := map[string]*capturedGroup{}
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
				results[name] = &capturedGroup{startIndex: -1, endIndex: -1, value: ""}
			}
			continue
		}
		// Assign the correct value
		results[name] = &capturedGroup{startIndex: startIndex, endIndex: endIndex, value: str[startIndex:endIndex]}
	}
	return results
}

type capturedGroup struct {
	startIndex int
	endIndex   int
	value      string
}

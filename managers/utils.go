package managers

import (
	"fmt"
	"gonovate/core"
	"regexp"
	"slices"
	"strings"
)

// Build a PackageSettings object out of the various settings that can be relevant.
func buildMergedPackageSettings(initialPackageSettings, priorityPackageSettings *core.PackageSettings, possiblePackageRules []*core.Rule, currentFile string) (*core.PackageSettings, error) {
	// Build the packageSettings which holds all relevant rules
	packageSettings := &core.PackageSettings{}
	// Merge the initial package settings (usually from the manager)
	packageSettings.MergeWith(initialPackageSettings)
	// Initially apply the priority settings (as they can be used to evaluate further matches)
	packageSettings.MergeWith(priorityPackageSettings)
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
				isMatch, err := core.FilePathMatchesPattern(currentFile, rule.Matches.Files...)
				if err != nil {
					return nil, err
				}
				if isMatch {
					isAnyMatch = true
				}
			}
			// PackageName
			if !isAnyMatch && len(rule.Matches.PackageNames) > 0 {
				if packageSettings.PackageName != "" && slices.Contains(rule.Matches.PackageNames, packageSettings.PackageName) {
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
			// Merge the current rules package settings
			packageSettings.MergeWith(rule.PackageSettings)
			// Make sure that the priority settings are not overwritten
			packageSettings.MergeWith(priorityPackageSettings)
		}
	}

	return packageSettings, nil
}

// Find all named matches in the given string, returning an list of objects with start/end-index and the value for each named match
func findAllNamedMatchesWithIndex(regex *regexp.Regexp, str string, includeNotMatchedOptional bool, n int) []map[string][]*capturedGroup {
	matchIndexPairsList := regex.FindAllStringSubmatchIndex(str, n)
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
					results[name] = append(results[name], &capturedGroup{StartIndex: -1, EndIndex: -1, Key: name, Value: ""})
				}
				continue
			}
			// Assign the correct value
			results[name] = append(results[name], &capturedGroup{StartIndex: startIndex, EndIndex: endIndex, Key: name, Value: str[startIndex:endIndex]})
		}
		allResults = append(allResults, results)
	}

	return allResults
}

type capturedGroup struct {
	StartIndex int
	EndIndex   int
	Key        string
	Value      string
}

func (cg capturedGroup) String() string {
	return fmt.Sprintf("%d->%d:%s:%s", cg.StartIndex, cg.EndIndex, cg.Key, strings.ReplaceAll(strings.ReplaceAll(cg.Value, "\r", "\\r"), "\n", "\\n"))
}

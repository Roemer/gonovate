package shared

import (
	"fmt"
	"regexp"
	"strings"
)

func FindNamedMatchesWithIndex(regex *regexp.Regexp, str string, includeNotMatchedOptional bool) map[string][]*CapturedGroup {
	match := FindAllNamedMatchesWithIndex(regex, str, includeNotMatchedOptional, 1)
	if match != nil {
		return match[0]
	}
	return nil
}

// Find all named matches in the given string, returning an list of objects with start/end-index and the value for each named match
func FindAllNamedMatchesWithIndex(regex *regexp.Regexp, str string, includeNotMatchedOptional bool, n int) []map[string][]*CapturedGroup {
	matchIndexPairsList := regex.FindAllStringSubmatchIndex(str, n)
	if matchIndexPairsList == nil {
		// No matches
		return nil
	}

	subexpNames := regex.SubexpNames()
	allResults := []map[string][]*CapturedGroup{}
	for _, matchIndexPairs := range matchIndexPairsList {
		results := map[string][]*CapturedGroup{}
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
					results[name] = append(results[name], &CapturedGroup{StartIndex: -1, EndIndex: -1, Key: name, Value: ""})
				}
				continue
			}
			// Assign the correct value
			results[name] = append(results[name], &CapturedGroup{StartIndex: startIndex, EndIndex: endIndex, Key: name, Value: str[startIndex:endIndex]})
		}
		allResults = append(allResults, results)
	}

	return allResults
}

type CapturedGroup struct {
	StartIndex int
	EndIndex   int
	Key        string
	Value      string
}

func (cg *CapturedGroup) String() string {
	return fmt.Sprintf("%d->%d:%s:%s", cg.StartIndex, cg.EndIndex, cg.Key, strings.ReplaceAll(strings.ReplaceAll(cg.Value, "\r", "\\r"), "\n", "\\n"))
}

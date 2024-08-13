package shared

import (
	"fmt"
	"regexp"
	"strings"
)

func NormalizeString(value string, maxLength int) string {
	// Assign the initial value
	normalizedString := value
	// Make it lowercase
	normalizedString = strings.ToLower(normalizedString)
	// Remove all sort of invalid characters
	punctuationMatcher := regexp.MustCompile("[][!\"#$%&'()*+,/:;<=>?@\\\\^_`{|}~]")
	normalizedString = punctuationMatcher.ReplaceAllString(normalizedString, "-")
	// Replace multiple-hyphens with a single one
	repeatingHyphenMatcher := regexp.MustCompile("-{2,}")
	normalizedString = repeatingHyphenMatcher.ReplaceAllString(normalizedString, "-")
	// Replace multiple-dots with a single one
	repeatingDotMatcher := regexp.MustCompile(`\.{2,}`)
	normalizedString = repeatingDotMatcher.ReplaceAllString(normalizedString, ".")
	// Replace other invalid characters
	normalizedString = strings.ReplaceAll(normalizedString, "ä", "a")
	normalizedString = strings.ReplaceAll(normalizedString, "ö", "o")
	normalizedString = strings.ReplaceAll(normalizedString, "ü", "u")
	// Make sure it does not end with any of the defined chars
	invalidEndingMatcher := regexp.MustCompile(`[:\-./]+$`)
	normalizedString = invalidEndingMatcher.ReplaceAllString(normalizedString, "")
	// Shorten if needed
	if maxLength > 0 && len(normalizedString) > maxLength {
		middle := maxLength / 2
		firstBit := normalizedString[0:middle]
		lastBit := normalizedString[len(normalizedString)-(maxLength-middle)+2:]
		normalizedString = fmt.Sprintf("%s__%s", firstBit, lastBit)
	}
	// Make sure it does not end with a any of the defined chars (again)
	normalizedString = invalidEndingMatcher.ReplaceAllString(normalizedString, "")
	return normalizedString
}

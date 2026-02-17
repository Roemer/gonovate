package common

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
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

type TitleBuilderSettings struct {
	TitleTemplate  string
	DependencyName string
	GroupName      string
	NewRelease     *ReleaseInfo
}

type BranchNameBuilderSettings struct {
	BranchNameTemplate string
	BaseBranch         string
	DependencyName     string
	GroupName          string
	NewRelease         *ReleaseInfo
}

func BuildTitle(settings *TitleBuilderSettings) (string, error) {
	if settings == nil {
		return "", fmt.Errorf("settings is nil")
	}
	templateString := settings.TitleTemplate
	if templateString == "" {
		templateString = "Update {{if .GroupName}}group '{{.GroupName}}'{{else}}'{{.DependencyName}}' to '{{.NewVersion}}'{{end}}"
	}
	tmpl, err := template.New("tmpl").Option("missingkey=error").Parse(templateString)
	if err != nil {
		return "", fmt.Errorf("title template parse error (template=%q, dependency=%q, group=%q): %w", templateString, settings.DependencyName, settings.GroupName, err)
	}

	newVersion := ""
	if settings.NewRelease != nil {
		newVersion = settings.NewRelease.VersionString
	}
	updateType := ""
	if settings.NewRelease != nil {
		updateType = string(settings.NewRelease.UpdateType)
	}

	// Prepare the template data
	data := struct {
		GroupName      string
		DependencyName string
		NewVersion     string
		UpdateType     string
	}{
		GroupName:      settings.GroupName,
		DependencyName: settings.DependencyName,
		NewVersion:     newVersion,
		UpdateType:     updateType,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("title template execution error (dependency=%q, group=%q): %w", settings.DependencyName, settings.GroupName, err)
	}

	title := strings.TrimSpace(buf.String())
	if title == "" {
		return "", fmt.Errorf("generated title is empty (template=%q, dependency=%q, group=%q)", templateString, settings.DependencyName, settings.GroupName)
	}
	return title, nil
}

func BuildBranchName(settings *BranchNameBuilderSettings) (string, error) {
	if settings == nil {
		return "", fmt.Errorf("settings is nil")
	}
	templateString := settings.BranchNameTemplate
	if templateString == "" {
		templateString = "{{if .GroupName}}{{.GroupName}}{{else}}{{.BaseBranch}}-{{.DependencyName}}-{{.NewVersion}}{{end}}"
	}

	tmpl, err := template.New("tmpl").Option("missingkey=error").Parse(templateString)
	if err != nil {
		return "", fmt.Errorf("branch-name template parse error (template=%q, dependency=%q, group=%q): %w", templateString, settings.DependencyName, settings.GroupName, err)
	}

	newVersion := ""
	if settings.NewRelease != nil {
		newVersion = settings.NewRelease.VersionString
	}
	updateType := ""
	if settings.NewRelease != nil {
		updateType = string(settings.NewRelease.UpdateType)
	}

	// Prepare the template data
	data := struct {
		GroupName      string
		BaseBranch     string
		DependencyName string
		NewVersion     string
		UpdateType     string

		// Raw values (exposed for templates that need original strings)
		GroupNameRaw      string
		BaseBranchRaw     string
		DependencyNameRaw string
		NewVersionRaw     string
	}{
		GroupName:      NormalizeString(settings.GroupName, 20),
		BaseBranch:     NormalizeString(settings.BaseBranch, 20),
		DependencyName: NormalizeString(settings.DependencyName, 40),
		NewVersion:     NormalizeString(newVersion, 0),
		UpdateType:     NormalizeString(updateType, 0),
		// Raw values
		GroupNameRaw:      settings.GroupName,
		BaseBranchRaw:     settings.BaseBranch,
		DependencyNameRaw: settings.DependencyName,
		NewVersionRaw:     newVersion,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("branch-name template execution error (dependency=%q, group=%q): %w", settings.DependencyName, settings.GroupName, err)
	}

	normalizedBranchName := NormalizeString(buf.String(), 200)
	if normalizedBranchName == "" {
		return "", fmt.Errorf("normalized branch name is empty (template=%q, dependency=%q, group=%q)", templateString, settings.DependencyName, settings.GroupName)
	}
	return normalizedBranchName, nil
}

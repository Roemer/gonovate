package platforms

import (
	"fmt"
	"gonovate/core"
	"regexp"
	"strings"
)

type gitPlatform struct {
	platformBase
	BaseBranch string
}

func (p *gitPlatform) CreateBranch(change *core.ChangeMeta) error {
	branchName := fmt.Sprintf("gonovate/%s-%s",
		p.normalizeString(change.PackageName, 40),
		p.normalizeString(change.NewRelease.Version.Raw, 0))

	change.Data["branchName"] = branchName

	p.logger.Debug(fmt.Sprintf("Creating branch '%s'", branchName))

	_, _, err := core.Git{}.Run("checkout", "-B", branchName)
	return err
}

func (p *gitPlatform) AddAll() error {
	_, _, err := core.Git{}.Run("add", "--all")
	return err
}

func (p *gitPlatform) Commit(change *core.ChangeMeta) error {
	// Build the commit message
	msg := fmt.Sprintf("Update %s from %s to %s", change.PackageName, change.CurrentVersion.Raw, change.NewRelease.Version.Raw)
	// Store it for later use
	change.Data["msg"] = msg

	// Build the arguments
	args := []string{
		"commit",
		"--message=" + msg,
	}
	// Optionally add the author if it is set
	if p.Config.PlatformSettings != nil && p.Config.PlatformSettings.GitAuthor != "" {
		args = append(args, "--author="+p.Config.PlatformSettings.GitAuthor)
	}

	// Execute the command
	_, _, err := core.Git{}.Run(args...)
	return err
}

func (p *gitPlatform) PushBranch() error {
	_, _, err := core.Git{}.Run("push", "-u", "origin", "HEAD", "--force")
	return err
}

func (p *gitPlatform) CheckoutBaseBranch() error {
	_, _, err := core.Git{}.Run("checkout", p.BaseBranch)
	return err
}

func (p *gitPlatform) normalizeString(value string, maxLength int) string {
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
	// Make sure it does not end with a any of the defined chars
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

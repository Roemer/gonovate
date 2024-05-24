package platforms

import (
	"fmt"
	"gonovate/core"
	"log/slog"
	"regexp"
	"strings"
)

type GitPlatform struct {
	platformBase
	BaseBranch string
}

func NewGitPlatform(logger *slog.Logger, config *core.Config) *GitPlatform {
	platform := &GitPlatform{
		platformBase: platformBase{
			logger: logger,
			Config: config,
		},
		BaseBranch: "main",
	}
	return platform
}

func (p *GitPlatform) Type() string {
	return core.PLATFORM_TYPE_GIT
}

func (p *GitPlatform) SearchProjects() ([]*core.Project, error) {
	// Not available
	return nil, nil
}

func (p *GitPlatform) FetchProject(project *core.Project) error {
	// Not available
	return nil
}

func (p *GitPlatform) PrepareForChanges(change core.IChange) error {
	meta := change.GetMeta()
	branchName := fmt.Sprintf("gonovate/%s-%s",
		p.normalizeString(meta.PackageName, 40),
		p.normalizeString(meta.NewRelease.Version.Raw, 0))

	meta.Data["branchName"] = branchName

	p.logger.Debug(fmt.Sprintf("Creating branch '%s'", branchName))

	_, _, err := core.Git.Run("checkout", "-B", branchName)
	return err
}

func (p *GitPlatform) SubmitChanges(change core.IChange) error {
	meta := change.GetMeta()
	if _, _, err := core.Git.Run("add", "--all"); err != nil {
		return err
	}
	// Build the commit message
	msg := fmt.Sprintf("Update %s from %s to %s", meta.PackageName, meta.CurrentVersion.Raw, meta.NewRelease.Version.Raw)
	// Store it for later use
	meta.Data["msg"] = msg

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
	_, _, err := core.Git.Run(args...)
	return err
}

func (p *GitPlatform) PublishChanges(change core.IChange) error {
	_, _, err := core.Git.Run("push", "-u", "origin", "HEAD", "--force")
	return err
}

func (p *GitPlatform) NotifyChanges(change core.IChange) error {
	// Not available
	return nil
}

func (p *GitPlatform) ResetToBase() error {
	_, _, err := core.Git.Run("checkout", p.BaseBranch)
	return err
}

func (p *GitPlatform) normalizeString(value string, maxLength int) string {
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

package platforms

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type gitPlatform struct {
	platformBase
}

func (p *gitPlatform) CreateBranch(packageName, oldVersion, newVersion string) error {
	branchName := fmt.Sprintf("gonovate/update-%s-%s-%s", p.normalizeString(packageName, 30), p.normalizeString(oldVersion, 0), p.normalizeString(newVersion, 0))

	err := p.runGitCommand("checkout", "-B", branchName)
	if err != nil {
		return err
	}

	return nil
}

func (p *gitPlatform) AddAll() error {
	err := p.runGitCommand("add", "--all")
	if err != nil {
		return err
	}

	return nil
}

func (p *gitPlatform) Commit(packageName, oldVersion, newVersion string) error {
	msg := fmt.Sprintf("Update %s from %s to %s", packageName, oldVersion, newVersion)

	err := p.runGitCommand("commit", "-m", msg)
	if err != nil {
		return err
	}

	return nil
}

func (p *gitPlatform) PushBranch() error {
	err := p.runGitCommand("push", "-u", "origin", "HEAD", "--force")
	if err != nil {
		return err
	}

	return nil
}

func (p *gitPlatform) CheckoutBaseBranch() error {
	err := p.runGitCommand("checkout", "main")
	if err != nil {
		return err
	}

	return nil
}

func (p *gitPlatform) runGitCommand(arguments ...string) error {
	cmd := exec.Command("git", arguments...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func (p *gitPlatform) normalizeString(value string, maxLength int) string {
	// Assign the initial value
	normalizedString := value
	// Make it lowercase
	normalizedString = strings.ToLower(normalizedString)
	// Remove all sort of invalid characters
	punctuationMatcher := regexp.MustCompile("[][!\"#$%&'()*+,./:;<=>?@\\\\^_`{|}~]")
	normalizedString = punctuationMatcher.ReplaceAllString(normalizedString, "-")
	// Replace multiple-hyphens with a single one
	repeatingHyphenMatcher := regexp.MustCompile("-{2,}")
	normalizedString = repeatingHyphenMatcher.ReplaceAllString(normalizedString, "-")
	// Replace other invalid characters
	normalizedString = strings.ReplaceAll(normalizedString, "ä", "a")
	normalizedString = strings.ReplaceAll(normalizedString, "ö", "o")
	normalizedString = strings.ReplaceAll(normalizedString, "ü", "u")
	// Make sure it does not end with a hyphen
	invalidEndingMatcher := regexp.MustCompile("-+$")
	normalizedString = invalidEndingMatcher.ReplaceAllString(normalizedString, "")
	// Shorten if needed
	if maxLength > 0 && len(normalizedString) > maxLength {
		normalizedString = normalizedString[0:maxLength]
	}
	// Make sure it does not end with a hyphen (again)
	normalizedString = invalidEndingMatcher.ReplaceAllString(normalizedString, "")
	return normalizedString
}

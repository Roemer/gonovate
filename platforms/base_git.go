package platforms

import (
	"fmt"
	"gonovate/core"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type gitPlatform struct {
	platformBase
}

func (p *gitPlatform) CreateBranch(change *core.Change) error {
	branchName := fmt.Sprintf("gonovate/update-%s-%s-to-%s",
		p.normalizeString(change.PackageName, 30),
		p.normalizeString(change.OldVersion, 0),
		p.normalizeString(change.NewVersion, 0))

	change.Data["branchName"] = branchName

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

func (p *gitPlatform) Commit(change *core.Change) error {
	msg := fmt.Sprintf("Update %s from %s to %s", change.PackageName, change.OldVersion, change.NewVersion)

	change.Data["msg"] = msg

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

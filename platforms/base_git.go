package platforms

import (
	"bytes"
	"fmt"
	"gonovate/core"
	"os/exec"
	"regexp"
	"strings"
)

type gitPlatform struct {
	platformBase
	BaseBranch string
}

func (p *gitPlatform) CreateBranch(change *core.Change) error {
	branchName := fmt.Sprintf("gonovate/%s-%s",
		p.normalizeString(change.PackageName, 40),
		p.normalizeString(change.NewVersion, 0))

	change.Data["branchName"] = branchName

	_, _, err := p.runGitCommand("checkout", "-B", branchName)
	return err
}

func (p *gitPlatform) AddAll() error {
	_, _, err := p.runGitCommand("add", "--all")
	return err
}

func (p *gitPlatform) Commit(change *core.Change) error {
	// Build the commit message
	msg := fmt.Sprintf("Update %s from %s to %s", change.PackageName, change.OldVersion, change.NewVersion)
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
	_, _, err := p.runGitCommand(args...)
	return err
}

func (p *gitPlatform) PushBranch() error {
	_, _, err := p.runGitCommand("push", "-u", "origin", "HEAD", "--force")
	return err
}

func (p *gitPlatform) CheckoutBaseBranch() error {
	_, _, err := p.runGitCommand("checkout", p.BaseBranch)
	return err
}

func (p *gitPlatform) runGitCommand(arguments ...string) (string, string, error) {
	cmd := exec.Command("git", arguments...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	outStr, errStr := p.processOutputString(stdoutBuf.String()), p.processOutputString(stderrBuf.String())
	if err != nil {
		err = fmt.Errorf("git command failed: %w", err)
	}
	return outStr, errStr, err
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

func (p *gitPlatform) processOutputString(value string) string {
	return strings.TrimRight(value, "\r\n")
}

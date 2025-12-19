package platforms

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/samber/lo"
)

var authorRegex = regexp.MustCompile(`^(?P<name>[^<>]+)(?:\s+<(?P<email>.*?)>)?$`)

type GitPlatform struct {
	*platformBase
}

func NewGitPlatform(settings *common.PlatformSettings) *GitPlatform {
	platform := &GitPlatform{
		platformBase: newPlatformBase(settings),
	}
	return platform
}

func (p *GitPlatform) Type() common.PlatformType {
	return common.PLATFORM_TYPE_GIT
}

func (p *GitPlatform) FetchProject(project *common.Project) error {
	// Not available
	return nil
}

func (p *GitPlatform) LookupAuthor() (string, string, error) {
	return "gonovate-bot", "bot@gonovate.org", nil
}

func (p *GitPlatform) PrepareForChanges(updateGroup *common.UpdateGroup) error {
	p.logger.Debug(fmt.Sprintf("Creating branch '%s'", updateGroup.BranchName))
	_, _, err := common.Git.Run("checkout", "-B", updateGroup.BranchName)
	return err
}

func (p *GitPlatform) SubmitChanges(updateGroup *common.UpdateGroup) error {
	if _, _, err := common.Git.Run("add", "--all"); err != nil {
		return err
	}

	// Prepare the slice for the arguments
	args := []string{}

	// Optionally set the committer if it is set
	if p.settings != nil && p.settings.GitAuthor != "" {
		name, email := splitAuthor(p.settings.GitAuthor)
		args = append(args, "-c", "user.name="+name)
		args = append(args, "-c", "user.email="+email)
	} else {
		// Or look it up from the platforms default
		name, email, err := p.LookupAuthor()
		if err != nil {
			return err
		}
		args = append(args, "-c", "user.name="+name)
		args = append(args, "-c", "user.email="+email)
	}

	// Build the commit arguments
	args = append(args, "commit", "--message="+updateGroup.Title)

	// Execute the command
	_, _, err := common.Git.Run(args...)
	return err
}

func (p *GitPlatform) IsNewOrChanged(updateGroup *common.UpdateGroup) (bool, error) {
	_, _, err := common.Git.Run("ls-remote", "--exit-code", "origin", updateGroup.BranchName)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 2 {
			// The branch does not exist
			return true, nil
		}
		// There was an error
		return false, err
	}
	// The branch exists, get the diff
	stdOut, _, err := common.Git.Run("diff", fmt.Sprintf("origin/%s", updateGroup.BranchName), updateGroup.BranchName, "--name-status")
	if err != nil {
		return false, err
	}
	if len(stdOut) > 0 {
		return true, nil
	}
	return false, err
}

func (p *GitPlatform) PublishChanges(updateGroup *common.UpdateGroup) error {
	_, _, err := common.Git.Run("push", "-u", "origin", "HEAD", "--force")
	return err
}

func (p *GitPlatform) NotifyChanges(project *common.Project, updateGroup *common.UpdateGroup) error {
	// Not available
	return nil
}

func (p *GitPlatform) ResetToBase(baseBranch string) error {
	_, _, err := common.Git.Run("checkout", baseBranch)
	return err
}

func (p *GitPlatform) Cleanup(cleanupSettings *PlatformCleanupSettings) error {
	// Not available
	return nil
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

func (p *GitPlatform) getRemoteName() string {
	// TODO: Should maybe be dynamic, eg. by "git remote -v"
	return "origin"
}

func (p *GitPlatform) getRemoteGonovateBranches(remoteName string, branchPrefix string) ([]string, error) {
	// Get the remote branches
	stdout, _, err := common.Git.Run("ls-remote", "--heads", remoteName)
	if err != nil {
		return nil, err
	}
	allBranches := strings.Split(stdout, "\n")

	// Map to only the branch name
	lsRemoteRegex := regexp.MustCompile(`^[a-z0-9]+\s+refs/heads/(.*)$`)
	allBranches = lo.Map(allBranches, func(x string, _ int) string {
		matches := lsRemoteRegex.FindStringSubmatch(x)
		if len(matches) != 2 {
			return ""
		}
		return strings.TrimSpace(matches[1])
	})

	// Filter to those that are relevant for gonovate
	gonovateBranches := lo.Filter(allBranches, func(x string, _ int) bool {
		return strings.HasPrefix(x, branchPrefix)
	})

	return gonovateBranches, nil
}

func splitAuthor(author string) (string, string) {
	matchMap := common.FindNamedMatchesWithIndex(authorRegex, author, true)
	return matchMap["name"][0].Value, matchMap["email"][0].Value
}

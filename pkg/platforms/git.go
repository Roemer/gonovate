package platforms

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/samber/lo"
)

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

func (p *GitPlatform) PrepareForChanges(updateGroup *common.UpdateGroup) error {
	p.logger.Debug(fmt.Sprintf("Creating branch '%s'", updateGroup.BranchName))
	_, _, err := common.Git.Run("checkout", "-B", updateGroup.BranchName)
	return err
}

func (p *GitPlatform) SubmitChanges(updateGroup *common.UpdateGroup) error {
	if _, _, err := common.Git.Run("add", "--all"); err != nil {
		return err
	}

	// Build the arguments
	args := []string{
		"commit",
		"--message=" + updateGroup.Title,
	}
	// Optionally add the author if it is set
	if p.settings != nil && p.settings.GitAuthor != "" {
		args = append(args, "--author="+p.settings.GitAuthor)
	}

	// Execute the command
	_, _, err := common.Git.Run(args...)
	return err
}

func (p *GitPlatform) PublishChanges(updateGroup *common.UpdateGroup) error {
	_, _, err := common.Git.Run("push", "-u", "origin", "HEAD", "--force")
	return err
}

func (p *GitPlatform) NotifyChanges(project *common.Project, updateGroup *common.UpdateGroup) error {
	// Not available
	return nil
}

func (p *GitPlatform) ResetToBase() error {
	_, _, err := common.Git.Run("checkout", p.settings.BaseBranch)
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

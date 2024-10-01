package platforms

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
	"github.com/samber/lo"
)

type GitPlatform struct {
	platformBase
}

func NewGitPlatform(logger *slog.Logger, config *config.RootConfig) *GitPlatform {
	platform := &GitPlatform{
		platformBase: platformBase{
			logger: logger,
			Config: config,
		},
	}
	return platform
}

func (p *GitPlatform) Type() shared.PlatformType {
	return shared.PLATFORM_TYPE_GIT
}

func (p *GitPlatform) FetchProject(project *shared.Project) error {
	// Not available
	return nil
}

func (p *GitPlatform) PrepareForChanges(updateGroup *shared.UpdateGroup) error {
	p.logger.Debug(fmt.Sprintf("Creating branch '%s'", updateGroup.BranchName))
	_, _, err := shared.Git.Run("checkout", "-B", updateGroup.BranchName)
	return err
}

func (p *GitPlatform) SubmitChanges(updateGroup *shared.UpdateGroup) error {
	if _, _, err := shared.Git.Run("add", "--all"); err != nil {
		return err
	}

	// Build the arguments
	args := []string{
		"commit",
		"--message=" + updateGroup.Title,
	}
	// Optionally add the author if it is set
	if p.Config.PlatformSettings != nil && p.Config.PlatformSettings.GitAuthor != "" {
		args = append(args, "--author="+p.Config.PlatformSettings.GitAuthor)
	}

	// Execute the command
	_, _, err := shared.Git.Run(args...)
	return err
}

func (p *GitPlatform) PublishChanges(updateGroup *shared.UpdateGroup) error {
	_, _, err := shared.Git.Run("push", "-u", "origin", "HEAD", "--force")
	return err
}

func (p *GitPlatform) NotifyChanges(project *shared.Project, updateGroup *shared.UpdateGroup) error {
	// Not available
	return nil
}

func (p *GitPlatform) ResetToBase() error {
	_, _, err := shared.Git.Run("checkout", p.Config.PlatformSettings.BaseBranch)
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
	stdout, _, err := shared.Git.Run("ls-remote", "--heads", remoteName)
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

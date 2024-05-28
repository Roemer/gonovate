package platforms

import (
	"fmt"
	"gonovate/core"
	"log/slog"
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
		BaseBranch: config.PlatformSettings.BaseBranch,
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

func (p *GitPlatform) PrepareForChanges(changeSet *core.ChangeSet) error {
	branchName := changeSet.Id
	p.logger.Debug(fmt.Sprintf("Creating branch '%s'", branchName))
	_, _, err := core.Git.Run("checkout", "-B", branchName)
	return err
}

func (p *GitPlatform) SubmitChanges(changeSet *core.ChangeSet) error {
	if _, _, err := core.Git.Run("add", "--all"); err != nil {
		return err
	}

	// Build the arguments
	args := []string{
		"commit",
		"--message=" + changeSet.Title,
	}
	// Optionally add the author if it is set
	if p.Config.PlatformSettings != nil && p.Config.PlatformSettings.GitAuthor != "" {
		args = append(args, "--author="+p.Config.PlatformSettings.GitAuthor)
	}

	// Execute the command
	_, _, err := core.Git.Run(args...)
	return err
}

func (p *GitPlatform) PublishChanges(changeSet *core.ChangeSet) error {
	_, _, err := core.Git.Run("push", "-u", "origin", "HEAD", "--force")
	return err
}

func (p *GitPlatform) NotifyChanges(project *core.Project, changeSet *core.ChangeSet) error {
	// Not available
	return nil
}

func (p *GitPlatform) ResetToBase() error {
	_, _, err := core.Git.Run("checkout", p.BaseBranch)
	return err
}

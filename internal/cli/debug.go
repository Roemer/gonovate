package cli

import (
	"bufio"
	"flag"
	"fmt"
	"strings"

	"github.com/roemer/gonovate/internal/util"
)

func DebugCmd(args []string) error {
	// Flags and help for the command
	flagSet := flag.NewFlagSet("debug", flag.ExitOnError)
	flagSet.Usage = func() { printCmdUsage(flagSet, "debug", "subcommand") }
	flagSet.Parse(args)

	if flagSet.NArg() != 1 {
		return fmt.Errorf("no debug subcommand defined")
	}

	subcommand := flagSet.Arg(0)

	switch subcommand {
	case "clear-git-branches":
		if err := debugCmdClearGitBranches(); err != nil {
			return err
		}
	}

	return nil
}

func debugCmdClearGitBranches() error {
	stdout, _, err := util.Git.Run("branch", "--list", "gonovate/*")
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		branchName := strings.TrimSpace(scanner.Text())
		fmt.Printf("Deleting branch '%s'\n", branchName)
		_, _, err := util.Git.Run("branch", "-D", branchName)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return err
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"gonovate/core"
	"gonovate/managers"
	"gonovate/platforms"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/samber/lo"
)

// Holds information about a CLI command that can be executed
type Command struct {
	Name string
	Help string
	Run  func(args []string) error
}

// The list of CLI commands
var commands = []Command{
	{Name: "help", Help: "Prints this help", Run: helpCmd},
	{Name: "run", Help: "Runs the gonovate process", Run: runCmd},
	{Name: "debug", Help: "Tools usefull for/while debugging", Run: debugCmd},
}

func main() {
	// CLI flags
	flag.Usage = printUsage
	flag.Parse()

	// A command need to be passed
	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Command and the command arguments
	subCmd := flag.Arg(0)
	subCmdArgs := flag.Args()[1:]

	// Run the command
	runCommand(subCmd, subCmdArgs)
}

// Prints the base usage
func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  gonovate [flags] <command> [command flags]")
	fmt.Fprintln(os.Stderr, "")

	fmt.Fprintln(os.Stderr, "Commands:")
	for _, cmd := range commands {
		fmt.Fprintf(os.Stderr, "  %-8s %s\n", cmd.Name, cmd.Help)
	}

	// Uncomment if there are flags
	//fmt.Fprintln(os.Stderr, "Flags:")
	//fmt.Fprintln(os.Stderr, "")
	//flag.PrintDefaults()

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Run `gonovate <command> -h` to get help for a specific command\n\n")
}

// Prints the help for a command
func printCmdUsage(flagSet *flag.FlagSet, commandName, nonFlagArgs string) {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintf(os.Stderr, "  gonovate %s [flags]", commandName)
	if nonFlagArgs != "" {
		fmt.Fprint(os.Stderr, " "+nonFlagArgs)
	}
	fmt.Fprintln(os.Stderr, "")

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Flags:")
	flagSet.PrintDefaults()
}

// Tries to run the given command
func runCommand(name string, args []string) {
	cmdIdx := slices.IndexFunc(commands, func(cmd Command) bool {
		return cmd.Name == name
	})

	if cmdIdx < 0 {
		fmt.Fprintf(os.Stderr, "command \"%s\" not found\n\n", name)
		flag.Usage()
		os.Exit(1)
	}

	if err := commands[cmdIdx].Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}

// Run help command
func helpCmd(_ []string) error {
	flag.Usage()
	return nil
}

// Run run command
func runCmd(args []string) error {
	// Flags and help for the command
	var verbose bool
	var configFile string
	var workingDirectory string
	flagSet := flag.NewFlagSet("run", flag.ExitOnError)
	flagSet.BoolVar(&verbose, "verbose", false, "The flag to set in order to get verbose output")
	flagSet.StringVar(&configFile, "config", "gonovate.json", "The path to the config file to read")
	flagSet.StringVar(&workingDirectory, "workDir", "", "The path to the working directory")
	flagSet.Usage = func() { printCmdUsage(flagSet, "run", "") }
	flagSet.Parse(args)

	// Run the command

	// Change the working directory
	if workingDirectory != "" && workingDirectory != "." {
		if err := os.Chdir(workingDirectory); err != nil {
			return err
		}
	}
	// Read the configuration
	config, err := core.ConfigLoader{}.LoadConfig(configFile)
	if err != nil {
		return err
	}

	// Create a logger
	desiredLogLevel := lo.Ternary(verbose, slog.LevelDebug, slog.LevelInfo)
	logger := slog.New(core.NewReadableTextHandler(os.Stdout, &core.ReadableTextHandlerOptions{Level: desiredLogLevel}))

	// Prepare the platform
	platform, err := platforms.GetPlatform(logger, config)
	if err != nil {
		return err
	}

	// Process the managers
	for _, managerConfig := range config.Managers {
		manager, err := managers.GetManager(logger, config, managerConfig, platform)
		if err != nil {
			return err
		}
		// Run the manager
		if err := manager.Run(); err != nil {
			return err
		}
	}

	return nil
}

// Run debug command
func debugCmd(args []string) error {
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
	stdout, _, err := core.Git.Run("branch", "--list", "gonovate/*")
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		branchName := strings.TrimSpace(scanner.Text())
		fmt.Printf("Deleting branch '%s'\n", branchName)
		_, _, err := core.Git.Run("branch", "-D", branchName)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return err
}

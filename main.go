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
	config, err := core.ConfigLoader.LoadConfig(configFile)
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

	// Get the projects
	projects := []*core.Project{}
	isDirect := false
	if config.PlatformSettings.Direct != nil {
		isDirect = *config.PlatformSettings.Direct
	}
	if isDirect {
		// Only use the first passed project
		projects = append(projects, &core.Project{Path: config.PlatformSettings.Projects[0]})
	} else {
		// Add all projects
		for _, p := range config.PlatformSettings.Projects {
			projects = append(projects, &core.Project{Path: p})
		}
	}
	if len(projects) == 0 {
		logger.Warn("No projects found to process")
		return nil
	}

	// Process the projects
	for _, project := range projects {
		// Prepare the config for the project
		projectConfig := &core.Config{}
		projectConfig.MergeWith(config)
		// Fetch the project if needed
		oldWorkdir := ""
		if !isDirect {
			logger.Info(fmt.Sprintf("Fetching project '%s'", project.Path))
			if err = platform.FetchProject(project); err != nil {
				return err
			}
			// Change working directory to the fetched project
			oldWorkdir, err = os.Getwd()
			if err != nil {
				return err
			}
			if err := os.Chdir(".gonovate-clone"); err != nil {
				return err
			}
			// If the project has its own config file, merge it
			if hasProjectConfig, err := core.ConfigLoader.HasProjectConfig(); err != nil {
				return err
			} else if hasProjectConfig {
				projectConfigFromFile, err := core.ConfigLoader.LoadConfig("")
				if err != nil {
					return err
				}
				projectConfig.MergeWith(projectConfigFromFile)
			}
		} else {
			logger.Debug("Using direct project")
		}

		if len(projectConfig.Managers) == 0 {
			logger.Warn("No managers found to process")
			continue
		}

		// Loop thru the managers
		for _, managerConfig := range projectConfig.Managers {
			// Get the manager
			manager, err := managers.GetManager(logger, projectConfig, managerConfig)
			if err != nil {
				return err
			}

			// Get the changes from the manager
			changes, err := manager.GetChanges()
			if err != nil {
				return err
			}

			// Group and sort the changes into changesets
			// TODO: For now, each change has its own changeset
			changeSets := []*core.ChangeSet{}
			for _, change := range changes {
				meta := change.GetMeta()
				// Build the title for the changeset
				title := fmt.Sprintf("Update %s from %s to %s", meta.PackageName, meta.CurrentVersion.Raw, meta.NewRelease.Version.Raw)
				// Build the identifier for the changeset
				id := fmt.Sprintf("gonovate/%s-%s",
					core.NormalizeString(meta.PackageName, 40),
					core.NormalizeString(meta.NewRelease.Version.Raw, 0))
				// Create the changeset
				changeSets = append(changeSets, &core.ChangeSet{
					Title:   title,
					Id:      id,
					Changes: []core.IChange{change},
				})
			}

			// Special case for the noop platform: apply all changes at once
			if platform.Type() == core.PLATFORM_TYPE_NOOP {
				if err := manager.ApplyChanges(changes); err != nil {
					return err
				}
				continue
			}
			// Otherwise, loop thru the changesets
			for _, changeSet := range changeSets {
				// Prepare the platform for a new changeset
				if err := platform.PrepareForChanges(changeSet); err != nil {
					return err
				}
				// Apply the changes
				if err := manager.ApplyChanges(changeSet.Changes); err != nil {
					return err
				}
				// Submit
				if err := platform.SubmitChanges(changeSet); err != nil {
					return err
				}
				// Publish
				if err := platform.PublishChanges(changeSet); err != nil {
					return err
				}
				// Notify
				if err := platform.NotifyChanges(project, changeSet); err != nil {
					return err
				}
				// Reset
				if err := platform.ResetToBase(); err != nil {
					return err
				}
			}
		}

		// Cleanup
		if oldWorkdir != "" {
			if err := os.Chdir(oldWorkdir); err != nil {
				return err
			}
		}
		if err := os.RemoveAll(".gonovate-clone"); err != nil {
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

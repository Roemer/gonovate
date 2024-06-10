package cli

import (
	"flag"
	"fmt"
	"gonovate/core"
	"gonovate/datasources"
	"gonovate/managers"
	"gonovate/platforms"
	"log/slog"
	"os"

	"github.com/roemer/gotaskr/goext"
	"github.com/samber/lo"
)

// Run run command
func RunCmd(args []string) error {
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
	hasProject := true
	if config.PlatformSettings.Direct != nil {
		isDirect = *config.PlatformSettings.Direct
	}
	if isDirect {
		// If no project is passed, use a fake project
		if len(config.PlatformSettings.Projects) == 0 {
			hasProject = false
			projects = append(projects, &core.Project{Path: "local/local"})
		} else {
			// Use the first passed project
			projects = append(projects, &core.Project{Path: config.PlatformSettings.Projects[0]})
		}
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

		// Warn when no managers are defined but continue (to perform the cleanup)
		if len(projectConfig.Managers) == 0 {
			logger.Warn("No managers found to process")
		}

		// V2
		// Loop thru the managers and collect the dependencies
		allDependencies := []*core.Dependency{}
		for _, managerConfig := range projectConfig.Managers {
			// Build the relevant settings for thsi manager, also collect all package settings that might apply for this manager
			mergedManagerSettings, possiblePackageRules := config.FilterForManager(managerConfig)
			goext.Pass(possiblePackageRules)
			goext.Pass(hasProject)

			// Skip the manager if it is disabled
			if mergedManagerSettings.Disabled != nil && *mergedManagerSettings.Disabled {
				logger.Info(fmt.Sprintf("Manager '%s': Skip as it is disabled", managerConfig.Id))
				continue
			}

			// DEBUG
			if managerConfig.Type != core.MANAGER_TYPE_GOMOD {
				continue
			}

			// Get the manager
			manager, err := managers.GetManager2(logger, managerConfig)
			if err != nil {
				return err
			}

			// Search for the files relevant for the manager
			logger.Debug(fmt.Sprintf("Searching files with %d pattern(s)", len(mergedManagerSettings.FilePatterns)))
			candidates, err := core.SearchFiles(".", mergedManagerSettings.FilePatterns, config.IgnorePatterns)
			logger.Debug(fmt.Sprintf("Found %d matching file(s)", len(candidates)))
			if err != nil {
				return err
			}

			// Loop thru the files
			for _, candidate := range candidates {
				// Extract the dependencies for this file
				currDependencies, err := manager.ExtractDependencies(candidate)
				if err != nil {
					return err
				}
				allDependencies = append(allDependencies, currDependencies...)
			}
		}

		// Search for updates for the dependencies
		for _, dependency := range allDependencies {
			// Lookup the correct datasource
			ds, err := datasources.GetDatasource(logger, nil, dependency.Datasource)
			if err != nil {
				return err
			}

			// Search for a new version
			newReleaseInfo, currentVersion, err := ds.SearchPackageUpdate(dependency.Version, nil)
			if err != nil {
				return err
			}
			fmt.Println(currentVersion.Raw)
			fmt.Println(newReleaseInfo.Version.Raw)
		}

		// Group the dependencies which have updates according to rules
		// Loop thru the groups
		// Create the branch
		// Apply the changes
		// Commit/Submit/Notify

		// Loop thru the managers
		/*for _, managerConfig := range projectConfig.Managers {
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
				branchName := fmt.Sprintf("%s%s-%s-%s",
					config.PlatformSettings.BranchPrefix,
					core.NormalizeString(config.PlatformSettings.BaseBranch, 20),
					core.NormalizeString(meta.PackageName, 40),
					core.NormalizeString(meta.NewRelease.Version.Raw, 0))
				// Create the changeset
				changeSets = append(changeSets, &core.ChangeSet{
					Title:      title,
					BranchName: branchName,
					Changes:    []core.IChange{change},
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
				if hasProject {
					// Only notify if a project was defined, otherwise we do not know where to notify
					if err := platform.NotifyChanges(project, changeSet); err != nil {
						return err
					}
				}
				// Reset
				if err := platform.ResetToBase(); err != nil {
					return err
				}
			}
		}*/

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

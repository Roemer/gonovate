package gonovate

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"slices"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/datasources"
	"github.com/roemer/gonovate/internal/pkg/managers"
	"github.com/roemer/gonovate/internal/pkg/platforms"
	"github.com/roemer/gonovate/internal/pkg/shared"
	"github.com/roemer/gonovate/pkg/logging"
	"github.com/samber/lo"
)

func RunCmd(args []string) error {
	// Flags and help for the command
	var verbose bool
	var configFile string
	var workingDirectory string
	var platformOverride string
	flagSet := flag.NewFlagSet("run", flag.ExitOnError)
	flagSet.BoolVar(&verbose, "verbose", false, "The flag to set in order to get verbose output")
	flagSet.BoolVar(&verbose, "v", verbose, "Alias for -verbose")
	flagSet.StringVar(&configFile, "config", "gonovate", "The path to the config file to read")
	flagSet.StringVar(&workingDirectory, "workDir", "", "The path to the working directory")
	flagSet.StringVar(&platformOverride, "platform", "", "Allows overriding the platform. Usefull for testing when setting to 'noop'.")
	flagSet.Usage = func() { printCmdUsage(flagSet, "run", "") }
	flagSet.Parse(args)

	// Create a logger
	desiredLogLevel := lo.Ternary(verbose, slog.LevelDebug, slog.LevelInfo)
	logger := slog.New(logging.NewReadableTextHandler(os.Stdout, &logging.ReadableTextHandlerOptions{Level: desiredLogLevel}))
	logger.Debug(fmt.Sprintf("Initialized logger with level: %s", desiredLogLevel))
	logger.Info("Starting gonovate run")

	// Change the working directory
	if workingDirectory != "" && workingDirectory != "." {
		logger.Debug(fmt.Sprintf("Changing working directory to: %s", workingDirectory))
		if err := os.Chdir(workingDirectory); err != nil {
			return err
		}
	}

	// Read the configuration
	rootConfig, err := config.Load(configFile)
	if err != nil {
		return err
	}

	// Process overrides
	if platformOverride != "" {
		rootConfig.Platform = shared.PlatformType(platformOverride)
	}

	// Prepare the platform
	platform, err := platforms.GetPlatform(logger, rootConfig)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Prepared platform: %s", platform.Type()))

	// Get the projects
	projects := []*shared.Project{}
	isInplace := false
	hasProject := true
	if rootConfig.PlatformSettings.Inplace != nil {
		isInplace = *rootConfig.PlatformSettings.Inplace
	}
	if isInplace {
		// If no project is passed, use a fake project
		if len(rootConfig.PlatformSettings.Projects) == 0 {
			hasProject = false
			projects = append(projects, &shared.Project{Path: "local/local"})
		} else {
			// Use the first passed project
			projects = append(projects, &shared.Project{Path: rootConfig.PlatformSettings.Projects[0]})
		}
	} else {
		// Add all projects
		for _, p := range rootConfig.PlatformSettings.Projects {
			projects = append(projects, &shared.Project{Path: p})
		}
	}
	if len(projects) == 0 {
		logger.Warn("No projects found to process")
		return nil
	}
	logger.Info(fmt.Sprintf("Processing %s", shared.GetSingularPluralStringSimple(projects, "project")))

	// Process the projects
	for _, project := range projects {
		logger.Info(fmt.Sprintf("Processing project '%s'", project.Path))
		// Prepare the config for the project
		projectConfig := &config.RootConfig{}
		projectConfig.MergeWith(rootConfig)
		// Fetch the project if needed
		oldWorkdir := ""
		if !isInplace {
			logger.Info("Fetching project from platform")
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
			if foundPath, err := config.HasProjectConfig(); err != nil {
				return err
			} else if foundPath != "" {
				projectConfigFromFile, err := config.Load(foundPath)
				if err != nil {
					return err
				}
				projectConfig.MergeWith(projectConfigFromFile)
			}
		} else {
			logger.Debug("Using inplace project")
		}

		// Warn when no managers are defined but continue (to perform the cleanup)
		if len(projectConfig.Managers) == 0 {
			logger.Warn("No managers found to process")
		}

		// Loop thru the managers and collect the dependencies
		allDependencies := []*shared.Dependency{}
		logger.Info(fmt.Sprintf("Searching for dependencies in %s", shared.GetSingularPluralStringSimple(projectConfig.Managers, "manager")))
		for _, managerConfig := range projectConfig.Managers {
			// Get the merged settings for the current manager
			mergedManagerSettings := projectConfig.GetMergedManagerSettings(managerConfig)

			// Skip the manager if it is disabled
			if mergedManagerSettings.Disabled != nil && *mergedManagerSettings.Disabled {
				logger.Info(fmt.Sprintf("Manager '%s': Skip as it is disabled", managerConfig.Id))
				continue
			}
			logger.Info(fmt.Sprintf("Processing Manager '%s' (%s)", managerConfig.Id, managerConfig.Type))

			// Get the manager
			manager, err := managers.GetManager(logger, projectConfig, managerConfig, mergedManagerSettings)
			if err != nil {
				return err
			}

			// Search for the files relevant for the manager
			logger.Debug(fmt.Sprintf("Searching files with %s", shared.GetSingularPluralStringSimple(mergedManagerSettings.FilePatterns, "pattern")))
			matchingFiles, err := shared.SearchFiles(".", mergedManagerSettings.FilePatterns, rootConfig.IgnorePatterns)
			logger.Debug(fmt.Sprintf("Found %s", shared.GetSingularPluralStringSimple(matchingFiles, "matching file")))
			if err != nil {
				return err
			}

			// Loop thru the files and collect the dependencies
			dependenciesInManager := []*shared.Dependency{}
			for _, matchingFile := range matchingFiles {
				logger.Debug(fmt.Sprintf("Processing file '%s'", matchingFile))
				// Extract the dependencies for this file
				currDependencies, err := manager.ExtractDependencies(matchingFile)
				if err != nil {
					return err
				}
				// Set some generic fields for all just found dependencies
				for _, dependency := range currDependencies {
					dependency.ManagerId = manager.Id()
					dependency.FilePath = matchingFile
				}
				logger.Debug(fmt.Sprintf("Found %s in file", shared.GetDependencyString(currDependencies)))
				dependenciesInManager = append(dependenciesInManager, currDependencies...)
			}
			// Add all dependencies
			logger.Info(fmt.Sprintf("Found %s in manager", shared.GetDependencyString(dependenciesInManager)))
			allDependencies = append(allDependencies, dependenciesInManager...)
		}
		logger.Info(fmt.Sprintf("Found %s in total", shared.GetDependencyString(allDependencies)))

		// Search for updates for the dependencies
		logger.Info("Searching for dependency updates")
		dependenciesWithUpdates := []*shared.Dependency{}
		for _, dependency := range allDependencies {
			logger.Info(fmt.Sprintf("Processing dependency '%s' (%s) from %s with version %s", dependency.Name, dependency.Datasource, dependency.ManagerId, dependency.Version))
			// Enrich the dependency with settings from the config/rules
			projectConfig.EnrichDependencyFromRules(dependency)

			// Skip the dependency if it was disabled
			if dependency.Skip != nil && *dependency.Skip {
				reason := ""
				if dependency.SkipReason != "" {
					reason = " Reason: " + dependency.SkipReason
				}
				logger.Info(fmt.Sprintf("Skipping dependency.%s", reason))
				continue
			}

			// Lookup the correct datasource
			ds, err := datasources.GetDatasource(logger, projectConfig, dependency.Datasource)
			if err != nil {
				return err
			}

			// Search for a new version
			newReleaseInfo, _, err := ds.SearchDependencyUpdate(dependency)
			if err != nil {
				return err
			}
			if newReleaseInfo != nil {
				dependency.NewRelease = newReleaseInfo
				dependenciesWithUpdates = append(dependenciesWithUpdates, dependency)
			}
		}
		logger.Info(fmt.Sprintf("Found %s with updates", shared.GetDependencyString(dependenciesWithUpdates)))

		// Group the dependencies which have updates according to group names
		updateGroups := []*shared.UpdateGroup{}
		for _, dependency := range dependenciesWithUpdates {
			var title, branchName string
			if dependency.GroupName != "" {
				title = fmt.Sprintf("Update group '%s'", dependency.GroupName)
				branchName = fmt.Sprintf("%s%s",
					projectConfig.PlatformSettings.BranchPrefix,
					dependency.GroupName)
			} else {
				title = fmt.Sprintf("Update '%s' to '%s'", dependency.Name, dependency.NewRelease.VersionString)
				branchName = fmt.Sprintf("%s%s-%s-%s",
					projectConfig.PlatformSettings.BranchPrefix,
					shared.NormalizeString(projectConfig.PlatformSettings.BaseBranch, 20),
					shared.NormalizeString(dependency.Name, 40),
					shared.NormalizeString(dependency.NewRelease.VersionString, 0))
			}

			// Check if such a group already exists
			idx := slices.IndexFunc(updateGroups, func(g *shared.UpdateGroup) bool { return g.BranchName == branchName })
			if idx >= 0 {
				// It does, so just add the dependency to the existing group
				updateGroups[idx].Dependencies = append(updateGroups[idx].Dependencies, dependency)
			} else {
				// Create the group
				newGroup := &shared.UpdateGroup{
					Title:        title,
					BranchName:   branchName,
					Dependencies: []*shared.Dependency{dependency},
				}
				updateGroups = append(updateGroups, newGroup)
			}
		}
		logger.Info(fmt.Sprintf("Created %s with dependency updates", shared.GetSingularPluralStringSimple(updateGroups, "group")))

		// Loop thru the groups
		for _, updateGroup := range updateGroups {
			logger.Info(fmt.Sprintf("Processing group '%s' with %s", updateGroup.Title, shared.GetDependencyString(updateGroup.Dependencies)))

			// Prepare the platform for a new changeset
			logger.Debug("Prepaparing for changes")
			if err := platform.PrepareForChanges(updateGroup); err != nil {
				return err
			}

			// Apply the changes
			for _, dependency := range updateGroup.Dependencies {
				logger.Info(fmt.Sprintf("Updating '%s' from '%s' to '%s'", dependency.Name, dependency.Version, dependency.NewRelease.VersionString))
				managerConfig := projectConfig.GetManagerConfigById(dependency.ManagerId)
				// Get the merged settings for the current manager
				mergedManagerSettings := projectConfig.GetMergedManagerSettings(managerConfig)

				// Get the manager
				manager, err := managers.GetManager(logger, projectConfig, managerConfig, mergedManagerSettings)
				if err != nil {
					return err
				}
				if err := manager.ApplyDependencyUpdate(dependency); err != nil {
					return err
				}

				// Run Post-Upgrade replacements
				hasPostUpgradeReplacements := len(dependency.PostUpgradeReplacements) > 0
				if hasPostUpgradeReplacements {
					// Read the file
					fileContentBytes, err := os.ReadFile(dependency.FilePath)
					if err != nil {
						return err
					}
					fileContent := string(fileContentBytes)
					// Apply the replacements
					for _, reStr := range dependency.PostUpgradeReplacements {
						re := regexp.MustCompile(reStr)
						fileContent, _ = shared.ReplaceMatchesInRegex(re, fileContent, map[string]string{
							"version": dependency.NewRelease.Version.Raw,
							"sha1":    dependency.NewRelease.AdditionalData["sha1"],
							"sha256":  dependency.NewRelease.AdditionalData["sha256"],
							"sha512":  dependency.NewRelease.AdditionalData["sha512"],
							"md5":     dependency.NewRelease.AdditionalData["md5"],
						})
					}
					// Write the file with the changes
					if err := os.WriteFile(dependency.FilePath, []byte(fileContent), os.ModePerm); err != nil {
						return err
					}
				}
			}

			// Submit
			logger.Debug("Submitting the changes")
			if err := platform.SubmitChanges(updateGroup); err != nil {
				return err
			}

			// Publish
			logger.Debug("Publishing the changes")
			if err := platform.PublishChanges(updateGroup); err != nil {
				return err
			}

			// Notify
			if hasProject {
				// Only notify if a project was defined, otherwise we do not know where to notify
				logger.Debug("Notifying the project about the changes")
				if err := platform.NotifyChanges(project, updateGroup); err != nil {
					return err
				}
			}

			// Reset
			logger.Debug("Resetting to the base branch")
			if err := platform.ResetToBase(); err != nil {
				return err
			}
		}

		// Cleanup the platform (eg. unused PRs/MRs)
		if err := platform.Cleanup(&platforms.PlatformCleanupSettings{
			Project:      project,
			UpdateGroups: updateGroups,
			BaseBranch:   projectConfig.PlatformSettings.BaseBranch,
			BranchPrefix: projectConfig.PlatformSettings.BranchPrefix,
		}); err != nil {
			return err
		}

		// Cleanup the working directory
		if oldWorkdir != "" {
			if err := os.Chdir(oldWorkdir); err != nil {
				return err
			}
		}
		if err := os.RemoveAll(".gonovate-clone"); err != nil {
			return err
		}
	}

	logger.Info("Gonovate finished successfully")

	return nil
}

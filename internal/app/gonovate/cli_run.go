package gonovate

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/roemer/gonovate/pkg/config"
	"github.com/roemer/gonovate/pkg/datasources"
	"github.com/roemer/gonovate/pkg/logging"
	"github.com/roemer/gonovate/pkg/managers"
	"github.com/roemer/gonovate/pkg/platforms"
	"github.com/samber/lo"
)

func RunCmd(args []string) error {
	// Flags and help for the command
	var verbose bool
	var configFiles stringSliceFlag
	var workingDirectory string
	var platformOverride string
	var exclusive string
	flagSet := flag.NewFlagSet("run", flag.ExitOnError)
	flagSet.BoolVar(&verbose, "verbose", false, "The flag to set in order to get verbose output.")
	flagSet.BoolVar(&verbose, "v", verbose, "Alias for -verbose.")
	flagSet.Var(&configFiles, "config", "The path to the config file to read. Can be passed multiple times.")
	flagSet.StringVar(&workingDirectory, "workDir", "", "The path to the working directory.")
	flagSet.StringVar(&platformOverride, "platform", "", "Allows overriding the platform. Usefull for testing when setting to 'noop'.")
	flagSet.StringVar(&exclusive, "exclusive", "", "Allows defining criterias for exclusive updating. The format is: key1=value1|key2=value2\nValid Keys are: dependency, datasource, file, manager, managerType")
	flagSet.StringVar(&exclusive, "e", exclusive, "Alias for -exclusive.")
	flagSet.Usage = func() { printCmdUsage(flagSet, "run", "") }
	flagSet.Parse(args)

	// Create a logger
	desiredLogLevel := lo.Ternary(verbose, slog.LevelDebug, slog.LevelInfo)
	logger := slog.New(logging.NewReadableTextHandler(os.Stdout, &logging.ReadableTextHandlerOptions{Level: desiredLogLevel}))
	logger.Debug(fmt.Sprintf("Initialized logger with level: %s", desiredLogLevel))
	logger.Info("Starting gonovate run")

	// Parse the exclusive flag
	topPriorityRules := []*config.Rule{}
	if exclusive != "" {
		// Rule that disables all managers and skips all dependencies
		exclusiveRule := &config.Rule{
			ManagerConfig:    &config.ManagerConfig{},
			DependencyConfig: &config.DependencyConfig{},
		}
		// Rule that enables the desired manager or dependency
		inclusiveRule := &config.Rule{
			Matches:          &config.RuleMatch{},
			ManagerConfig:    &config.ManagerConfig{Disabled: common.FalsePtr},
			DependencyConfig: &config.DependencyConfig{Skip: common.FalsePtr},
		}
		// Check the given values and assign them appropriate match
		hasManagerExclusive := false
		hasDependencyExclusive := false
		pairs := strings.Split(exclusive, "|")
		for _, pair := range pairs {
			values := strings.SplitN(pair, "=", 2)
			if len(values) < 2 {
				continue
			}
			key := values[0]
			value := strings.TrimSpace(values[1])
			if value == "" {
				// Skip empty values
				continue
			}
			switch strings.ToLower(key) {
			case "dependency":
				hasDependencyExclusive = true
				inclusiveRule.Matches.DependencyNames = []string{value}
			case "datasource":
				hasDependencyExclusive = true
				inclusiveRule.Matches.Datasources = []common.DatasourceType{common.DatasourceType(value)}
			case "file":
				hasDependencyExclusive = true
				inclusiveRule.Matches.Files = []string{value}
			case "manager":
				hasManagerExclusive = true
				inclusiveRule.Matches.Managers = []string{value}
			case "managertype":
				hasManagerExclusive = true
				inclusiveRule.Matches.ManagerTypes = []common.ManagerType{common.ManagerType(value)}
			}

		}
		// Make sure at least one value matched
		if hasManagerExclusive || hasDependencyExclusive {
			if hasManagerExclusive {
				// There is a rule that enables a specific manager, so disable all others
				exclusiveRule.ManagerConfig.Disabled = common.TruePtr
			}
			if hasDependencyExclusive {
				// There is a rule that enables a specific dependency, so skip all others
				exclusiveRule.DependencyConfig.Skip = common.TruePtr
				exclusiveRule.DependencyConfig.SkipReason = "Exclusive"
			}
			topPriorityRules = append(topPriorityRules, exclusiveRule)
			topPriorityRules = append(topPriorityRules, inclusiveRule)
		} else {
			logger.Warn(fmt.Sprintf("Exclusive flag passed but with incompatible values: %s", exclusive))
		}
	}

	// Change the working directory
	if workingDirectory != "" && workingDirectory != "." {
		logger.Debug(fmt.Sprintf("Changing working directory to: %s", workingDirectory))
		if err := os.Chdir(workingDirectory); err != nil {
			return err
		}
	}

	// Read the main configuration
	mainConfig := ""
	if len(configFiles) > 0 {
		mainConfig = configFiles[0]
	}
	gonovateConfig, err := config.Load(mainConfig)
	if err != nil {
		return err
	}

	// Merge additional config files
	if len(configFiles) > 1 {
		for _, configFile := range configFiles[1:] {
			additionalConfig, err := config.Load(configFile)
			if err != nil {
				return err
			}
			gonovateConfig.MergeWith(additionalConfig)
		}
	}

	// Process overrides
	if platformOverride != "" {
		if gonovateConfig.Platform == nil {
			gonovateConfig.Platform = &config.PlatformConfig{}
		}
		gonovateConfig.Platform.Type = common.PlatformType(platformOverride)
	}

	// Prepare the platform
	platformSettings := gonovateConfig.ToCommonPlatformSettings(logger)
	platform, err := platforms.GetPlatform(platformSettings)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Prepared platform: %s", platform.Type()))

	// Get the projects
	projects := []*common.Project{}
	isInplace := false
	hasProject := true
	if gonovateConfig.Platform.Inplace != nil {
		isInplace = *gonovateConfig.Platform.Inplace
	}
	if isInplace {
		// If no project is passed, use a fake project
		if len(gonovateConfig.Platform.Projects) == 0 {
			hasProject = false
			projects = append(projects, &common.Project{Path: "local/local"})
		} else {
			// Use the first passed project
			projects = append(projects, &common.Project{Path: gonovateConfig.Platform.Projects[0]})
		}
	} else {
		// Add all projects
		for _, p := range gonovateConfig.Platform.Projects {
			projects = append(projects, &common.Project{Path: p})
		}
	}
	if len(projects) == 0 {
		logger.Warn("No projects found to process")
		return nil
	}
	logger.Info(fmt.Sprintf("Processing %d project(s)", len(projects)))

	// Process the projects
	for _, project := range projects {
		logger.Info(fmt.Sprintf("Processing project '%s'", project.Path))
		// Prepare the config for the project
		projectConfig := &config.GonovateConfig{}
		projectConfig.MergeWith(gonovateConfig)
		// Also add all the top priority rules (from exclusive) at the end
		projectConfig.Rules = append(projectConfig.Rules, topPriorityRules...)
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
		allDependencies := []*common.Dependency{}
		logger.Info(fmt.Sprintf("Searching for dependencies in %d manager(s)", len(projectConfig.Managers)))
		for _, manager := range projectConfig.Managers {
			// Get the merged configs for the current manager
			mergedManagerConfig := projectConfig.GetMergedManagerConfig(manager)

			// Skip the manager if it is disabled
			if mergedManagerConfig.Disabled != nil && *mergedManagerConfig.Disabled {
				logger.Info(fmt.Sprintf("Manager '%s': Skip as it is disabled", manager.Id))
				continue
			}
			logger.Info(fmt.Sprintf("Processing Manager '%s' (%s)", manager.Id, manager.Type))

			// Get the manager
			managerSettings := &common.ManagerSettings{
				Logger:      logger,
				Id:          manager.Id,
				ManagerType: manager.Type,
				RegexManagerSettings: &common.RegexManagerSettings{
					MatchStringPresets: projectConfig.MatchStringPresetsToPresets(),
					MatchStrings:       mergedManagerConfig.MatchStrings,
				},
				DevcontainerManagerSettings: mergedManagerConfig.ToCommonDevcontainerManagerSettings(),
			}
			managerInstance, err := managers.GetManager(managerSettings)
			if err != nil {
				return err
			}

			// Search for the files relevant for the manager
			logger.Debug(fmt.Sprintf("Searching files with %d pattern(s)", len(mergedManagerConfig.FilePatterns)))
			matchingFiles, err := common.SearchFiles(".", mergedManagerConfig.FilePatterns, gonovateConfig.IgnorePatterns)
			logger.Debug(fmt.Sprintf("Found %d matching file(s)", len(matchingFiles)))
			if err != nil {
				return err
			}

			// Loop thru the files and collect the dependencies
			dependenciesInManager := []*common.Dependency{}
			for _, matchingFile := range matchingFiles {
				logger.Debug(fmt.Sprintf("Processing file '%s'", matchingFile))
				// Extract the dependencies for this file
				currDependencies, err := managerInstance.ExtractDependencies(matchingFile)
				if err != nil {
					return err
				}
				logger.Debug(fmt.Sprintf("Found %d dependencies in file", len(currDependencies)))
				dependenciesInManager = append(dependenciesInManager, currDependencies...)
			}
			// Add all dependencies
			logger.Info(fmt.Sprintf("Found %d dependencies in manager", len(dependenciesInManager)))
			allDependencies = append(allDependencies, dependenciesInManager...)
		}
		logger.Info(fmt.Sprintf("Found %d dependencies in total", len(allDependencies)))

		// Search for updates for the dependencies
		logger.Info("Searching for dependency updates")
		dependenciesWithUpdates := []*common.Dependency{}
		for _, dependency := range allDependencies {
			logger.Info(fmt.Sprintf("Processing dependency '%s' (%s) from %s with version %s", dependency.Name, dependency.Datasource, dependency.ManagerInfo.ManagerId, dependency.Version))
			// Apply the config to the dependency
			if err := projectConfig.ApplyToDependency(dependency); err != nil {
				return err
			}

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
			datasourceSettings := &common.DatasourceSettings{
				Logger:         logger,
				DatasourceType: dependency.Datasource,
				HostRules:      projectConfig.HostRules,
			}
			ds, err := datasources.GetDatasource(datasourceSettings)
			if err != nil {
				return err
			}

			// Search for a new version
			newReleaseInfo, err := ds.SearchDependencyUpdate(dependency)
			if err != nil {
				return err
			}
			if newReleaseInfo != nil {
				dependency.NewRelease = newReleaseInfo
				dependenciesWithUpdates = append(dependenciesWithUpdates, dependency)
			}
		}
		logger.Info(fmt.Sprintf("Found %d dependencies with updates", len(dependenciesWithUpdates)))

		// Group the dependencies which have updates according to group names
		updateGroups := []*common.UpdateGroup{}
		for _, dependency := range dependenciesWithUpdates {
			var title, branchName string
			if dependency.GroupName != "" {
				title = fmt.Sprintf("Update group '%s'", dependency.GroupName)
				branchName = fmt.Sprintf("%s%s",
					projectConfig.Platform.BranchPrefix,
					dependency.GroupName)
			} else {
				title = fmt.Sprintf("Update '%s' to '%s'", dependency.Name, dependency.NewRelease.VersionString)
				branchName = fmt.Sprintf("%s%s-%s-%s",
					projectConfig.Platform.BranchPrefix,
					common.NormalizeString(projectConfig.Platform.BaseBranch, 20),
					common.NormalizeString(dependency.Name, 40),
					common.NormalizeString(dependency.NewRelease.VersionString, 0))
			}

			// Check if such a group already exists
			idx := slices.IndexFunc(updateGroups, func(g *common.UpdateGroup) bool { return g.BranchName == branchName })
			if idx >= 0 {
				// It does, so just add the dependency to the existing group
				updateGroups[idx].Dependencies = append(updateGroups[idx].Dependencies, dependency)
			} else {
				// Create the group
				newGroup := &common.UpdateGroup{
					Title:        title,
					BranchName:   branchName,
					Dependencies: []*common.Dependency{dependency},
				}
				updateGroups = append(updateGroups, newGroup)
			}
		}
		logger.Info(fmt.Sprintf("Created %d group(s) with dependency updates", len(updateGroups)))

		// Loop thru the groups
		for _, updateGroup := range updateGroups {
			logger.Info(fmt.Sprintf("Processing group '%s' with %d dependencies", updateGroup.Title, len(updateGroup.Dependencies)))

			// Prepare the platform for a new changeset
			logger.Debug("Prepaparing for changes")
			if err := platform.PrepareForChanges(updateGroup); err != nil {
				return err
			}

			// Apply the changes
			for _, dependency := range updateGroup.Dependencies {
				logger.Info(fmt.Sprintf("Updating '%s' from '%s' to '%s'", dependency.Name, dependency.Version, dependency.NewRelease.VersionString))
				manager := projectConfig.GetManagerById(dependency.ManagerInfo.ManagerId)
				// Get the merged configs for the current manager
				mergedManagerConfig := projectConfig.GetMergedManagerConfig(manager)

				// Get the manager
				managerSettings := &common.ManagerSettings{
					Logger:      logger,
					Id:          manager.Id,
					ManagerType: manager.Type,
					RegexManagerSettings: &common.RegexManagerSettings{
						MatchStringPresets: projectConfig.MatchStringPresetsToPresets(),
						MatchStrings:       mergedManagerConfig.MatchStrings,
					},
					DevcontainerManagerSettings: mergedManagerConfig.ToCommonDevcontainerManagerSettings(),
				}
				managerInstance, err := managers.GetManager(managerSettings)
				if err != nil {
					return err
				}
				if err := managerInstance.ApplyDependencyUpdate(dependency); err != nil {
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
						fileContent, _ = common.ReplaceMatchesInRegex(re, fileContent, map[string]string{
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
			BaseBranch:   projectConfig.Platform.BaseBranch,
			BranchPrefix: projectConfig.Platform.BranchPrefix,
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

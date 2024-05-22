package main

import (
	"flag"
	"fmt"
	"gonovate/core"
	"gonovate/managers"
	"gonovate/platforms"
	"log/slog"
	"os"
)

type processSettings struct {
	configFile       string
	workingDirectory string
}

func main() {
	// CLI flags
	help := flag.Bool("help", false, "The flag to set in order to display the help")
	configFile := flag.String("config", "gonovate.json", "The path to the config file to read")
	workingDirectory := flag.String("workDir", "", "The path to the working directory")
	flag.Parse()

	// Show help
	if *help {
		fmt.Println("Usage: gonovate <flags>")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Process
	err := process(processSettings{
		configFile:       *configFile,
		workingDirectory: *workingDirectory,
	})
	if err != nil {
		panic(err)
	}
}

func process(processSettings processSettings) error {
	// Change the working directory
	if processSettings.workingDirectory != "" {
		if err := os.Chdir(processSettings.workingDirectory); err != nil {
			return err
		}
	}
	// Read the configuration
	config, err := core.ReadConfig(processSettings.configFile)
	if err != nil {
		return err
	}

	fmt.Println(config)

	// Create a logger
	logger := slog.New(core.NewReadableTextHandler(os.Stdout, &core.ReadableTextHandlerOptions{Level: slog.LevelDebug}))

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

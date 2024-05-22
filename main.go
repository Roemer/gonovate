package main

import (
	"fmt"
	"gonovate/core"
	"gonovate/managers"
	"gonovate/platforms"
	"log/slog"
	"os"
)

func main() {
	err := process()
	if err != nil {
		panic(err)
	}
}

func process() error {
	config, err := core.ReadConfig("gonovate.json")
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

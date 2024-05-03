package main

import (
	"fmt"
	"gonovate/core"
	"gonovate/managers"
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
	config, err := core.ReadConfig("config_examples/gonovate.json")
	if err != nil {
		return err
	}

	fmt.Println(config)

	// Create a logger
	logger := slog.New(core.NewReadableTextHandler(os.Stdout, &core.ReadableTextHandlerOptions{Level: slog.LevelDebug}))

	for _, manager := range config.Managers {
		// Handle the different manager types
		switch manager.Type {
		case core.MANAGER_TYPE_REGEX:
			managerInstance := managers.NewRegexManager(logger, config, manager)
			if err := managerInstance.Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func ReadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "gonovate.json"
	}
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening config file '%s': %w", configPath, err)
	}
	defer configFile.Close()

	config := &Config{}
	if err = json.NewDecoder(configFile).Decode(config); err != nil {
		return nil, fmt.Errorf("failed parsing config file '%s': %w", configPath, err)
	}

	// Create a new object for the final merged config
	finalConfig := &Config{}
	// Process the "Extends" presets first
	for _, presetLookupInfo := range config.Extends {
		// Read the extended preset
		extendsConfig, err := readExtendsConfig(configPath, presetLookupInfo)
		if err != nil {
			return nil, err
		}
		// Merge the extended preset into the final config
		finalConfig.MergeWith(extendsConfig)
	}
	// Merge the original config into the final config
	finalConfig.MergeWith(config)
	// Return the final config
	return finalConfig, nil
}

func readExtendsConfig(currentPath, presetLookupInfo string) (*Config, error) {
	// Presets without a type default to file
	if !strings.Contains(presetLookupInfo, ":") {
		presetLookupInfo = "file:" + presetLookupInfo
	}
	parts := strings.Split(presetLookupInfo, ":")
	presetType := parts[0]
	presetPath := parts[1]
	// Process file preset
	if presetType == "file" {
		if path.Ext(presetPath) == "" {
			presetPath += ".json"
		}
		// Try resolving the preset path
		var err error
		presetPath, err = getPresetPath(currentPath, presetPath)
		if err != nil {
			return nil, err
		}
	}
	// Read the config from the path
	return ReadConfig(presetPath)
}

func getPresetPath(currentPath string, presetPath string) (string, error) {
	// If the path is absolute, use it if it exists
	if filepath.IsAbs(presetPath) {
		if exists, err := FileExists(presetPath); err != nil {
			return "", err
		} else if exists {
			return presetPath, nil
		}
	} else {
		// Search it based on the current config directory
		tempPresetPath := filepath.Clean(filepath.Join(filepath.Dir(currentPath), presetPath))
		if exists, err := FileExists(tempPresetPath); err != nil {
			return "", err
		} else if exists {
			return tempPresetPath, nil
		}
		// Search it based on the current executable directory
		executablePath, err := os.Executable()
		if err != nil {
			return "", err
		}
		tempPresetPath = filepath.Clean(filepath.Join(filepath.Dir(executablePath), presetPath))
		if exists, err := FileExists(tempPresetPath); err != nil {
			return "", err
		} else if exists {
			return tempPresetPath, nil
		}
		// Search based on the current executable directory but in the presets subfolder
		tempPresetPath = filepath.Clean(filepath.Join(filepath.Dir(executablePath), "presets", presetPath))
		if exists, err := FileExists(tempPresetPath); err != nil {
			return "", err
		} else if exists {
			return tempPresetPath, nil
		}
		// Search based on the current file that is executed (probably dev with go run . only)
		if _, filename, _, ok := runtime.Caller(0); ok {
			rootPath := filepath.Dir(filepath.Dir(filename))
			tempPresetPath = filepath.Clean(filepath.Join(rootPath, "presets", presetPath))
			if exists, err := FileExists(tempPresetPath); err != nil {
				return "", err
			} else if exists {
				return tempPresetPath, nil
			}
		}
	}
	return "", fmt.Errorf("preset file not found for '%s'", presetPath)
}

func Ptr[T any](value T) *T {
	return &value
}

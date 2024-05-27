package core

import (
	"encoding/json"
	"fmt"
	"gonovate/presets"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	infoTypeFile string = "file"
)

type configLoader struct{}

var ConfigLoader configLoader = configLoader{}

type configInfo struct {
	Type     string
	Location string
}

func newConfigInfo(info string) (*configInfo, error) {
	if info == "" {
		return nil, fmt.Errorf("empty config info")
	}
	parts := strings.SplitN(info, ":", 2)
	var configType, configLoc string
	if len(parts) == 1 {
		configType = infoTypeFile
		configLoc = parts[0]
	} else {
		configType = parts[0]
		configLoc = parts[1]
	}
	// Append the json extension if needed
	if path.Ext(configLoc) == "" {
		configLoc += ".json"
	}
	// Create the info object
	return &configInfo{
		Type:     configType,
		Location: configLoc,
	}, nil
}

func (c configLoader) HasProjectConfig() (bool, error) {
	return FileExists("gonovate.json")
}

func (c configLoader) LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "gonovate.json"
	}
	configInfo, err := newConfigInfo(configPath)
	if err != nil {
		return nil, err
	}
	return c.loadConfig(nil, configInfo)
}

func (c configLoader) loadConfig(parentInfo, newInfo *configInfo) (*Config, error) {
	var newConfig *Config
	var err error
	// Try load the config according to the type
	if newInfo.Type == infoTypeFile {
		newConfig, err = c.loadConfigFromFile(parentInfo, newInfo)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unknown preset type '%s'", newInfo.Type)
	}

	// Create a new object for the merged config with the presets
	mergedConfig := &Config{}
	// Process the "Extends" presets first
	for _, presetLookupInfo := range newConfig.Extends {
		presetInfo, err := newConfigInfo(presetLookupInfo)
		if err != nil {
			return nil, err
		}
		// Read the extended preset
		extendsConfig, err := c.loadConfig(newInfo, presetInfo)
		if err != nil {
			return nil, err
		}
		// Merge the extended preset into the merged config
		mergedConfig.MergeWith(extendsConfig)
	}
	// Merge the original config into the merged config
	mergedConfig.MergeWith(newConfig)

	// Return the merged config
	return mergedConfig, nil
}

func (c configLoader) loadConfigFromFile(parentInfo, newInfo *configInfo) (*Config, error) {
	// If the path is absolute, use it directly
	if filepath.IsAbs(newInfo.Location) {
		return c.readConfigFromFile(newInfo.Location)
	}

	// A: Try load it from the current folder
	if exists, err := FileExists(newInfo.Location); err != nil {
		return nil, err
	} else if exists {
		return c.readConfigFromFile(newInfo.Location)
	}

	// B: Search in the folder of the parent config
	if parentInfo != nil && parentInfo.Type == infoTypeFile && parentInfo.Location != "" {
		tempPresetPath := filepath.Clean(filepath.Join(filepath.Dir(parentInfo.Location), newInfo.Location))
		if exists, err := FileExists(tempPresetPath); err != nil {
			return nil, err
		} else if exists {
			return c.readConfigFromFile(tempPresetPath)
		}
	}

	// C: Search in the current executable directory
	if executablePath, err := os.Executable(); err != nil {
		return nil, err
	} else {
		tempPresetPath := filepath.Clean(filepath.Join(filepath.Dir(executablePath), newInfo.Location))
		if exists, err := FileExists(tempPresetPath); err != nil {
			return nil, err
		} else if exists {
			return c.readConfigFromFile(tempPresetPath)
		}
		// Also search in the presets subfolder
		tempPresetPath = filepath.Clean(filepath.Join(filepath.Dir(executablePath), "presets", newInfo.Location))
		if exists, err := FileExists(tempPresetPath); err != nil {
			return nil, err
		} else if exists {
			return c.readConfigFromFile(tempPresetPath)
		}
	}

	// D: Search based on the current file that is executed (probably dev with go run . only)
	if _, filename, _, ok := runtime.Caller(0); ok {
		rootPath := filepath.Dir(filepath.Dir(filename))
		tempPresetPath := filepath.Clean(filepath.Join(rootPath, newInfo.Location))
		if exists, err := FileExists(tempPresetPath); err != nil {
			return nil, err
		} else if exists {
			return c.readConfigFromFile(tempPresetPath)
		}
		// Also search in the presets subfolder
		tempPresetPath = filepath.Clean(filepath.Join(rootPath, "presets", newInfo.Location))
		if exists, err := FileExists(tempPresetPath); err != nil {
			return nil, err
		} else if exists {
			return c.readConfigFromFile(tempPresetPath)
		}
	}

	// E: Search from the embedded presets
	hasEmbedded := false
	if err := fs.WalkDir(presets.Presets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if path == newInfo.Location {
				hasEmbedded = true
				return fs.SkipAll
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if hasEmbedded {
		return c.readConfigFromEmbeddedFile(newInfo.Location)
	}

	// Nothing found at all
	return nil, fmt.Errorf("file not found for '%s'", newInfo.Location)
}

func (c configLoader) readConfigFromFile(configPath string) (*Config, error) {
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening file '%s': %w", configPath, err)
	}
	defer configFile.Close()

	config := &Config{}
	if err = json.NewDecoder(configFile).Decode(config); err != nil {
		return nil, fmt.Errorf("failed parsing file '%s': %w", configPath, err)
	}
	return config, nil
}

func (c configLoader) readConfigFromEmbeddedFile(configPath string) (*Config, error) {
	configFile, err := presets.Presets.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening embedded file '%s': %w", configPath, err)
	}
	defer configFile.Close()

	config := &Config{}
	if err = json.NewDecoder(configFile).Decode(config); err != nil {
		return nil, fmt.Errorf("failed parsing embedded file '%s': %w", configPath, err)
	}
	return config, nil
}

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/roemer/gonovate/internal/pkg/presets"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

// Loads the given configuration
func Load(configPath string) (*RootConfig, error) {
	if configPath == "" {
		configPath = "local:gonovate.json"
	}
	if !strings.Contains(configPath, ":") {
		configPath = fmt.Sprintf("local:%s", configPath)
	}
	configInfo, err := newConfigInfo(configPath)
	if err != nil {
		return nil, err
	}
	return loadConfig(nil, configInfo)
}

////////////////////////////////////////////////////////////
// Internal
////////////////////////////////////////////////////////////

const (
	infoTypePreset string = "preset"
	infoTypeLocal  string = "local"
	infoTypeWeb    string = "web"
)

var httpSchemeRegex = regexp.MustCompile(`^https?://.+`)

// Holds information about the type and location of a config
type configInfo struct {
	Type     string
	Location string
}

func newConfigInfo(info string) (*configInfo, error) {
	if info == "" {
		return nil, fmt.Errorf("empty config info")
	}

	var configType, configLoc string

	if httpSchemeRegex.MatchString(info) {
		// The info is an url, so use web
		configType = infoTypeWeb
		configLoc = info
	} else {
		parts := strings.SplitN(info, ":", 2)
		if len(parts) == 1 {
			configType = infoTypePreset
			configLoc = parts[0]
		} else {
			configType = parts[0]
			configLoc = parts[1]
		}
		// Append the json extension if needed
		if path.Ext(configLoc) == "" {
			configLoc += ".json"
		}
	}
	// Create the info object
	return &configInfo{
		Type:     configType,
		Location: configLoc,
	}, nil
}

func loadConfig(parentInfo, newInfo *configInfo) (*RootConfig, error) {
	var newConfig *RootConfig
	var err error
	// Try load the config according to the type
	switch newInfo.Type {
	case infoTypePreset:
		newConfig, err = loadConfigFromEmbeddedFile(newInfo.Location)
	case infoTypeLocal:
		newConfig, err = loadConfigFromFile(parentInfo, newInfo)
	case infoTypeWeb:
		newConfig, err = loadConfigFromWeb(newInfo.Location)
	default:
		return nil, fmt.Errorf("unknown config type '%s'", newInfo.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("failed reading config '%s:%s': %w", newInfo.Type, newInfo.Location, err)
	}

	// PreProcess the config
	newConfig.PostLoadProcess()

	// Create a new object for the merged config with the presets
	mergedConfig := &RootConfig{}
	// Process the "Extends" presets first
	for _, presetLookupInfo := range newConfig.Extends {
		presetInfo, err := newConfigInfo(presetLookupInfo)
		if err != nil {
			return nil, err
		}
		// Read the extended preset
		extendsConfig, err := loadConfig(newInfo, presetInfo)
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

func loadConfigFromFile(parentInfo, newInfo *configInfo) (*RootConfig, error) {
	// If the path is absolute, use it directly
	if filepath.IsAbs(newInfo.Location) {
		return readConfigFromFile(newInfo.Location)
	}

	// Try load it from the current folder
	if exists, err := shared.FileExists(newInfo.Location); err != nil {
		return nil, err
	} else if exists {
		return readConfigFromFile(newInfo.Location)
	}

	// Search in the folder of the parent config
	if parentInfo != nil && parentInfo.Type == infoTypeLocal && parentInfo.Location != "" {
		tempPresetPath := filepath.Clean(filepath.Join(filepath.Dir(parentInfo.Location), newInfo.Location))
		if exists, err := shared.FileExists(tempPresetPath); err != nil {
			return nil, err
		} else if exists {
			return readConfigFromFile(tempPresetPath)
		}
	}

	// Search in the current executable directory
	if executablePath, err := os.Executable(); err != nil {
		return nil, err
	} else {
		tempPresetPath := filepath.Clean(filepath.Join(filepath.Dir(executablePath), newInfo.Location))
		if exists, err := shared.FileExists(tempPresetPath); err != nil {
			return nil, err
		} else if exists {
			return readConfigFromFile(tempPresetPath)
		}
		// Also search in the presets subfolder
		tempPresetPath = filepath.Clean(filepath.Join(filepath.Dir(executablePath), "presets", newInfo.Location))
		if exists, err := shared.FileExists(tempPresetPath); err != nil {
			return nil, err
		} else if exists {
			return readConfigFromFile(tempPresetPath)
		}
	}

	// Search based on the current file that is executed (probably dev with go run . only)
	if _, filename, _, ok := runtime.Caller(0); ok {
		rootPath := filepath.Dir(filepath.Dir(filename))
		tempPresetPath := filepath.Clean(filepath.Join(rootPath, newInfo.Location))
		if exists, err := shared.FileExists(tempPresetPath); err != nil {
			return nil, err
		} else if exists {
			return readConfigFromFile(tempPresetPath)
		}
		// Also search in the presets subfolder
		tempPresetPath = filepath.Clean(filepath.Join(rootPath, "presets", newInfo.Location))
		if exists, err := shared.FileExists(tempPresetPath); err != nil {
			return nil, err
		} else if exists {
			return readConfigFromFile(tempPresetPath)
		}
	}

	// Nothing found at all
	return nil, fmt.Errorf("file not found for '%s'", newInfo.Location)
}

func readConfigFromFile(configPath string) (*RootConfig, error) {
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening file '%s': %w", configPath, err)
	}
	defer configFile.Close()

	config := &RootConfig{}
	if err = json.NewDecoder(configFile).Decode(config); err != nil {
		return nil, fmt.Errorf("failed parsing file '%s': %w", configPath, err)
	}
	return config, nil
}

func loadConfigFromEmbeddedFile(configPath string) (*RootConfig, error) {
	configFile, err := presets.Presets.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening embedded file '%s': %w", configPath, err)
	}
	defer configFile.Close()

	config := &RootConfig{}
	if err = json.NewDecoder(configFile).Decode(config); err != nil {
		return nil, fmt.Errorf("failed parsing embedded file '%s': %w", configPath, err)
	}
	return config, nil
}

func loadConfigFromWeb(url string) (*RootConfig, error) {
	content, err := shared.HttpUtil.DownloadToMemory(url)
	if err != nil {
		return nil, fmt.Errorf("failed downloading config from '%s': %w", url, err)
	}

	config := &RootConfig{}
	if err = json.Unmarshal(content, config); err != nil {
		return nil, fmt.Errorf("failed parsing config from '%s': %w", url, err)
	}
	return config, nil
}

package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/roemer/gonovate/pkg/presets"

	"github.com/goccy/go-yaml"
)

// Loads the given configuration
func Load(configPath string) (*GonovateConfig, error) {
	if configPath == "" {
		configPath = "local:gonovate"
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
	}
	// Create the info object
	return &configInfo{
		Type:     configType,
		Location: configLoc,
	}, nil
}

func loadConfig(parentInfo, newInfo *configInfo) (*GonovateConfig, error) {
	var newConfig *GonovateConfig
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
	mergedConfig := &GonovateConfig{}
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

func loadConfigFromFile(parentInfo, newInfo *configInfo) (*GonovateConfig, error) {
	// Build a list of paths that should be searched
	searchPaths := []string{}
	if filepath.IsAbs(newInfo.Location) {
		// For an absolute path, only use the absolute path
		searchPaths = append(searchPaths, newInfo.Location)
	} else {
		// Current folder
		searchPaths = append(searchPaths, newInfo.Location)

		// Folder of the parent config
		if parentInfo != nil && parentInfo.Type == infoTypeLocal && parentInfo.Location != "" {
			tempSearchPath := filepath.Clean(filepath.Join(filepath.Dir(parentInfo.Location), newInfo.Location))
			searchPaths = append(searchPaths, tempSearchPath)
		}

		// Current executable directory
		if executablePath, err := os.Executable(); err == nil {
			tempSearchPath := filepath.Clean(filepath.Join(filepath.Dir(executablePath), newInfo.Location))
			searchPaths = append(searchPaths, tempSearchPath)
		}

		// Based on the current file that is executed (probably with "go run" only)
		if _, filename, _, ok := runtime.Caller(0); ok {
			rootPath := filepath.Dir(filepath.Dir(filename))
			tempSearchPath := filepath.Clean(filepath.Join(rootPath, newInfo.Location))
			searchPaths = append(searchPaths, tempSearchPath)
		}
	}

	// Search thru the defined search paths
	hasExt := filepath.Ext(newInfo.Location) != ""
	finalValidConfigPath := ""
	for _, searchPath := range searchPaths {
		if hasExt {
			// We have an extension, directly search in the given path
			if exists, err := common.FileExists(searchPath); err != nil {
				return nil, err
			} else if exists {
				finalValidConfigPath = searchPath
				break
			}
		} else {
			// No extension, probe with the valid extensions
			if foundPath, err := SearchConfigFileFromPath(searchPath); err != nil {
				return nil, err
			} else if foundPath != "" {
				finalValidConfigPath = foundPath
				break
			}
		}
	}

	// Check if we have a valid path and if so, read it and return the config
	if finalValidConfigPath != "" {
		// Open the file
		configFile, err := os.Open(finalValidConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed opening file '%s': %w", finalValidConfigPath, err)
		}
		defer configFile.Close()

		// Decode the file
		config := &GonovateConfig{}
		if filepath.Ext(finalValidConfigPath) == ".json" {
			if err = json.NewDecoder(configFile).Decode(config); err != nil {
				return nil, fmt.Errorf("failed parsing file '%s': %w", finalValidConfigPath, err)
			}
		} else {
			if err = yaml.NewDecoder(configFile).Decode(config); err != nil {
				return nil, fmt.Errorf("failed parsing file '%s': %w", finalValidConfigPath, err)
			}
		}
		return config, nil
	}

	// Nothing found at all
	return nil, fmt.Errorf("file not found for '%s'", newInfo.Location)
}

func loadConfigFromEmbeddedFile(configPath string) (*GonovateConfig, error) {
	// Adjust the path to the config as they are all in a subfolder
	basePath := "configs"
	configPath = path.Join(basePath, configPath)

	// Get the extension
	ext := path.Ext(configPath)
	// If there is no extension, search for a yaml or json file
	if ext == "" {
		dirEntries, err := presets.Presets.ReadDir(path.Dir(configPath))
		if err != nil {
			return nil, err
		}
		foundPath, found := SearchConfigFileFromDirEntries(path.Base(configPath), dirEntries)
		if !found {
			return nil, fmt.Errorf("could not find a config for file '%s'", configPath)
		}
		configPath = path.Join(path.Dir(configPath), foundPath)
	}

	configFile, err := presets.Presets.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening embedded file '%s': %w", configPath, err)
	}
	defer configFile.Close()

	config := &GonovateConfig{}
	if path.Ext(configPath) == ".json" {
		if err = json.NewDecoder(configFile).Decode(config); err != nil {
			return nil, fmt.Errorf("failed parsing embedded file '%s': %w", configPath, err)
		}
	} else {
		if err = yaml.NewDecoder(configFile).Decode(config); err != nil {
			return nil, fmt.Errorf("failed parsing embedded file '%s': %w", configPath, err)
		}
	}
	return config, nil
}

func loadConfigFromWeb(urlString string) (*GonovateConfig, error) {
	// Check if the url is valid
	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	// Download it
	content, err := common.HttpUtil.DownloadToMemory(urlString)
	if err != nil {
		return nil, fmt.Errorf("failed downloading config from '%s': %w", urlString, err)
	}

	// Unmarshal it
	config := &GonovateConfig{}
	if path.Ext(parsedUrl.Path) == ".json" {
		if err = json.Unmarshal(content, config); err != nil {
			return nil, fmt.Errorf("failed parsing config from '%s': %w", urlString, err)
		}
	} else {
		if err = yaml.Unmarshal(content, config); err != nil {
			return nil, fmt.Errorf("failed parsing config from '%s': %w", urlString, err)
		}
	}
	return config, nil
}

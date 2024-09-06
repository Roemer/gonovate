package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// Checks if a project has its own configuration.
func HasProjectConfig() (string, error) {
	foundPath, err := SearchConfigFileFromPath("gonovate")
	return foundPath, err
}

// Searches for a config file with the given base name
func SearchConfigFileFromPath(searchPath string) (string, error) {
	dirPath := filepath.Dir(searchPath)
	baseName := filepath.Base(searchPath)
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	foundFileName, found := SearchConfigFileFromDirEntries(baseName, dirEntries)
	if found {
		return filepath.Join(dirPath, foundFileName), nil
	}
	return "", nil
}

func SearchConfigFileFromDirEntries(baseName string, entries []fs.DirEntry) (string, bool) {
	filesToCheck := []string{
		baseName + ".json",
		baseName + ".yaml",
		baseName + ".yml",
	}
	for _, entry := range entries {
		if slices.Contains(filesToCheck, entry.Name()) {
			return entry.Name(), true
		}
	}
	return "", false
}

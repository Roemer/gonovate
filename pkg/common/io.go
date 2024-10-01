package common

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

func FileExists(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err == nil {
		return !info.IsDir(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// Checks if the given filePath matches at least one of the given patterns.
func FilePathMatchesPattern(filePath string, patterns ...string) (bool, error) {
	if patterns == nil {
		return true, nil
	}
	for _, pattern := range patterns {
		isMatch, err := doublestar.Match(filepath.ToSlash(pattern), filepath.ToSlash(filePath))
		if err != nil {
			return false, err
		}
		if isMatch {
			return true, nil
		}
	}
	return false, nil
}

func SearchFiles(rootPath string, matchPaths []string, ignorePatterns []string) ([]string, error) {
	matchedFiles := []string{}
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		// Check if the path is in the ignore list and skip it
		for _, ignorePattern := range ignorePatterns {
			isMatch, err := FilePathMatchesPattern(path, ignorePattern)
			if err != nil {
				return err
			}
			if isMatch {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		// Continue on folders
		if d.IsDir() {
			return nil
		}
		// Check if the file matches
		for _, matchPath := range matchPaths {
			isMatch, _ := doublestar.Match(filepath.ToSlash(matchPath), filepath.ToSlash(path))
			if isMatch {
				matchedFiles = append(matchedFiles, path)
			}
		}
		return nil
	})
	return matchedFiles, err
}

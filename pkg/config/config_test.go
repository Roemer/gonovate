package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestPostLoadProcess(t *testing.T) {
	assert := assert.New(t)

	gonovateConfig := &GonovateConfig{
		Managers: []*Manager{
			{
				Id:   "Manager A",
				Type: common.MANAGER_TYPE_REGEX,
				ManagerConfig: &ManagerConfig{
					FilePatterns: []string{"pattern"},
				},
				DependencyConfig: &DependencyConfig{
					DependencyName: "depName",
					Versioning:     "1.0.0",
				},
			},
		},
	}

	// Test without pre-process
	assert.NotNil(gonovateConfig.Managers[0].ManagerConfig)
	assert.NotNil(gonovateConfig.Managers[0].DependencyConfig)
	assert.Len(gonovateConfig.Rules, 0)

	// Pre-process
	gonovateConfig.PostLoadProcess()

	// Test after pre-process
	assert.Nil(gonovateConfig.Managers[0].ManagerConfig)
	assert.Nil(gonovateConfig.Managers[0].DependencyConfig)
	assert.Len(gonovateConfig.Rules, 1)

	// Check rule
	checkRule := gonovateConfig.Rules[0]
	assert.ElementsMatch(checkRule.Matches.Managers, []string{"Manager A"})
	assert.NotNil(checkRule.ManagerConfig)
	assert.ElementsMatch(checkRule.ManagerConfig.FilePatterns, []string{"pattern"})
	assert.NotNil(checkRule.DependencyConfig)
	assert.Equal(checkRule.DependencyConfig.DependencyName, "depName")
	assert.Equal(checkRule.DependencyConfig.Versioning, "1.0.0")
}

func TestFileSearch(t *testing.T) {
	assert := assert.New(t)

	testFileSearch(assert, "gonovate.json", "gonovate")
	testFileSearch(assert, "gonovate.yaml", "gonovate")
	testFileSearch(assert, "gonovate.yml", "gonovate")

	testFileSearch(assert, "folder/gonovate.json", "folder/gonovate")
	testFileSearch(assert, "folder/gonovate.yaml", "folder/gonovate")
	testFileSearch(assert, "folder/gonovate.yml", "folder/gonovate")

	os.RemoveAll("folder")
}

func testFileSearch(assert *assert.Assertions, fileToCreate string, searchPath string) {
	// Create the file and folder structure
	err := os.MkdirAll(filepath.Dir(fileToCreate), os.ModePerm)
	assert.NoError(err)
	myfile, err := os.Create(fileToCreate)
	assert.NoError(err)
	myfile.Close()
	// Search for the file
	newPath, err := SearchConfigFileFromPath(searchPath)
	assert.NoError(err)
	assert.Equal(fileToCreate, filepath.ToSlash(newPath))
	// Delete the file
	os.Remove(fileToCreate)
}

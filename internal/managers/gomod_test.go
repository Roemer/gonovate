package managers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/roemer/gonovate/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestExtractGoVersion(t *testing.T) {
	assert := assert.New(t)

	path := filepath.Join("testdata", "go_mod_1")
	file, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("failed reading the testdata file")
	}

	manager := &GoModManager{}
	dependencies, err := manager.ExtractDependencies(string(file))
	assert.NoError(err)
	assert.NotNil(dependencies)
	assert.Greater(len(dependencies), 0)

	assertDependencyInSlice(assert, dependencies, "go", "1.21", "golang")
	assertDependencyInSlice(assert, dependencies, "github.com/roemer/gover", "v0.5.2", "direct")
	assertDependencyInSlice(assert, dependencies, "github.com/samber/lo", "v1.39.0", "direct")
	//assertDependencyInSlice(assert, dependencies, "example.com/indirectmodule1", "v1.0.0", "indirect")
	//assertDependencyInSlice(assert, dependencies, "example.com/indirectmodule2", "v2.0.0", "indirect")
}

func assertDependencyInSlice(assert *assert.Assertions, dependencies []*core.Dependency, expectedName string, expectedVersion string, expectedType string) {
	for _, dep := range dependencies {
		if dep.Name == expectedName && dep.Version == expectedVersion && dep.Type == expectedType {
			return
		}
	}
	assert.Failf("failed finding dependency", "dep: %s, ver: %s, type: %s, deps: %v", expectedName, expectedVersion, expectedType, dependencies)
}

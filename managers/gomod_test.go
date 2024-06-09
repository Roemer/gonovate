package managers

import (
	"os"
	"path/filepath"
	"testing"

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

	assertDependencyInSlice(assert, dependencies, "go", "1.14", "golang")
	assertDependencyInSlice(assert, dependencies, "example.com/thismodule", "v1.2.3", "direct")
	assertDependencyInSlice(assert, dependencies, "example.com/thatmodule", "v3.2.1", "direct")
	//assertDependencyInSlice(assert, dependencies, "example.com/indirectmodule1", "v1.0.0", "indirect")
	//assertDependencyInSlice(assert, dependencies, "example.com/indirectmodule2", "v2.0.0", "indirect")
}

func assertDependencyInSlice(assert *assert.Assertions, dependencies []*Dependency, expectedName string, expectedVersion string, expectedType string) {
	for _, dep := range dependencies {
		if dep.Name == expectedName && dep.Version == expectedVersion && dep.Type == expectedType {
			return
		}
	}
	assert.Failf("failed finding dependency", "dep: %s, ver: %s, type: %s, deps: %v", expectedName, expectedVersion, expectedType, dependencies)
}

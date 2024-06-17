package config

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/roemer/gonovate/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestMergeMultipleProjects(t *testing.T) {
	//assert := assert.New(t)

	globalConfig := &Config{
		Platform: core.PLATFORM_TYPE_GITLAB,
		Managers: []*Manager{
			{
				Id:   "Manager A",
				Type: core.MANAGER_TYPE_REGEX,
				packageSettings: &PackageSettings{
					PackageName: "init",
					Versioning:  "init",
				},
			},
		},
	}

	configA := &Config{
		Managers: []*Manager{
			{
				Id:              "Manager A",
				packageSettings: &PackageSettings{PackageName: "a"},
			},
		},
	}

	configB := &Config{
		Managers: []*Manager{
			{
				Id:              "Manager A",
				packageSettings: &PackageSettings{Versioning: "b"},
			},
		},
	}

	mergedConfigA := globalConfig.MergeWith(configA)
	mergedConfigB := globalConfig.MergeWith(configB)

	a, _ := json.MarshalIndent(mergedConfigA, "", "  ")
	b, _ := json.MarshalIndent(mergedConfigB, "", "  ")

	fmt.Println(string(a))
	fmt.Println(string(b))

	//go test -timeout 30s -run ^TestMergeMultipleProjects$ gonovate/core
}

func TestSomething(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Platform: core.PLATFORM_TYPE_GITLAB,
		Managers: []*Manager{
			{
				Id:              "Manager A",
				Type:            core.MANAGER_TYPE_REGEX,
				managerSettings: &ManagerSettings{Disabled: core.Ptr(false)},
			},
		},
		Rules: []*Rule{
			{
				Matches: &RuleMatch{
					Managers: []string{"Manager A"},
				},
				ManagerSettings: &ManagerSettings{
					Disabled: core.Ptr(true),
				},
			},
		},
	}
	assert.NotNil(config)
	managerSettings, packageRules := config.FilterForManager(config.Managers[0])
	assert.NotNil(managerSettings)
	assert.True(*managerSettings.Disabled)
	assert.Len(packageRules, 0)
}

func TestMatchStringPresets(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		MatchStringPresets: map[string]*MatchStringPreset{
			"test-0p": {
				MatchString: "0p",
			},
			"test-1p": {
				MatchString:       "1p-%s",
				ParameterDefaults: []string{"a"},
			},
			"test-2p": {
				MatchString:       "2p-%s-%s",
				ParameterDefaults: []string{"a", "b"},
			},
		},
	}
	assert.NotNil(config)

	resolved, err := config.ResolveMatchString("preset:test-0p")
	assert.NoError(err)
	assert.Equal("0p", resolved)

	resolved, err = config.ResolveMatchString("preset:test-1p")
	assert.NoError(err)
	assert.Equal("1p-a", resolved)
	resolved, err = config.ResolveMatchString("preset:test-1p()")
	assert.NoError(err)
	assert.Equal("1p-a", resolved)
	resolved, err = config.ResolveMatchString("preset:test-1p(b)")
	assert.NoError(err)
	assert.Equal("1p-b", resolved)

	resolved, err = config.ResolveMatchString("preset:test-2p")
	assert.NoError(err)
	assert.Equal("2p-a-b", resolved)
	resolved, err = config.ResolveMatchString("preset:test-2p()")
	assert.NoError(err)
	assert.Equal("2p-a-b", resolved)
	resolved, err = config.ResolveMatchString("preset:test-2p(c)")
	assert.NoError(err)
	assert.Equal("2p-c-b", resolved)
	resolved, err = config.ResolveMatchString("preset:test-2p(c,d)")
	assert.NoError(err)
	assert.Equal("2p-c-d", resolved)
	resolved, err = config.ResolveMatchString("preset:test-2p(,d)")
	assert.NoError(err)
	assert.Equal("2p-a-d", resolved)

	resolved, err = config.ResolveMatchString("preset:non-existing")
	assert.Error(err)
}

func TestVersioningPresets(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		VersioningPresets: map[string]string{
			"a": "foo",
		},
	}
	assert.NotNil(config)

	resolved, err := config.ResolveVersioning("preset:a")
	assert.NoError(err)
	assert.Equal("foo", resolved)

	resolved, err = config.ResolveVersioning("preset:b")
	assert.Error(err)

	resolved, err = config.ResolveVersioning("c")
	assert.NoError(err)
	assert.Equal("c", resolved)
}

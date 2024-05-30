package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Platform: PLATFORM_TYPE_GITLAB,
		Managers: []*Manager{
			{
				Id:              "Manager A",
				Type:            MANAGER_TYPE_REGEX,
				ManagerSettings: &ManagerSettings{Disabled: Ptr(false)},
			},
		},
		Rules: []*Rule{
			{
				Matches: &RuleMatch{
					Managers: []string{"Manager A"},
				},
				ManagerSettings: &ManagerSettings{
					Disabled: Ptr(true),
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

	resolved := config.ResolveMatchString("preset:test-0p")
	assert.Equal("0p", resolved)

	resolved = config.ResolveMatchString("preset:test-1p")
	assert.Equal("1p-a", resolved)
	resolved = config.ResolveMatchString("preset:test-1p()")
	assert.Equal("1p-a", resolved)
	resolved = config.ResolveMatchString("preset:test-1p(b)")
	assert.Equal("1p-b", resolved)

	resolved = config.ResolveMatchString("preset:test-2p")
	assert.Equal("2p-a-b", resolved)
	resolved = config.ResolveMatchString("preset:test-2p()")
	assert.Equal("2p-a-b", resolved)
	resolved = config.ResolveMatchString("preset:test-2p(c)")
	assert.Equal("2p-c-b", resolved)
	resolved = config.ResolveMatchString("preset:test-2p(c,d)")
	assert.Equal("2p-c-d", resolved)
	resolved = config.ResolveMatchString("preset:test-2p(,d)")
	assert.Equal("2p-a-d", resolved)
}

func TestVersioningPresets(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		VersioningPresets: map[string]string{
			"a": "foo",
		},
	}
	assert.NotNil(config)

	resolved := config.ResolveVersioning("preset:a")
	assert.Equal("foo", resolved)

	resolved = config.ResolveVersioning("preset:b")
	assert.Equal("preset:b", resolved)

	resolved = config.ResolveVersioning("c")
	assert.Equal("c", resolved)
}

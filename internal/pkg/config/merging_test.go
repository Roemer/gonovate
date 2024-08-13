package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMatchStringPresets(t *testing.T) {
	assert := assert.New(t)

	configA := &RootConfig{
		MatchStringPresets: map[string]*MatchStringPreset{
			"a_key":         {MatchString: "a_value", ParameterDefaults: []string{}},
			"overwrite_key": {MatchString: "overwrite_value_a", ParameterDefaults: []string{"a_pd"}},
		},
	}
	configB := &RootConfig{
		MatchStringPresets: map[string]*MatchStringPreset{
			"b_key":         {MatchString: "b_value", ParameterDefaults: []string{}},
			"overwrite_key": {MatchString: "overwrite_value_b", ParameterDefaults: []string{"b_pd"}},
		},
	}
	merged := configA.MergeWithAsCopy(configB)

	assert.Len(merged.MatchStringPresets, 3)
	assert.Contains(merged.MatchStringPresets, "a_key")
	assert.Contains(merged.MatchStringPresets, "b_key")
	assert.Contains(merged.MatchStringPresets, "overwrite_key")
	assert.Equal("a_value", merged.MatchStringPresets["a_key"].MatchString)
	assert.Equal("b_value", merged.MatchStringPresets["b_key"].MatchString)
	assert.Equal("overwrite_value_b", merged.MatchStringPresets["overwrite_key"].MatchString)
	assert.ElementsMatch(merged.MatchStringPresets["overwrite_key"].ParameterDefaults, []string{"b_pd"})
}

func TestMergeVersioningPresets(t *testing.T) {
	assert := assert.New(t)

	configA := &RootConfig{
		VersioningPresets: map[string]string{
			"a_key":         "a_value",
			"overwrite_key": "overwrite_value_a",
		},
	}
	configB := &RootConfig{
		VersioningPresets: map[string]string{
			"b_key":         "b_value",
			"overwrite_key": "overwrite_value_b",
		},
	}
	merged := configA.MergeWithAsCopy(configB)

	assert.Len(merged.VersioningPresets, 3)
	assert.Contains(merged.VersioningPresets, "a_key")
	assert.Contains(merged.VersioningPresets, "b_key")
	assert.Contains(merged.VersioningPresets, "overwrite_key")
	assert.Equal("a_value", merged.VersioningPresets["a_key"])
	assert.Equal("b_value", merged.VersioningPresets["b_key"])
	assert.Equal("overwrite_value_b", merged.VersioningPresets["overwrite_key"])
}

func TestMergeExtends(t *testing.T) {
	assert := assert.New(t)

	configA := &RootConfig{
		Extends: []string{"extend_a", "extend_both"},
	}
	configB := &RootConfig{
		Extends: []string{"extend_b", "extend_both"},
	}
	merged := configA.MergeWithAsCopy(configB)

	assert.Len(merged.Extends, 3)
	assert.Equal([]string{"extend_a", "extend_both", "extend_b"}, merged.Extends)
}

func TestMergeIgnorePatterns(t *testing.T) {
	assert := assert.New(t)

	configA := &RootConfig{
		IgnorePatterns: []string{"ignore_a", "ignore_both"},
	}
	configB := &RootConfig{
		IgnorePatterns: []string{"ignore_b", "ignore_both"},
	}
	merged := configA.MergeWithAsCopy(configB)

	assert.Len(merged.IgnorePatterns, 3)
	assert.Equal([]string{"ignore_a", "ignore_both", "ignore_b"}, merged.IgnorePatterns)
}

func TestMergePlatformSettings(t *testing.T) {
	assert := assert.New(t)

	configA := &RootConfig{
		PlatformSettings: &PlatformSettings{
			Token: "token_a",
		},
	}
	configB := &RootConfig{
		PlatformSettings: &PlatformSettings{
			Token: "token_b",
		},
	}
	merged := configA.MergeWithAsCopy(configB)

	assert.Equal("token_b", merged.PlatformSettings.Token)
}

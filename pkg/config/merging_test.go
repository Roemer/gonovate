package config

import (
	"testing"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestMergeMatchStringPresets(t *testing.T) {
	assert := assert.New(t)

	configA := &GonovateConfig{
		MatchStringPresets: map[string]*MatchStringPreset{
			"a_key":         {MatchString: "a_value", ParameterDefaults: []string{}},
			"overwrite_key": {MatchString: "overwrite_value_a", ParameterDefaults: []string{"a_pd"}},
		},
	}
	configB := &GonovateConfig{
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

	configA := &GonovateConfig{
		VersioningPresets: map[string]string{
			"a_key":         "a_value",
			"overwrite_key": "overwrite_value_a",
		},
	}
	configB := &GonovateConfig{
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

	configA := &GonovateConfig{
		Extends: []string{"extend_a", "extend_both"},
	}
	configB := &GonovateConfig{
		Extends: []string{"extend_b", "extend_both"},
	}
	merged := configA.MergeWithAsCopy(configB)

	assert.Len(merged.Extends, 3)
	assert.Equal([]string{"extend_a", "extend_both", "extend_b"}, merged.Extends)
}

func TestMergeIgnorePatterns(t *testing.T) {
	assert := assert.New(t)

	configA := &GonovateConfig{
		IgnorePatterns: []string{"ignore_a", "ignore_both"},
	}
	configB := &GonovateConfig{
		IgnorePatterns: []string{"ignore_b", "ignore_both"},
	}
	merged := configA.MergeWithAsCopy(configB)

	assert.Len(merged.IgnorePatterns, 3)
	assert.Equal([]string{"ignore_a", "ignore_both", "ignore_b"}, merged.IgnorePatterns)
}

func TestMergePlatformConfig(t *testing.T) {
	assert := assert.New(t)

	configA := &GonovateConfig{
		PlatformConfig: &PlatformConfig{
			Token: "token_a",
		},
	}
	configB := &GonovateConfig{
		PlatformConfig: &PlatformConfig{
			Token: "token_b",
		},
	}
	merged := configA.MergeWithAsCopy(configB)

	assert.Equal("token_b", merged.PlatformConfig.Token)
}

func TestMergeDevcontainerConfig(t *testing.T) {
	assert := assert.New(t)

	configA := &GonovateConfig{
		Managers: []*Manager{
			{
				Id: "manager",
				ManagerConfig: &ManagerConfig{
					DevcontainerConfig: map[string][]*DevcontainerFeatureDependency{
						"feature1": {
							&DevcontainerFeatureDependency{
								Property:       "prop1-1",
								Datasource:     common.DATASOURCE_TYPE_DOCKER,
								DependencyName: "dep1-1",
							},
						},
						"feature2": {
							{
								Property:       "prop2-1",
								Datasource:     common.DATASOURCE_TYPE_DOCKER,
								DependencyName: "dep2-1",
							},
						},
					},
				},
			},
		},
	}
	configB := &GonovateConfig{
		Managers: []*Manager{
			{
				Id: "manager",
				ManagerConfig: &ManagerConfig{
					DevcontainerConfig: map[string][]*DevcontainerFeatureDependency{
						"feature2": {
							&DevcontainerFeatureDependency{
								Property:       "prop2-2",
								Datasource:     common.DATASOURCE_TYPE_DOCKER,
								DependencyName: "dep2-2",
							},
							&DevcontainerFeatureDependency{
								Property:       "prop2-1",
								Datasource:     common.DATASOURCE_TYPE_MAVEN,
								DependencyName: "dep2-1-new",
							},
						},
						"feature3": {
							{
								Property:       "prop3-1",
								Datasource:     common.DATASOURCE_TYPE_DOCKER,
								DependencyName: "dep3-1",
							},
						},
					},
				},
			},
		},
	}

	merged := configA.MergeWithAsCopy(configB)

	settingsToCheck := merged.Managers[0].ManagerConfig.DevcontainerConfig
	assert.Len(settingsToCheck, 3)
	assert.Contains(settingsToCheck, "feature1")
	assert.Contains(settingsToCheck, "feature2")
	assert.Contains(settingsToCheck, "feature3")

	{
		feat1 := settingsToCheck["feature1"]
		assert.Len(feat1, 1)
		assert.Equal(feat1[0].Property, "prop1-1")
		assert.Equal(feat1[0].Datasource, common.DATASOURCE_TYPE_DOCKER)
		assert.Equal(feat1[0].DependencyName, "dep1-1")
	}

	{
		feat2 := settingsToCheck["feature2"]
		assert.Len(feat2, 2)
		assert.Equal(feat2[0].Property, "prop2-1")
		assert.Equal(feat2[0].Datasource, common.DATASOURCE_TYPE_MAVEN)
		assert.Equal(feat2[0].DependencyName, "dep2-1-new")
		assert.Equal(feat2[1].Property, "prop2-2")
		assert.Equal(feat2[1].Datasource, common.DATASOURCE_TYPE_DOCKER)
		assert.Equal(feat2[1].DependencyName, "dep2-2")
	}

	{
		feat3 := settingsToCheck["feature3"]
		assert.Len(feat3, 1)
		assert.Equal(feat3[0].Property, "prop3-1")
		assert.Equal(feat3[0].Datasource, common.DATASOURCE_TYPE_DOCKER)
		assert.Equal(feat3[0].DependencyName, "dep3-1")
	}
}

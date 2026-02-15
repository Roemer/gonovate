package config

import (
	"testing"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestApplyToDependency(t *testing.T) {
	assert := assert.New(t)

	cfg := &GonovateConfig{
		Managers: []*Manager{
			{
				Id: "manager",
				ManagerConfig: &ManagerConfig{
					FilePatterns: []string{"file.txt"},
				},
			},
		},
		Rules: []*Rule{
			{
				Matches: &RuleMatch{DependencyNames: []string{"dependency"}},
				DependencyConfig: &DependencyConfig{
					Datasource: common.DATASOURCE_TYPE_GOMOD,
				},
			},
			{
				Matches: &RuleMatch{Datasources: []common.DatasourceType{common.DATASOURCE_TYPE_GOMOD}},
				DependencyConfig: &DependencyConfig{
					GroupName: "gomod",
				},
			},
			{
				Matches: &RuleMatch{CurrentVersion: "1.0.1"},
				DependencyConfig: &DependencyConfig{
					Labels: []string{"MyLabel"},
				},
			},
		},
	}

	dependency := &common.Dependency{
		Name:     "dependency",
		Version:  "1.0.0",
		FilePath: "file.txt",
	}

	err := cfg.ApplyToDependency(dependency)
	assert.NoError(err)
	assert.Equal(common.DATASOURCE_TYPE_GOMOD, dependency.Datasource)
	assert.Equal("gomod", dependency.GroupName)
	assert.Equal([]string{"MyLabel"}, dependency.Labels)
}

// In this thest, the datasource of a rule should not override an already set datasource of the dependency
func TestApplyToDependencyWithExistingDatasource(t *testing.T) {
	assert := assert.New(t)

	cfg := &GonovateConfig{
		Managers: []*Manager{
			{
				Id: "manager",
				ManagerConfig: &ManagerConfig{
					FilePatterns: []string{"file.txt"},
				},
			},
		},
		Rules: []*Rule{
			{
				Matches: &RuleMatch{DependencyNames: []string{"dependency"}},
				DependencyConfig: &DependencyConfig{
					Datasource: common.DATASOURCE_TYPE_GOMOD,
				},
			},
			{
				Matches: &RuleMatch{Datasources: []common.DatasourceType{common.DATASOURCE_TYPE_GOMOD}},
				DependencyConfig: &DependencyConfig{
					GroupName: "gomod",
				},
			},
		},
	}

	dependency := &common.Dependency{
		Name:       "dependency",
		Version:    "1.0.0",
		FilePath:   "file.txt",
		Datasource: common.DATASOURCE_TYPE_ARTIFACTORY,
	}

	err := cfg.ApplyToDependency(dependency)
	assert.NoError(err)
	assert.Equal(common.DATASOURCE_TYPE_ARTIFACTORY, dependency.Datasource)
	assert.Equal("", dependency.GroupName)
}

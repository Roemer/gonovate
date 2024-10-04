package config

import (
	"log/slog"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/roemer/gonovate/pkg/presets"
	"github.com/samber/lo"
)

func (cfg *GonovateConfig) MatchStringPresetsToPresets() map[string]*presets.MatchStringPreset {
	return lo.MapEntries(cfg.MatchStringPresets, func(key string, value *MatchStringPreset) (string, *presets.MatchStringPreset) {
		return key, &presets.MatchStringPreset{
			MatchString:       value.MatchString,
			ParameterDefaults: value.ParameterDefaults,
		}
	})
}

func (managerConfig *ManagerConfig) ToCommonDevcontainerManagerSettings() *common.DevcontainerManagerSettings {
	return &common.DevcontainerManagerSettings{
		FeatureDependencies: lo.MapEntries(managerConfig.DevcontainerConfig,
			func(k string, value []*DevcontainerFeatureDependency) (string, []*common.DevcontainerManagerFeatureDependency) {
				return k, lo.Map(value, func(v *DevcontainerFeatureDependency, _ int) *common.DevcontainerManagerFeatureDependency {
					return &common.DevcontainerManagerFeatureDependency{
						Property:       v.Property,
						DependencyName: v.DependencyName,
						Datasource:     v.Datasource,
					}
				})
			},
		),
	}
}

func (cfg *GonovateConfig) ToCommonPlatformSettings(logger *slog.Logger) *common.PlatformSettings {
	return &common.PlatformSettings{
		Logger:     logger,
		Platform:   cfg.Platform,
		Token:      cfg.PlatformConfig.Token,
		Endpoint:   cfg.PlatformConfig.Endpoint,
		GitAuthor:  cfg.PlatformConfig.GitAuthor,
		BaseBranch: cfg.PlatformConfig.BaseBranch,
	}
}

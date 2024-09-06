package gonovate

import (
	"github.com/roemer/gonovate/pkg/common"
	"github.com/roemer/gonovate/pkg/config"
	"github.com/roemer/gonovate/pkg/datasources"
	"github.com/roemer/gonovate/pkg/managers"
)

// Get a manager with the given settings.
func GetManager(settings *common.ManagerSettings) (common.IManager, error) {
	return managers.GetManager(settings)
}

// Get a datasource with the given settings.
func GetDatasource(settings *common.DatasourceSettings) (common.IDatasource, error) {
	return datasources.GetDatasource(settings)
}

// Load the default configuration.
func LoadDefaultConfig() (*config.GonovateConfig, error) {
	return LoadConfig("preset:defaults")
}

// Load a given configuration.
func LoadConfig(configPath string) (*config.GonovateConfig, error) {
	return config.Load(configPath)
}

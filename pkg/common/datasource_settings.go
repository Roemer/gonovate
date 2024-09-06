package common

import "log/slog"

type DatasourceSettings struct {
	// The logger to use for the datasource.
	Logger *slog.Logger
	// The type of the datasource.
	DatasourceType DatasourceType
	// Host rules that might apply when using this datasource.
	HostRules []*HostRule
}

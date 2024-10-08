package common

import "log/slog"

type DatasourceSettings struct {
	// The logger to use for the datasource.
	Logger *slog.Logger
	// Host rules that might apply when using this datasource.
	HostRules []*HostRule
}

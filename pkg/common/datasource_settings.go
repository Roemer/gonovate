package common

import (
	"log/slog"

	"github.com/roemer/gonovate/pkg/cache"
)

type DatasourceSettings struct {
	// The logger to use for the datasource.
	Logger *slog.Logger
	// Host rules that might apply when using this datasource.
	HostRules []*HostRule
	// An optional cache to use.
	Cache cache.Cache[[]*ReleaseInfo]
}

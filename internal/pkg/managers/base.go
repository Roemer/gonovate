package managers

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/roemer/gonovate/internal/pkg/config"
	"github.com/roemer/gonovate/internal/pkg/shared"
)

// This is the interface that needs to be implemented by all managers.
type IManager interface {
	// Gets the id of the manager
	Id() string
	// Extracts all dependencies from the manager.
	ExtractDependencies(filePath string) ([]*shared.Dependency, error)
	// Applies a dependency update with the manager.
	ApplyDependencyUpdate(dependency *shared.Dependency) error
}

type managerBase struct {
	logger     *slog.Logger
	impl       IManager
	id         string
	rootConfig *config.RootConfig
	settings   *config.ManagerSettings
}

func GetManager(logger *slog.Logger, rootConfig *config.RootConfig, managerConfig *config.ManagerConfig, managerSettings *config.ManagerSettings) (IManager, error) {
	switch managerConfig.Type {
	case shared.MANAGER_TYPE_DOCKERFILE:
		return NewDockerfileManager(logger, managerConfig.Id, rootConfig, managerSettings), nil
	case shared.MANAGER_TYPE_GOMOD:
		return NewGoModManager(logger, managerConfig.Id, rootConfig, managerSettings), nil
	case shared.MANAGER_TYPE_INLINE:
		return NewInlineManager(logger, managerConfig.Id, rootConfig, managerSettings), nil
	case shared.MANAGER_TYPE_REGEX:
		return NewRegexManager(logger, managerConfig.Id, rootConfig, managerSettings), nil
	}
	return nil, fmt.Errorf("no manager defined for type '%s'", managerConfig.Type)
}

func (manager *managerBase) Id() string {
	return manager.id
}

// Returns a single dependency from a dependency slice.
func (manager *managerBase) getSingleDependency(dependencyName string, allDependencies []*shared.Dependency) (*shared.Dependency, error) {
	idx := slices.IndexFunc(allDependencies, func(dep *shared.Dependency) bool { return dep.Name == dependencyName })
	if idx < 0 {
		return nil, fmt.Errorf("failed to find dependency '%s'", dependencyName)
	}
	return allDependencies[idx], nil
}

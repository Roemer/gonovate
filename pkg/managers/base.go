package managers

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/roemer/gonovate/pkg/common"
)

// The internal base class for a manager.
type managerBase struct {
	id          string
	managerType common.ManagerType
	logger      *slog.Logger
	impl        common.IManager
	settings    *common.ManagerSettings
}

func newManagerBase(id string, managerType common.ManagerType, settings *common.ManagerSettings) *managerBase {
	return &managerBase{
		id:          id,
		managerType: managerType,
		logger:      settings.Logger.With(slog.String("manager", id)),
		settings:    settings,
	}
}

func GetManager(id string, managerType common.ManagerType, settings *common.ManagerSettings) (common.IManager, error) {
	switch managerType {
	case common.MANAGER_TYPE_DEVCONTAINER:
		return NewDevcontainerManager(id, settings), nil
	case common.MANAGER_TYPE_DOCKER_COMPOSE:
		return NewDockerComposeManager(id, settings), nil
	case common.MANAGER_TYPE_DOCKERFILE:
		return NewDockerfileManager(id, settings), nil
	case common.MANAGER_TYPE_GOMOD:
		return NewGoModManager(id, settings), nil
	case common.MANAGER_TYPE_INLINE:
		return NewInlineManager(id, settings), nil
	case common.MANAGER_TYPE_REGEX:
		return NewRegexManager(id, settings), nil
	}
	return nil, fmt.Errorf("no manager defined for type '%s'", managerType)
}

func (manager *managerBase) Id() string {
	return manager.id
}

func (manager *managerBase) Type() common.ManagerType {
	return manager.managerType
}

func (manager *managerBase) Settings() *common.ManagerSettings {
	return manager.settings
}

// Creates a new dependency some fields prefilled.
func (manager *managerBase) newDependency(name string, datasource common.DatasourceType, version string, filePath string) *common.Dependency {
	return &common.Dependency{
		Name:       name,
		Datasource: datasource,
		Version:    version,
		FilePath:   filePath,
		ManagerInfo: &common.ManagerInfo{
			ManagerId: manager.id,
		},
		AdditionalData: map[string]string{},
	}
}

// Returns a single dependency from a dependency slice.
func (manager *managerBase) getSingleDependency(dependencyName string, allDependencies []*common.Dependency) (*common.Dependency, error) {
	idx := slices.IndexFunc(allDependencies, func(dep *common.Dependency) bool { return dep.Name == dependencyName })
	if idx < 0 {
		return nil, fmt.Errorf("failed to find dependency '%s'", dependencyName)
	}
	return allDependencies[idx], nil
}

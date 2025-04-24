package managers

import (
	"log/slog"
	"testing"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestKubernetesManagerExtractSingle(t *testing.T) {
	assert := assert.New(t)

	manager := NewKubernetesManager("manager", &common.ManagerSettings{
		Logger: slog.Default(),
	})
	dependencies, err := manager.ExtractDependencies(`../../testdata/kubernetes/simple.yml`)
	assert.NoError(err)
	assert.NotNil(dependencies)
	assert.Len(dependencies, 1)
}

func TestKubernetesManagerExtractMulti(t *testing.T) {
	assert := assert.New(t)

	manager := NewKubernetesManager("manager", &common.ManagerSettings{
		Logger: slog.Default(),
	})
	dependencies, err := manager.ExtractDependencies(`../../testdata/kubernetes/multi.yml`)
	assert.NoError(err)
	assert.NotNil(dependencies)
	assert.Len(dependencies, 1)
}

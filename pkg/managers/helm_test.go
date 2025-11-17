package managers

import (
	"log/slog"
	"testing"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestKubernetesManagerExtract(t *testing.T) {
	assert := assert.New(t)

	manager := NewHelmManager("manager", &common.ManagerSettings{
		Logger: slog.Default(),
	})
	dependencies, err := manager.ExtractDependencies(`../../testdata/helm/Chart.yaml`)
	assert.NoError(err)
	assert.NotNil(dependencies)
	assert.Len(dependencies, 2)
}

package managers

import (
	"log/slog"
	"testing"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestDockerComposeManagerExtract(t *testing.T) {
	assert := assert.New(t)

	manager := NewDockerComposeManager("manager", &common.ManagerSettings{
		Logger: slog.Default(),
	})
	dependencies, err := manager.ExtractDependencies(`../../testdata/docker-compose/a/docker-compose.yml`)
	assert.NoError(err)
	assert.NotNil(dependencies)
	assert.Len(dependencies, 2)
}

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Platform: PLATFORM_TYPE_GITLAB,
		Managers: []*Manager{
			{
				Id:              "Manager A",
				Type:            MANAGER_TYPE_REGEX,
				ManagerSettings: &ManagerSettings{Disabled: Ptr(false)},
			},
		},
		Rules: []*Rule{
			{
				Matches: &RuleMatch{
					Managers: []string{"Manager A"},
				},
				ManagerSettings: &ManagerSettings{
					Disabled: Ptr(true),
				},
			},
		},
	}
	assert.NotNil(config)
	managerSettings, packageRules := config.FilterForManager(config.Managers[0])
	assert.NotNil(managerSettings)
	assert.True(*managerSettings.Disabled)
	assert.Len(packageRules, 0)
}

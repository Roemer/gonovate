package platforms

import (
	"log/slog"
	"testing"

	"github.com/roemer/gonovate/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestGetCorrectPlatform(t *testing.T) {
	assert := assert.New(t)

	platform, err := GetPlatform(&common.PlatformSettings{Logger: slog.Default(), Platform: common.PLATFORM_TYPE_GIT})
	assert.NoError(err)
	assert.NotNil(platform)
	assert.Equal(common.PLATFORM_TYPE_GIT, platform.Type())
	assert.IsType(&GitPlatform{}, platform)

	platform, err = GetPlatform(&common.PlatformSettings{Logger: slog.Default(), Platform: common.PLATFORM_TYPE_GITEA})
	assert.NoError(err)
	assert.NotNil(platform)
	assert.Equal(common.PLATFORM_TYPE_GITEA, platform.Type())
	assert.IsType(&GiteaPlatform{}, platform)

	platform, err = GetPlatform(&common.PlatformSettings{Logger: slog.Default(), Platform: common.PLATFORM_TYPE_GITHUB})
	assert.NoError(err)
	assert.NotNil(platform)
	assert.Equal(common.PLATFORM_TYPE_GITHUB, platform.Type())
	assert.IsType(&GitHubPlatform{}, platform)

	platform, err = GetPlatform(&common.PlatformSettings{Logger: slog.Default(), Platform: common.PLATFORM_TYPE_GITLAB})
	assert.NoError(err)
	assert.NotNil(platform)
	assert.Equal(common.PLATFORM_TYPE_GITLAB, platform.Type())
	assert.IsType(&GitlabPlatform{}, platform)

	platform, err = GetPlatform(&common.PlatformSettings{Logger: slog.Default(), Platform: common.PLATFORM_TYPE_NOOP})
	assert.NoError(err)
	assert.NotNil(platform)
	assert.Equal(common.PLATFORM_TYPE_NOOP, platform.Type())
	assert.IsType(&NoopPlatform{}, platform)
}

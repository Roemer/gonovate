package presets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchStringPresets(t *testing.T) {
	assert := assert.New(t)

	matchStringPresets := map[string]*MatchStringPreset{
		"test-0p": {
			MatchString: "0p",
		},
		"test-1p": {
			MatchString:       "1p-%s",
			ParameterDefaults: []string{"a"},
		},
		"test-2p": {
			MatchString:       "2p-%s-%s",
			ParameterDefaults: []string{"a", "b"},
		},
	}

	resolved, err := ResolveMatchString("preset:test-0p", matchStringPresets)
	assert.NoError(err)
	assert.Equal("0p", resolved)

	resolved, err = ResolveMatchString("preset:test-1p", matchStringPresets)
	assert.NoError(err)
	assert.Equal("1p-a", resolved)
	resolved, err = ResolveMatchString("preset:test-1p()", matchStringPresets)
	assert.NoError(err)
	assert.Equal("1p-a", resolved)
	resolved, err = ResolveMatchString("preset:test-1p(b)", matchStringPresets)
	assert.NoError(err)
	assert.Equal("1p-b", resolved)

	resolved, err = ResolveMatchString("preset:test-2p", matchStringPresets)
	assert.NoError(err)
	assert.Equal("2p-a-b", resolved)
	resolved, err = ResolveMatchString("preset:test-2p()", matchStringPresets)
	assert.NoError(err)
	assert.Equal("2p-a-b", resolved)
	resolved, err = ResolveMatchString("preset:test-2p(c)", matchStringPresets)
	assert.NoError(err)
	assert.Equal("2p-c-b", resolved)
	resolved, err = ResolveMatchString("preset:test-2p(c,d)", matchStringPresets)
	assert.NoError(err)
	assert.Equal("2p-c-d", resolved)
	resolved, err = ResolveMatchString("preset:test-2p(,d)", matchStringPresets)
	assert.NoError(err)
	assert.Equal("2p-a-d", resolved)

	_, err = ResolveMatchString("preset:non-existing", matchStringPresets)
	assert.Error(err)
}

func TestVersioningPresets(t *testing.T) {
	assert := assert.New(t)

	versioningPresets := map[string]string{
		"a": "foo",
	}

	resolved, err := ResolveVersioning("preset:a", versioningPresets)
	assert.NoError(err)
	assert.Equal("foo", resolved)

	_, err = ResolveVersioning("preset:b", versioningPresets)
	assert.Error(err)

	resolved, err = ResolveVersioning("c", versioningPresets)
	assert.NoError(err)
	assert.Equal("c", resolved)
}

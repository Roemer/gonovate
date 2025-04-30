package platforms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAuthor(t *testing.T) {
	assert := assert.New(t)

	name, email := splitAuthor("gonovate-bot <bot@gonovate.org>")
	assert.Equal("gonovate-bot", name)
	assert.Equal("bot@gonovate.org", email)

	name, email = splitAuthor("gonovate-bot <>")
	assert.Equal("gonovate-bot", name)
	assert.Equal("", email)
}

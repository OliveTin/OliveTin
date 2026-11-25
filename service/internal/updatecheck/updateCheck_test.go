package updatecheck

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersionLater(t *testing.T) {
	assert.Equal(t, "none", parseIfVersionIsLater("1.0.0", "1.0.0"))
	assert.Equal(t, "1.1.1", parseIfVersionIsLater("1.0.0", "1.1.1"))
	assert.Equal(t, "none", parseIfVersionIsLater("2.0.0", "1.10.1"))
	assert.Equal(t, "version-parse-failure", parseIfVersionIsLater("1.0.0", "1.2.3.4"))
	assert.Equal(t, "version-parse-failure", parseIfVersionIsLater("asdf", "foobar"))
	assert.Equal(t, "3000.23.0", parseIfVersionIsLater("3000.22.0", "3000.23.0"))
	assert.Equal(t, "none", parseIfVersionIsLater("3000.10.0", "3000.9.0"))
	assert.Equal(t, "v1.1.0", parseIfVersionIsLater("v1.0.0", "v1.1.0"))
	assert.Equal(t, "version-parse-failure", parseIfVersionIsLater("dev", "3000.23.0"))
}

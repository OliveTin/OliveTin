package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateDefaultConfig(t *testing.T) {
	c := DefaultConfig()

	assert.NotNil(t, c, "Create a default config")
}

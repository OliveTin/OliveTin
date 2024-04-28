package stringvariables

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEntityCount(t *testing.T) {
	SetEntityCount("waffles", 3)

	assert.Equal(t, 3, GetEntityCount("waffles"))
}

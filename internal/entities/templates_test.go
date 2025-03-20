package entities

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEntityCount(t *testing.T) {
	AddEntity("waffles", "0", nil)
	AddEntity("waffles", "1", nil)
	AddEntity("waffles", "2", nil)

	assert.Equal(t, 3, len(GetEntities("waffles")))
}

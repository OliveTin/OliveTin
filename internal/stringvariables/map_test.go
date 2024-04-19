package stringvariables

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAndSet(t *testing.T) {
	Set("foo", "bar")
	Set("salutation", "hello")

	assert.Equal(t, "bar", Get("foo"))
	assert.Equal(t, "", Get("not exist"))
}

func TestGetall(t *testing.T) {
	ret := GetAll()

	assert.NotEmpty(t, ret)
}

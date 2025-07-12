package entityfiles

import (
	"testing"
	"github.com/stretchr/testify/assert"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
)

func TestLoadObjectPerLineJsonFile(t *testing.T) {
	filename := "testdata/object-per-line.json"

	assert.Equal(t, "", sv.Get("entities.testrow.0.val"), "Value should match expected value")

	loadEntityFileJson(filename, "testrow")

	assert.Equal(t, "1234567890", sv.Get("entities.testrow.0.val"), "Value should match expected value")
}

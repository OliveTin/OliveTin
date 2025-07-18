package entityfiles

import (
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadObjectPerLineJsonFile(t *testing.T) {
	filename := "testdata/object-per-line.json"

	assert.Equal(t, "", sv.Get("entities.testrow.0.val"), "Value should match expected value")

	loadEntityFileJson(filename, "testrow")

	assert.Equal(t, "1234567890", sv.Get("entities.testrow.0.val"), "Value should match expected value")
}

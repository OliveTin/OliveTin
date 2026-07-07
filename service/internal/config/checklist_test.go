package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseChecklistValueJSON(t *testing.T) {
	t.Parallel()

	values, err := ParseChecklistValue(`["documents","photos"]`)
	require.NoError(t, err)
	assert.Equal(t, []string{"documents", "photos"}, values)

	values, err = ParseChecklistValue(`["kitchen,bedroom","hallway"]`)
	require.NoError(t, err)
	assert.Equal(t, []string{"kitchen,bedroom", "hallway"}, values)
}

func TestParseChecklistValueLegacyCommaSeparated(t *testing.T) {
	t.Parallel()

	values, err := ParseChecklistValue("documents, photos")
	require.NoError(t, err)
	assert.Equal(t, []string{"documents", "photos"}, values)
}

func TestParseChecklistValueRejectsEmptyLegacySegment(t *testing.T) {
	t.Parallel()

	_, err := ParseChecklistValue("documents,,photos")
	require.Error(t, err)
}

func TestFormatChecklistValueJSON(t *testing.T) {
	t.Parallel()

	assert.Equal(t, `["documents","photos"]`, FormatChecklistValue([]string{"documents", "photos"}))
	assert.Equal(t, `["kitchen,bedroom"]`, FormatChecklistValue([]string{"kitchen,bedroom"}))
	assert.Empty(t, FormatChecklistValue(nil))
}

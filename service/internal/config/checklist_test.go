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

func TestParseChecklistValueSingleValue(t *testing.T) {
	t.Parallel()

	values, err := ParseChecklistValue("documents")
	require.NoError(t, err)
	assert.Equal(t, []string{"documents"}, values)
}

func TestParseChecklistValueRejectsLegacyCommaSeparated(t *testing.T) {
	t.Parallel()

	_, err := ParseChecklistValue("documents, photos")
	require.Error(t, err)
}

func TestParseChecklistValueRejectsEmptyJSONSegment(t *testing.T) {
	t.Parallel()

	_, err := ParseChecklistValue(`["documents","","photos"]`)
	require.Error(t, err)
}

func TestFormatChecklistValueJSON(t *testing.T) {
	t.Parallel()

	encoded, err := FormatChecklistValue([]string{"documents", "photos"})
	require.NoError(t, err)
	assert.Equal(t, `["documents","photos"]`, encoded)

	encoded, err = FormatChecklistValue([]string{"kitchen,bedroom"})
	require.NoError(t, err)
	assert.Equal(t, `["kitchen,bedroom"]`, encoded)

	encoded, err = FormatChecklistValue(nil)
	require.NoError(t, err)
	assert.Empty(t, encoded)
}

package webhooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const jsonPathTestPayload = `{
	"repository": {"name": "OliveTin", "private": false, "stars": 5, "topics": null},
	"commits": [{"id": "a1"}, {"id": "b2"}],
	"ref": "refs/heads/main",
	"key with space": "spaced"
}`

func newTestJSONMatcher(t *testing.T) *JSONMatcher {
	t.Helper()
	matcher, err := NewJSONMatcher([]byte(jsonPathTestPayload))
	require.NoError(t, err)
	return matcher
}

func TestExtractValue(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"$.repository.name", "OliveTin"},
		{"$.ref", "refs/heads/main"},
		{"$.repository.private", "false"},
		{"$.repository.stars", "5"},
		{"$.repository.topics", "null"},
		{"$.commits[0].id", "a1"},
		{"$.commits[*].id", `["a1","b2"]`},
		{"$..id", `["a1","b2"]`},
		{"$['key with space']", "spaced"},
		{"repository.name", "OliveTin"},
	}

	matcher := newTestJSONMatcher(t)

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			value, err := matcher.ExtractValue(tt.path)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestExtractValueErrorsWhenPathIsAbsentOrInvalid(t *testing.T) {
	matcher := newTestJSONMatcher(t)

	for _, path := range []string{"$.missing", "$.repository.missing", "$.commits[5]", "$.ref.deeper", "", "$["} {
		t.Run(path, func(t *testing.T) {
			_, err := matcher.ExtractValue(path)
			assert.Error(t, err)
		})
	}
}

func TestMatchPath(t *testing.T) {
	matcher := newTestJSONMatcher(t)

	matches, err := matcher.MatchPath("$.repository.name", "OliveTin")
	require.NoError(t, err)
	assert.True(t, matches)

	matches, err = matcher.MatchPath("$.repository.private", "false")
	require.NoError(t, err)
	assert.True(t, matches)

	matches, err = matcher.MatchPath("$.repository.name", "Other")
	require.NoError(t, err)
	assert.False(t, matches)

	_, err = matcher.MatchPath("$.missing", "x")
	assert.Error(t, err)
}

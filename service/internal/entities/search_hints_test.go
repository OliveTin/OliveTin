package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSearchHintsReturnsLightweightIdentities(t *testing.T) {
	ClearEntitiesOfType("search_hint_host")
	ClearEntitiesOfType("search_hint_app")
	t.Cleanup(func() {
		ClearEntitiesOfType("search_hint_host")
		ClearEntitiesOfType("search_hint_app")
	})

	AddEntity("search_hint_host", "0", map[string]any{
		"name":     "web01",
		"secret":   "should-not-be-copied-by-search-hints",
		"hostname": "192.168.1.10",
	})
	AddEntity("search_hint_app", "app-1", map[string]any{
		"title": "Frontend",
	})

	hints := ListSearchHints()
	byKey := map[string]SearchHint{}
	for _, hint := range hints {
		byKey[hint.Type+":"+hint.UniqueKey] = hint
	}

	host, ok := byKey["search_hint_host:0"]
	require.True(t, ok)
	assert.Equal(t, "web01", host.Title)
	assert.Equal(t, "search_hint_host", host.Type)
	assert.Equal(t, "0", host.UniqueKey)

	app, ok := byKey["search_hint_app:app-1"]
	require.True(t, ok)
	assert.Equal(t, "Frontend", app.Title)
}

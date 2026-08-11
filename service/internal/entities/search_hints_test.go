package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSearchHintsReturnsLightweightIdentities(testContext *testing.T) {
	ClearEntitiesOfType("search_hint_host")
	ClearEntitiesOfType("search_hint_app")
	testContext.Cleanup(func() {
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

	host, hostFound := byKey["search_hint_host:0"]
	require.True(testContext, hostFound)
	assert.Equal(testContext, "web01", host.Title)
	assert.Equal(testContext, "search_hint_host", host.Type)
	assert.Equal(testContext, "0", host.UniqueKey)

	app, appFound := byKey["search_hint_app:app-1"]
	require.True(testContext, appFound)
	assert.Equal(testContext, "Frontend", app.Title)
}

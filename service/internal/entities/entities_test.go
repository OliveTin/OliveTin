package entities

import (
	"os"
	"path/filepath"
	"testing"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadObjectPerLineJsonFile(t *testing.T) {
	/*
		filename := "testdata/object-per-line.json"

		assert.Equal(t, "", GetEntity("testrow", "0"), "Value should match expected value")

		loadEntityFileJson(filename, "testrow")

		assert.Equal(t, "1234567890", GetEntity("testrow", "0"), "Value should match expected value")
	*/
}

func TestGetEntityInstancesOrdered_numericKeys(t *testing.T) {
	ClearEntitiesOfType("order_test")
	defer ClearEntitiesOfType("order_test")

	AddEntity("order_test", "2", map[string]any{"title": "Second"})
	AddEntity("order_test", "0", map[string]any{"title": "Zeroth"})
	AddEntity("order_test", "10", map[string]any{"title": "Tenth"})
	AddEntity("order_test", "1", map[string]any{"title": "First"})

	ordered := GetEntityInstancesOrdered("order_test")
	require.Len(t, ordered, 4, "should return 4 entities")
	assert.Equal(t, "0", ordered[0].UniqueKey, "first key should be 0")
	assert.Equal(t, "1", ordered[1].UniqueKey, "second key should be 1")
	assert.Equal(t, "2", ordered[2].UniqueKey, "third key should be 2")
	assert.Equal(t, "10", ordered[3].UniqueKey, "fourth key should be 10 (numeric order)")
}

func TestGetEntityInstancesOrdered_lexicographicKeys(t *testing.T) {
	ClearEntitiesOfType("order_test_lex")
	defer ClearEntitiesOfType("order_test_lex")

	AddEntity("order_test_lex", "zebra", map[string]any{"title": "Z"})
	AddEntity("order_test_lex", "alpha", map[string]any{"title": "A"})
	AddEntity("order_test_lex", "beta", map[string]any{"title": "B"})

	ordered := GetEntityInstancesOrdered("order_test_lex")
	require.Len(t, ordered, 3, "should return 3 entities")
	assert.Equal(t, "alpha", ordered[0].UniqueKey)
	assert.Equal(t, "beta", ordered[1].UniqueKey)
	assert.Equal(t, "zebra", ordered[2].UniqueKey)
}

func TestGetEntityInstancesOrdered_emptyOrMissing(t *testing.T) {
	ordered := GetEntityInstancesOrdered("nonexistent_type")
	assert.Nil(t, ordered)

	ClearEntitiesOfType("empty_test")
	ordered = GetEntityInstancesOrdered("empty_test")
	assert.Nil(t, ordered)
}

func TestSyncEntityFileWatchers_doesNotDuplicateWatchers(t *testing.T) {
	ResetEntityWatchersForTests()
	t.Cleanup(ResetEntityWatchersForTests)

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "vehicles.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte("- title: car1\n"), 0o600))

	cfg := config.DefaultConfig()
	cfg.SetDir(dir)
	cfg.Entities = []*config.EntityFile{
		{Name: "vehicle", File: "vehicles.yaml"},
	}

	SyncEntityFileWatchers(cfg)
	require.Equal(t, 1, watchedPathCountForTests())
	require.Len(t, GetEntityInstancesOrdered("vehicle"), 1)

	SyncEntityFileWatchers(cfg)
	assert.Equal(t, 1, watchedPathCountForTests())
}

func TestSyncEntityFileWatchers_picksUpNewEntityType(t *testing.T) {
	ResetEntityWatchersForTests()
	t.Cleanup(func() {
		ResetEntityWatchersForTests()
		ClearEntitiesOfType("vehicle")
		ClearEntitiesOfType("server")
	})

	dir := t.TempDir()
	vehiclePath := filepath.Join(dir, "vehicles.yaml")
	serverPath := filepath.Join(dir, "servers.yaml")
	require.NoError(t, os.WriteFile(vehiclePath, []byte("- title: car1\n"), 0o600))

	cfg := config.DefaultConfig()
	cfg.SetDir(dir)
	cfg.Entities = []*config.EntityFile{
		{Name: "vehicle", File: "vehicles.yaml"},
	}

	SyncEntityFileWatchers(cfg)
	require.Equal(t, 1, watchedPathCountForTests())

	require.NoError(t, os.WriteFile(serverPath, []byte("- name: srv1\n"), 0o600))
	cfg.Entities = append(cfg.Entities, &config.EntityFile{Name: "server", File: "servers.yaml"})

	SyncEntityFileWatchers(cfg)
	require.Equal(t, 2, watchedPathCountForTests())
	require.Len(t, GetEntityInstancesOrdered("vehicle"), 1)
	require.Len(t, GetEntityInstancesOrdered("server"), 1)
}

func TestLoadEntityFile_preservesEntitiesOnTransientFailure(t *testing.T) {
	const entityName = "preserve_on_fail"
	ClearEntitiesOfType(entityName)
	defer ClearEntitiesOfType(entityName)

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "hosts.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte("- title: kept\n"), 0o600))

	loadEntityFile(yamlPath, entityName)
	require.Len(t, GetEntityInstancesOrdered(entityName), 1)

	loadEntityFile(filepath.Join(dir, "missing.yaml"), entityName)
	require.Len(t, GetEntityInstancesOrdered(entityName), 1, "read failure should keep last good entities")

	require.NoError(t, os.WriteFile(yamlPath, []byte("not: valid: yaml: ["), 0o600))
	loadEntityFile(yamlPath, entityName)
	require.Len(t, GetEntityInstancesOrdered(entityName), 1, "parse failure should keep last good entities")

	jsonPath := filepath.Join(dir, "hosts.json")
	require.NoError(t, os.WriteFile(jsonPath, []byte("{\"title\":\"json-kept\"}\n"), 0o600))
	loadEntityFile(jsonPath, entityName)
	require.Len(t, GetEntityInstancesOrdered(entityName), 1)

	require.NoError(t, os.WriteFile(jsonPath, []byte("{bad json"), 0o600))
	loadEntityFile(jsonPath, entityName)
	ordered := GetEntityInstancesOrdered(entityName)
	require.Len(t, ordered, 1, "JSON parse failure should keep last good entities")
	assert.Equal(t, "json-kept", ordered[0].Title)
}

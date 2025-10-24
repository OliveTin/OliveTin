package config

import (
	"os"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
)

func TestUserLoadingFromConfig(t *testing.T) {
	// Create a temporary test config file
	testConfig := `
authLocalUsers:
  enabled: true
  users:
    - username: testuser1
      usergroup: admin
      password: password1
    - username: testuser2
      usergroup: guest
      password: password2

actions:
  - title: Test Action
    shell: echo "test"
`

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	assert.NoError(t, err, "Should create temporary file")
	defer os.Remove(tmpFile.Name())

	// Write test config to file
	_, err = tmpFile.WriteString(testConfig)
	assert.NoError(t, err, "Should write test config to file")
	tmpFile.Close()

	// Load config using koanf
	k := koanf.New(".")
	err = k.Load(file.Provider(tmpFile.Name()), yaml.Parser())
	assert.NoError(t, err, "Should load config file")

	// Create config struct and load it
	cfg := &Config{}
	AppendSource(cfg, k, tmpFile.Name())

	// Test that authLocalUsers was loaded correctly
	assert.True(t, cfg.AuthLocalUsers.Enabled, "AuthLocalUsers should be enabled")
	assert.Equal(t, 2, len(cfg.AuthLocalUsers.Users), "Should load 2 users")

	// Test individual users
	user1 := cfg.FindUserByUsername("testuser1")
	assert.NotNil(t, user1, "Should find testuser1")
	assert.Equal(t, "testuser1", user1.Username, "User1 should have correct username")
	assert.Equal(t, "admin", user1.Usergroup, "User1 should have correct usergroup")
	assert.Equal(t, "password1", user1.Password, "User1 should have correct password")

	user2 := cfg.FindUserByUsername("testuser2")
	assert.NotNil(t, user2, "Should find testuser2")
	assert.Equal(t, "testuser2", user2.Username, "User2 should have correct username")
	assert.Equal(t, "guest", user2.Usergroup, "User2 should have correct usergroup")
	assert.Equal(t, "password2", user2.Password, "User2 should have correct password")

	// Test non-existent user
	assert.Nil(t, cfg.FindUserByUsername("nonexistent"), "Should return nil for non-existent user")
}

func TestUserLoadingWithEmptyUsers(t *testing.T) {
	// Test config with enabled but no users
	testConfig := `
authLocalUsers:
  enabled: true
  users: []

actions:
  - title: Test Action
    shell: echo "test"
`

	tmpFile, err := os.CreateTemp("", "test_config_empty_*.yaml")
	assert.NoError(t, err, "Should create temporary file")
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testConfig)
	assert.NoError(t, err, "Should write test config to file")
	tmpFile.Close()

	k := koanf.New(".")
	err = k.Load(file.Provider(tmpFile.Name()), yaml.Parser())
	assert.NoError(t, err, "Should load config file")

	cfg := &Config{}
	AppendSource(cfg, k, tmpFile.Name())

	assert.True(t, cfg.AuthLocalUsers.Enabled, "AuthLocalUsers should be enabled")
	assert.Equal(t, 0, len(cfg.AuthLocalUsers.Users), "Should have 0 users")
	assert.Nil(t, cfg.FindUserByUsername("anyuser"), "Should return nil for any user")
}

func TestUserLoadingWithDisabledAuth(t *testing.T) {
	// Test config with disabled auth
	testConfig := `
authLocalUsers:
  enabled: false
  users:
    - username: testuser
      usergroup: admin
      password: password

actions:
  - title: Test Action
    shell: echo "test"
`

	tmpFile, err := os.CreateTemp("", "test_config_disabled_*.yaml")
	assert.NoError(t, err, "Should create temporary file")
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testConfig)
	assert.NoError(t, err, "Should write test config to file")
	tmpFile.Close()

	k := koanf.New(".")
	err = k.Load(file.Provider(tmpFile.Name()), yaml.Parser())
	assert.NoError(t, err, "Should load config file")

	cfg := &Config{}
	AppendSource(cfg, k, tmpFile.Name())

	assert.False(t, cfg.AuthLocalUsers.Enabled, "AuthLocalUsers should be disabled")
	assert.Equal(t, 1, len(cfg.AuthLocalUsers.Users), "Should still load users even when disabled")

	// User should still be findable even when auth is disabled
	user := cfg.FindUserByUsername("testuser")
	assert.NotNil(t, user, "Should find user even when auth is disabled")
}

func TestUserLoadingWithoutAuthSection(t *testing.T) {
	// Test config without authLocalUsers section
	testConfig := `
actions:
  - title: Test Action
    shell: echo "test"
`

	tmpFile, err := os.CreateTemp("", "test_config_no_auth_*.yaml")
	assert.NoError(t, err, "Should create temporary file")
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testConfig)
	assert.NoError(t, err, "Should write test config to file")
	tmpFile.Close()

	k := koanf.New(".")
	err = k.Load(file.Provider(tmpFile.Name()), yaml.Parser())
	assert.NoError(t, err, "Should load config file")

	cfg := &Config{}
	AppendSource(cfg, k, tmpFile.Name())

	// Should have default values
	assert.False(t, cfg.AuthLocalUsers.Enabled, "AuthLocalUsers should be disabled by default")
	assert.Equal(t, 0, len(cfg.AuthLocalUsers.Users), "Should have 0 users by default")
	assert.Nil(t, cfg.FindUserByUsername("anyuser"), "Should return nil for any user")
}

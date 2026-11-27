package otjwt

import (
	"crypto/rsa"
	"os"
	"path/filepath"
	"sync"
	"testing"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rotationKey struct {
	private *rsa.PrivateKey
	pubPem  []byte
	token   string
}

func newRotationKey(t *testing.T) rotationKey {
	t.Helper()

	privateKey, pubPem := generateRSAKeyPair(t)

	return rotationKey{
		private: privateKey,
		pubPem:  pubPem,
		token:   createJWTTokenWithExpirationAndAudience(t, privateKey, 1000, ""),
	}
}

func localKeyConfig(keyPath string) *config.Config {
	cfg := config.DefaultConfig()
	cfg.AuthJwtPubKeyPath = keyPath

	return cfg
}

func tokenAccepted(cfg *config.Config, token string) bool {
	parsed, err := parseJwtTokenWithLocalKey(cfg, token)

	return err == nil && parsed != nil && parsed.Valid
}

func writeKeyFile(t *testing.T, path string, content []byte) {
	t.Helper()

	require.NoError(t, os.WriteFile(path, content, 0o600))
}

func requireOnlyAccepted(t *testing.T, cfg *config.Config, accepted rotationKey, rejected rotationKey) {
	t.Helper()

	assert.True(t, tokenAccepted(cfg, accepted.token), "token signed by the current key must be accepted")
	assert.False(t, tokenAccepted(cfg, rejected.token), "token signed by the old key must be rejected")
}

func TestLocalKeyReplacedInPlaceIsUsedWithoutRestart(t *testing.T) {
	keyA, keyB := newRotationKey(t), newRotationKey(t)
	keyPath := filepath.Join(t.TempDir(), "jwt.pub")
	writeKeyFile(t, keyPath, keyA.pubPem)
	cfg := localKeyConfig(keyPath)

	requireOnlyAccepted(t, cfg, keyA, keyB)

	writeKeyFile(t, keyPath, keyB.pubPem)

	requireOnlyAccepted(t, cfg, keyB, keyA)
}

func TestLocalKeyReplacedByRenameIsUsedWithoutRestart(t *testing.T) {
	keyA, keyB := newRotationKey(t), newRotationKey(t)
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "jwt.pub")
	writeKeyFile(t, keyPath, keyA.pubPem)
	cfg := localKeyConfig(keyPath)

	requireOnlyAccepted(t, cfg, keyA, keyB)

	staged := filepath.Join(dir, "jwt.pub.new")
	writeKeyFile(t, staged, keyB.pubPem)
	require.NoError(t, os.Rename(staged, keyPath))

	requireOnlyAccepted(t, cfg, keyB, keyA)
}

// Kubernetes updates a mounted secret by pointing a symlink at a new directory.
func TestLocalKeyReplacedBySymlinkSwapIsUsedWithoutRestart(t *testing.T) {
	keyA, keyB := newRotationKey(t), newRotationKey(t)
	dir := t.TempDir()

	for name, key := range map[string]rotationKey{"..2026_a": keyA, "..2026_b": keyB} {
		require.NoError(t, os.Mkdir(filepath.Join(dir, name), 0o750))
		writeKeyFile(t, filepath.Join(dir, name, "jwt.pub"), key.pubPem)
	}

	require.NoError(t, os.Symlink("..2026_a", filepath.Join(dir, "..data")))
	require.NoError(t, os.Symlink(filepath.Join("..data", "jwt.pub"), filepath.Join(dir, "jwt.pub")))
	cfg := localKeyConfig(filepath.Join(dir, "jwt.pub"))

	requireOnlyAccepted(t, cfg, keyA, keyB)

	require.NoError(t, os.Symlink("..2026_b", filepath.Join(dir, "..data_tmp")))
	require.NoError(t, os.Rename(filepath.Join(dir, "..data_tmp"), filepath.Join(dir, "..data")))

	requireOnlyAccepted(t, cfg, keyB, keyA)
}

func TestLocalKeyDeletedRejectsEveryToken(t *testing.T) {
	keyA := newRotationKey(t)
	keyPath := filepath.Join(t.TempDir(), "jwt.pub")
	writeKeyFile(t, keyPath, keyA.pubPem)
	cfg := localKeyConfig(keyPath)

	require.True(t, tokenAccepted(cfg, keyA.token))

	require.NoError(t, os.Remove(keyPath))
	assert.False(t, tokenAccepted(cfg, keyA.token), "a deleted key file must not leave the old key trusted")

	writeKeyFile(t, keyPath, keyA.pubPem)
	assert.True(t, tokenAccepted(cfg, keyA.token), "restoring the key file must work without a restart")
}

func TestLocalKeyCorruptedRejectsEveryToken(t *testing.T) {
	keyA := newRotationKey(t)
	keyPath := filepath.Join(t.TempDir(), "jwt.pub")
	writeKeyFile(t, keyPath, keyA.pubPem)
	cfg := localKeyConfig(keyPath)

	require.True(t, tokenAccepted(cfg, keyA.token))

	writeKeyFile(t, keyPath, []byte("not a public key"))

	assert.False(t, tokenAccepted(cfg, keyA.token), "a corrupted key file must not leave the old key trusted")
}

func TestLocalKeyUnchangedFileReusesParsedKey(t *testing.T) {
	keyA := newRotationKey(t)
	keyPath := filepath.Join(t.TempDir(), "jwt.pub")
	writeKeyFile(t, keyPath, keyA.pubPem)

	first, err := localKey.get(keyPath)
	require.NoError(t, err)

	second, err := localKey.get(keyPath)
	require.NoError(t, err)

	assert.Same(t, first, second, "an unchanged key file must not be parsed again")
}

func TestLocalKeyRotationDuringConcurrentRequests(t *testing.T) {
	keyA, keyB := newRotationKey(t), newRotationKey(t)
	keyPath := filepath.Join(t.TempDir(), "jwt.pub")
	writeKeyFile(t, keyPath, keyA.pubPem)
	cfg := localKeyConfig(keyPath)

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 50 {
				tokenAccepted(cfg, keyA.token)
				tokenAccepted(cfg, keyB.token)
			}
		})
	}

	writeKeyFile(t, keyPath, keyB.pubPem)
	wg.Wait()

	requireOnlyAccepted(t, cfg, keyB, keyA)
}

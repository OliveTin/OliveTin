package otjwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitRemoteJwksKeepsBackgroundContext(t *testing.T) {
	// Smoke-check that production helper constructs a verifier without a
	// cancelled parent context. Behavioural refresh coverage is below.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	t.Cleanup(srv.Close)

	k, err := initRemoteJwks(srv.URL)
	require.NoError(t, err)
	require.NotNil(t, k)
}

func TestJwksRefreshContinuesWhenContextNotCancelled(t *testing.T) {
	ctx := context.Background()
	store := jwkset.NewMemoryStorage()

	oldPriv, oldPub := mustGenerateRSAKey(t)
	newPriv, newPub := mustGenerateRSAKey(t)

	require.NoError(t, store.KeyWrite(ctx, mustJWK(t, "old-kid", oldPub)))

	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		raw, err := store.JSONPublic(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(raw)
	}))
	t.Cleanup(srv.Close)

	refreshCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	verifier, err := keyfunc.NewDefaultOverrideCtx(refreshCtx, []string{srv.URL}, keyfunc.Override{
		RefreshInterval:         100 * time.Millisecond,
		RefreshErrorHandlerFunc: silentJWKSRefreshErrorHandler,
	})
	require.NoError(t, err)

	oldToken := mustSignJWT(t, oldPriv, "old-kid")
	_, err = jwt.Parse(oldToken, verifier.Keyfunc)
	require.NoError(t, err, "token signed by initial JWKS key must verify")

	mu.Lock()
	_, err = store.KeyDelete(ctx, "old-kid")
	require.NoError(t, err)
	require.NoError(t, store.KeyWrite(ctx, mustJWK(t, "new-kid", newPub)))
	mu.Unlock()

	require.Eventually(t, func() bool {
		newToken := mustSignJWT(t, newPriv, "new-kid")
		_, parseErr := jwt.Parse(newToken, verifier.Keyfunc)
		return parseErr == nil
	}, 3*time.Second, 100*time.Millisecond, "JWKS refresh must pick up rotated signing keys")

	_, err = jwt.Parse(oldToken, verifier.Keyfunc)
	assert.Error(t, err, "token signed by revoked key should stop verifying after refresh")
}

func TestJwksRefreshStopsWhenContextCancelled(t *testing.T) {
	ctx := context.Background()
	store := jwkset.NewMemoryStorage()

	oldPriv, oldPub := mustGenerateRSAKey(t)
	newPriv, newPub := mustGenerateRSAKey(t)

	require.NoError(t, store.KeyWrite(ctx, mustJWK(t, "old-kid", oldPub)))

	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		raw, err := store.JSONPublic(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(raw)
	}))
	t.Cleanup(srv.Close)

	refreshCtx, cancel := context.WithCancel(context.Background())
	verifier, err := keyfunc.NewDefaultOverrideCtx(refreshCtx, []string{srv.URL}, keyfunc.Override{
		RefreshInterval:         100 * time.Millisecond,
		RefreshErrorHandlerFunc: silentJWKSRefreshErrorHandler,
	})
	require.NoError(t, err)
	cancel() // reproduces the pre-fix WithTimeout/defer-cancel bug

	oldToken := mustSignJWT(t, oldPriv, "old-kid")
	_, err = jwt.Parse(oldToken, verifier.Keyfunc)
	require.NoError(t, err)

	mu.Lock()
	_, err = store.KeyDelete(ctx, "old-kid")
	require.NoError(t, err)
	require.NoError(t, store.KeyWrite(ctx, mustJWK(t, "new-kid", newPub)))
	mu.Unlock()

	time.Sleep(500 * time.Millisecond)

	newToken := mustSignJWT(t, newPriv, "new-kid")
	_, err = jwt.Parse(newToken, verifier.Keyfunc)
	assert.Error(t, err, "cancelled JWKS context must not refresh to the rotated key")
}

func silentJWKSRefreshErrorHandler(string) func(context.Context, error) {
	return func(context.Context, error) {}
}

func mustGenerateRSAKey(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey, &privateKey.PublicKey
}

func mustJWK(t *testing.T, kid string, pub *rsa.PublicKey) jwkset.JWK {
	t.Helper()
	jwk, err := jwkset.NewJWKFromKey(pub, jwkset.JWKOptions{
		Metadata: jwkset.JWKMetadataOptions{KID: kid},
	})
	require.NoError(t, err)
	return jwk
}

func mustSignJWT(t *testing.T, privateKey *rsa.PrivateKey, kid string) string {
	t.Helper()
	token := jwt.New(jwt.SigningMethodRS256)
	token.Header["kid"] = kid
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = "test"
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	signed, err := token.SignedString(privateKey)
	require.NoError(t, err)
	return signed
}

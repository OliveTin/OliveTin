package otjwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"

	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateRSAKeyPair(t *testing.T) (*rsa.PrivateKey, []byte) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	pubKey := &privateKey.PublicKey
	pkixPubKey, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}

	pubPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pkixPubKey,
		},
	)

	return privateKey, pubPem
}

func createKeys(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()

	tmpFile, err := os.CreateTemp(os.TempDir(), "olivetin-jwt-")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = tmpFile.Close() }()

	t.Logf("Created File: %s", tmpFile.Name())

	privateKey, pubPem := generateRSAKeyPair(t)

	if err := os.WriteFile(tmpFile.Name(), pubPem, 0644); err != nil {
		t.Fatalf("error when dumping pubKey: %s \n", err)
	}

	return privateKey, tmpFile.Name()
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	return mux
}

func createJWTTokenWithExpirationAndAudience(t *testing.T, privateKey *rsa.PrivateKey, expire int64, audience string) string {
	t.Helper()

	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["nbf"] = time.Now().Unix() - 1000
	claims["exp"] = time.Now().Unix() + expire
	claims["sub"] = "test"
	claims["olivetinGroup"] = "test"
	if audience != "" {
		claims["aud"] = audience
	}

	tokenStr, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign JWT token: %v", err)
	}
	return tokenStr
}

func setupJWTTestHandler(t *testing.T, cfg *config.Config) http.Handler {
	t.Helper()

	mux := newMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		context := &authpublic.AuthCheckingContext{
			Request: r,
			Config:  cfg,
		}
		user := CheckUserFromJwtHeader(context)

		if user == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		assert.Equal(t, "test", user.Username)
		assert.Equal(t, "test", user.UsergroupLine)
	})
	return mux
}

func verifyJWTResponse(t *testing.T, res *http.Response, expectCode int) {
	t.Helper()

	defer func() { _ = res.Body.Close() }()
	assert.Equal(t, expectCode, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err, "reading JWT response body")
	t.Logf("Response body: %s", string(body))
}

func testJwkValidation(t *testing.T, expire int64, expectCode int) {
	t.Helper()

	testJwkValidationWithAudience(t, expire, expectCode, "", "")
}

func testJwkValidationWithAudience(t *testing.T, expire int64, expectCode int, configAudience, tokenAudience string) {
	t.Helper()

	privateKey, publicKeyPath := createKeys(t)
	defer func() { _ = os.Remove(publicKeyPath) }()

	cfg := config.DefaultConfig()
	cfg.AuthJwtPubKeyPath = publicKeyPath
	cfg.AuthJwtClaimUsername = "sub"
	cfg.AuthJwtClaimUserGroup = "olivetinGroup"
	cfg.AuthJwtHeader = "Authorization"
	cfg.AuthJwtAud = configAudience

	tokenStr := createJWTTokenWithExpirationAndAudience(t, privateKey, expire, tokenAudience)
	handler := setupJWTTestHandler(t, cfg)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	res := makeJWTRequest(t, srv, tokenStr) //nolint:bodyclose // closed by verifyJWTResponse

	verifyJWTResponse(t, res, expectCode)
}

func TestJWTSignatureVerificationSucceeds(t *testing.T) {
	testJwkValidation(t, 1000, http.StatusOK)
}

func TestJWTSignatureVerificationFails(t *testing.T) {
	testJwkValidation(t, -500, http.StatusForbidden)
}

func TestJWTAudienceValidationRejectsWrongAudience(t *testing.T) {
	testJwkValidationWithAudience(t, 1000, http.StatusForbidden, "expected-audience", "wrong-audience")
}

func TestJWTAudienceValidationAcceptsCorrectAudience(t *testing.T) {
	testJwkValidationWithAudience(t, 1000, http.StatusOK, "expected-audience", "expected-audience")
}

func createJWTTokenWithGroups(t *testing.T, privateKey *rsa.PrivateKey, groups any) string {
	t.Helper()

	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["nbf"] = time.Now().Unix() - 1000
	claims["exp"] = time.Now().Unix() + 2000
	claims["sub"] = "test"
	claims["olivetinGroup"] = groups

	tokenStr, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign JWT token: %v", err)
	}
	return tokenStr
}

func makeJWTRequest(t *testing.T, srv *httptest.Server, tokenStr string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Client err: %+v", err)
	}

	return res
}

func TestJWTHeader(t *testing.T) {
	privateKey, publicKeyPath := createKeys(t)
	defer func() { _ = os.Remove(publicKeyPath) }()

	cfg := config.DefaultConfig()
	cfg.AuthJwtPubKeyPath = publicKeyPath
	cfg.AuthJwtClaimUsername = "sub"
	cfg.AuthJwtClaimUserGroup = "olivetinGroup"
	cfg.AuthJwtHeader = "Authorization"

	tokenStr := createJWTTokenWithGroups(t, privateKey, []string{"test", "test2"})

	mux := newMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		context := &authpublic.AuthCheckingContext{
			Request: r,
			Config:  cfg,
		}
		user := CheckUserFromJwtHeader(context)

		if user == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		assert.Equal(t, "test", user.Username)
		assert.Equal(t, "test test2", user.UsergroupLine)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	res := makeJWTRequest(t, srv, tokenStr) //nolint:bodyclose // closed by verifyJWTResponse
	verifyJWTResponse(t, res, http.StatusOK)
}

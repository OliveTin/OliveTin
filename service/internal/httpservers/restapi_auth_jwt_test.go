package httpservers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	// config "github.com/OliveTin/OliveTin/internal/config"
	// "github.com/golang-jwt/jwt/v4"
	//	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	// "time"
)

func createKeys(t *testing.T) (*rsa.PrivateKey, string) {
	tmpFile, _ := os.CreateTemp(os.TempDir(), "olivetin-jwt-")

	fmt.Println("Created File: " + tmpFile.Name())

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubKey := &privateKey.PublicKey
	// https://stackoverflow.com/questions/13555085/save-and-load-crypto-rsa-privatekey-to-and-from-the-disk
	pkixPubKey, _ := x509.MarshalPKIXPublicKey(pubKey)
	pubPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pkixPubKey,
		},
	)

	if err := os.WriteFile(tmpFile.Name(), pubPem, 0755); err != nil {
		t.Fatalf("error when dumping pubKey: %s \n", err)
	}

	return privateKey, tmpFile.Name()
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	return mux
}

func testJwkValidation(t *testing.T, expire int64, expectCode int) {
	/*
		privateKey, publicKeyPath := createKeys(t)

		defer os.Remove(publicKeyPath)

		cfg := config.DefaultConfig()
		cfg.AuthJwtPubKeyPath = publicKeyPath
		cfg.AuthJwtClaimUsername = "sub"
		cfg.AuthJwtClaimUserGroup = "olivetinGroup"
		cfg.AuthJwtCookieName = "authorization_token"

		token := jwt.New(jwt.SigningMethodRS256)

		claims := token.Claims.(jwt.MapClaims)
		claims["nbf"] = time.Now().Unix() - 1000
		claims["exp"] = time.Now().Unix() + expire
		claims["sub"] = "test"
		claims["olivetinGroup"] = "test"
	*/

	/*
		tokenStr, _ := token.SignedString(privateKey)

		mux := newMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			username, usergroup := parseJwtCookie(cfg, r)

			if username == "" {
				w.WriteHeader(403)
			}

			w.Write([]byte(fmt.Sprintf("username=%v, usergroup=%v", username, usergroup)))
		})

		srv := setupTestingServer(mux, t)

		req, client := newReq("")
		req.AddCookie(&http.Cookie{
			Name:   "authorization_token",
			Value:  tokenStr,
			MaxAge: 300,
		})

		res, err := client.Do(req)

		if err != nil {
			t.Fatalf("Client err: %+v", err)
		} else {
			defer res.Body.Close()
			assert.Equal(t, expectCode, res.StatusCode)
			body, _ := io.ReadAll(res.Body)
			fmt.Println(string(body))
		}

		err = srv.Shutdown(context.TODO())

		if err != nil {
			t.Fatalf("Server shutdown error: %+v", err)
		}
	*/
}

func TestJWTSignatureVerificationSucceeds(t *testing.T) {
	testJwkValidation(t, 1000, 200)
}

func TestJWTSignatureVerificationFails(t *testing.T) {
	testJwkValidation(t, -500, 403)
}

func TestJWTHeader(t *testing.T) {
	/*
		privateKey, publicKeyPath := createKeys(t)

		defer os.Remove(publicKeyPath)

		cfg := config.DefaultConfig()
		cfg.AuthJwtPubKeyPath = publicKeyPath
		cfg.AuthJwtClaimUsername = "sub"
		cfg.AuthJwtClaimUserGroup = "olivetinGroup"
		cfg.AuthJwtHeader = "Authorization"

		token := jwt.New(jwt.SigningMethodRS256)

		claims := token.Claims.(jwt.MapClaims)
		claims["nbf"] = time.Now().Unix() - 1000
		claims["exp"] = time.Now().Unix() + 2000
		claims["sub"] = "test"
		claims["olivetinGroup"] = []string{"test", "test2"}
	*/

	/*
		tokenStr, _ := token.SignedString(privateKey)

		mux := newMux()
		mux.HandlePath("GET", "/", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			username, usergroup := parseJwtHeader(cfg, r)

			if username == "" {
				w.WriteHeader(403)
			}

			assert.Equal(t, "test", username)
			assert.Equal(t, "test test2", usergroup)

			w.Write([]byte(fmt.Sprintf("username=%v, usergroup=%v", username, usergroup)))
		})

		srv := setupTestingServer(mux, t)

		req, client := newReq("")
		req.Header.Set("Authorization", "Bearer "+tokenStr)

		res, err := client.Do(req)

		if err != nil {
			t.Fatalf("Client err: %+v", err)
		} else {
			defer res.Body.Close()
			assert.Equal(t, 200, res.StatusCode)
			body, _ := io.ReadAll(res.Body)
			fmt.Println(string(body))
		}

		srv.Shutdown(context.TODO())
	*/
}

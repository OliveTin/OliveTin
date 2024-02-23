package httpservers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/cors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func createKeys() (*rsa.PrivateKey, string) {
	tmpFile, _ := os.CreateTemp(os.TempDir(), "olivetin-jwt-")
	defer os.Remove(tmpFile.Name())

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
		fmt.Printf("error when dumping pubKey: %s \n", err)
	}

	return privateKey, tmpFile.Name()
}

func testBase(t *testing.T, expire int64, expectCode int) {
	privateKey, publicKeyPath := createKeys()

	// default config + overrides
	cfg := config.DefaultConfig()
	cfg.AuthJwtPubKeyPath = publicKeyPath
	cfg.AuthJwtClaimUsername = "sub"
	cfg.AuthJwtClaimUserGroup = "olivetinGroup"
	cfg.AuthJwtCookieName = "authorization_token"
	SetGlobalRestConfig(cfg) // ugly, setting global var, we should pass configs as params to modules... :/

	token := jwt.New(jwt.SigningMethodRS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["nbf"] = time.Now().Unix() - 1000
	claims["exp"] = time.Now().Unix() + expire
	claims["sub"] = "test"
	claims["olivetinGroup"] = "test"

	tokenStr, _ := token.SignedString(privateKey)

	// init mux endpoint like in restapi.go (but using dummy response handler)
	mux := runtime.NewServeMux(
		runtime.WithMetadata(parseRequestMetadata), // i am guessing this is critical middleware for authorizing request cookie
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
			},
		}),
	)
	mux.HandlePath("GET", "/", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		username, usergroup := parseJwtCookie(r)
		if username == "" {
			w.WriteHeader(403)
		}
		w.Write([]byte(fmt.Sprintf("username=%v, usergroup=%v", username, usergroup)))
	})

	// make server and attach handler
	setupTestingServer(mux, t)

	// make http client and send request to myself
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:1337/", nil)
	cookie := &http.Cookie{
		Name:   "authorization_token",
		Value:  tokenStr,
		MaxAge: 300,
	}
	req.AddCookie(cookie)
	res, err := client.Do(req)

	if err != nil {
		assert.Equal(t, expectCode, -1)
	} else {
		defer res.Body.Close()
		assert.Equal(t, expectCode, res.StatusCode)
		body, _ := io.ReadAll(res.Body)
		fmt.Println(string(body))
	}
}

func setupTestingServer(mux *runtime.ServeMux, t *testing.T) {
	lis, err := net.Listen("tcp", ":1337")

	if err != nil || lis == nil {
		t.Errorf("Could not listen %v %v", err, lis)
		return
	}

	srv := &http.Server{Handler: cors.AllowCors(mux)}

	go startTestingServer(lis, srv, t)
}

func startTestingServer(lis net.Listener, srv *http.Server, t *testing.T) {
	if srv == nil {
		t.Errorf("srv is nil. Could not listen")
		return
	}

	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Errorf("couldn't start server: %v", err)
		}
	}()
}

func TestJWTSignatureVerificationSucceeds(t *testing.T) {
	// testBase(t, 1000, 200)
}

func TestJWTSignatureVerificationFails(t *testing.T) {
	testBase(t, -500, 403)
}

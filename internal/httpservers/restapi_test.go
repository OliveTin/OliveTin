package httpservers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	config2 "github.com/OliveTin/OliveTin/internal/config"
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

func TestJWTSignatureVerificationSucceeds(t *testing.T) {
	tmpFile, _ := os.CreateTemp(os.TempDir(), "olivetin-jwt-")
	//defer os.Remove(tmpFile.Name())

	fmt.Println("Created File: " + tmpFile.Name())

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubKey := &privateKey.PublicKey
	pubPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pubKey),
		},
	)
	if err := os.WriteFile(tmpFile.Name(), pubPem, 0755); err != nil {
		fmt.Printf("error when dumping pubKey: %s \n", err)
	}
	// default config + overrides, don't know how to pass it to rest server to use it
	config := config2.DefaultConfig()
	config.AuthJwtPubKeyPath = tmpFile.Name()
	config.AuthJwtClaimUsername = "sub"
	config.AuthJwtClaimUserGroup = "olivetinGroup"
	config.AuthJwtCookieName = "authorization_token"
	SetGlobalRestConfig(config) // ugly, setting global var, we should pass configs as params to modules... :/

	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(10 * time.Minute)
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
		w.Write([]byte(fmt.Sprintf("username=%v, usergroup=%v", username, usergroup)))
	})

	// make server and attach handler
	srv := &http.Server{Handler: cors.AllowCors(mux)}
	lis, _ := net.Listen("tcp", ":1337")

	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Errorf("couldn't start server: %v", err)
		}
	}()

	// make http client and send request to myself
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:1337/", nil)
	cookie := &http.Cookie{
		Name:   "authorization_token",
		Value:  tokenStr,
		MaxAge: 300,
	}
	req.AddCookie(cookie)
	res, _ := client.Do(req)
	defer res.Body.Close()
	assert.Equal(t, 200, res.StatusCode)
	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))
}

func TestJWTSignatureVerificationFails(t *testing.T) {

}

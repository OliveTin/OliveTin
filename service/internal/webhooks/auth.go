package webhooks

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
)

type AuthVerifier struct {
	config config.WebhookConfig
}

func NewAuthVerifier(cfg config.WebhookConfig) *AuthVerifier {
	return &AuthVerifier{config: cfg}
}

func (v *AuthVerifier) Verify(r *http.Request, payload []byte) bool {
	switch v.config.AuthType {
	case "hmac-sha256":
		return v.verifyHMAC256(r, payload)
	case "hmac-sha1":
		return v.verifyHMAC1(r, payload)
	case "bearer":
		return v.verifyBearer(r)
	case "basic":
		return v.verifyBasic(r)
	case "none", "":
		return true
	default:
		log.WithFields(log.Fields{
			"authType": v.config.AuthType,
		}).Warnf("Unknown auth type, rejecting")
		return false
	}
}

func (v *AuthVerifier) verifyHMAC256(r *http.Request, payload []byte) bool {
	if v.config.Secret == "" {
		log.Warnf("HMAC-SHA256 auth requires secret")
		return false
	}

	headerName := v.config.AuthHeader
	if headerName == "" {
		headerName = "X-Webhook-Signature"
	}

	signature := r.Header.Get(headerName)
	if signature == "" {
		log.Debugf("Missing signature header: %s", headerName)
		return false
	}

	expectedSig := strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(v.config.Secret))
	mac.Write(payload)
	computedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(computedSig))
}

func (v *AuthVerifier) verifyHMAC1(r *http.Request, payload []byte) bool {
	if v.config.Secret == "" {
		log.Warnf("HMAC-SHA1 auth requires secret")
		return false
	}

	headerName := v.config.AuthHeader
	if headerName == "" {
		headerName = "X-Webhook-Signature"
	}

	signature := r.Header.Get(headerName)
	if signature == "" {
		log.Debugf("Missing signature header: %s", headerName)
		return false
	}

	expectedSig := strings.TrimPrefix(signature, "sha1=")

	mac := hmac.New(sha1.New, []byte(v.config.Secret))
	mac.Write(payload)
	computedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(computedSig))
}

func (v *AuthVerifier) verifyBearer(r *http.Request) bool {
	if v.config.Secret == "" {
		log.Warnf("Bearer auth requires secret")
		return false
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		log.Debugf("Missing or invalid Bearer token")
		return false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token == v.config.Secret
}

func (v *AuthVerifier) verifyBasic(r *http.Request) bool {
	if v.config.Secret == "" {
		log.Warnf("Basic auth requires secret")
		return false
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		log.Debugf("Missing Basic auth header")
		return false
	}

	parts := strings.SplitN(v.config.Secret, ":", 2)
	if len(parts) == 2 {
		return username == parts[0] && password == parts[1]
	}

	return password == v.config.Secret
}

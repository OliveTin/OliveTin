package authpublic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionFingerprintIsEmptyWithoutSession(t *testing.T) {
	t.Parallel()

	user := &AuthenticatedUser{Username: "guest"}

	assert.Empty(t, user.SessionFingerprint())
}

func TestSessionFingerprintDoesNotContainSessionID(t *testing.T) {
	t.Parallel()

	sid := "3f2b1c7a-8d4e-4f1a-9b6c-2e5d7a9c1b3f"
	user := &AuthenticatedUser{SID: sid}

	fingerprint := user.SessionFingerprint()

	assert.Len(t, fingerprint, sessionFingerprintLength)
	assert.NotContains(t, sid, fingerprint)
	assert.NotContains(t, fingerprint, sid)
}

func TestSessionFingerprintIsStablePerSession(t *testing.T) {
	t.Parallel()

	first := &AuthenticatedUser{SID: "session-a"}
	again := &AuthenticatedUser{SID: "session-a"}
	other := &AuthenticatedUser{SID: "session-b"}

	assert.Equal(t, first.SessionFingerprint(), again.SessionFingerprint())
	assert.NotEqual(t, first.SessionFingerprint(), other.SessionFingerprint())
}

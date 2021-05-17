package grpcapi

import (
	"github.com/stretchr/testify/assert"
	"testing"

	config "github.com/jamesread/OliveTin/internal/config"
)

func TestCreateApi(t *testing.T) {
	api := newServer()

	assert.NotNil(t, api, "Create new server")

	cfg = config.DefaultConfig()
}

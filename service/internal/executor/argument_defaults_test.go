package executor

import (
	"testing"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateArgumentDefaultSkipsTemplates(t *testing.T) {
	action := &config.Action{Title: "Deploy", ID: "deploy"}
	arg := &config.ActionArgument{
		Name:    "host",
		Type:    "ascii_identifier",
		Default: `{{ host.name }}`,
	}

	assert.Nil(t, validateArgumentDefault(action, arg))
}

func TestValidateArgumentDefaultLiteralInvalid(t *testing.T) {
	action := &config.Action{Title: "Deploy", ID: "deploy"}
	arg := &config.ActionArgument{
		Name:    "host",
		Type:    "ascii_identifier",
		Default: "not a valid id!!",
	}

	issue := validateArgumentDefault(action, arg)
	require.NotNil(t, issue)
	assert.Equal(t, configissues.CodeArgDefaultInvalid, issue.Code)
}

func TestValidateArgumentDefaultLiteralValid(t *testing.T) {
	action := &config.Action{Title: "Deploy", ID: "deploy"}
	arg := &config.ActionArgument{
		Name:    "host",
		Type:    "ascii_identifier",
		Default: "example.com",
	}

	assert.Nil(t, validateArgumentDefault(action, arg))
}

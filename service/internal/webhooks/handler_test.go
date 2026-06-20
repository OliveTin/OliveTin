package webhooks

import (
	"testing"

	"github.com/stretchr/testify/assert"

	config "github.com/OliveTin/OliveTin/internal/config"
)

func TestFilterToDefinedArguments(t *testing.T) {
	action := &config.Action{
		Arguments: []config.ActionArgument{
			{Name: "repo", Type: "ascii_identifier"},
			{Name: "branch", Type: "ascii_identifier"},
		},
	}
	args := map[string]string{
		"repo":                    "my-repo",
		"branch":                  "main",
		"webhook_path":            "/deploy/prod",
		"webhook_header_x_custom": "malicious",
	}

	filtered := filterToDefinedArguments(args, action)

	assert.Equal(t, "my-repo", filtered["repo"])
	assert.Equal(t, "main", filtered["branch"])
	assert.Empty(t, filtered["webhook_path"])
	assert.Empty(t, filtered["webhook_header_x_custom"])
	assert.Len(t, filtered, 2)
}

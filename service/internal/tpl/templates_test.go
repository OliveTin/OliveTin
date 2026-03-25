package tpl

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/stretchr/testify/assert"
)

type templateJsonCase struct {
	name           string
	source         string
	ent            *entities.Entity
	args           map[string]any
	expectedOutput string
	expectError    bool
	checkJsonOnly  bool
}

func (tt templateJsonCase) run(t *testing.T) {
	output, err := ParseTemplateWithActionContext(tt.source, tt.ent, tt.args)
	if tt.expectError {
		assert.Error(t, err)
		return
	}
	assert.NoError(t, err)
	if tt.checkJsonOnly {
		strArgs := make(map[string]string)
		for k, v := range tt.args {
			strArgs[k] = fmt.Sprintf("%v", v)
		}
		assertJsonOutput(t, output, tt.expectedOutput, strArgs)
		return
	}
	assert.Equal(t, tt.expectedOutput, output)
}

func TestParseTemplateWithActionContext_Json(t *testing.T) {
	tests := []templateJsonCase{
		{
			name:           "Arguments piped to Json",
			source:         `echo {{ .Arguments | Json }}`,
			ent:            nil,
			args:           map[string]any{"value": "true", "ot_username": "alice"},
			expectedOutput: `echo `,
			expectError:    false,
			checkJsonOnly:  true,
		},
		{
			name:           "CurrentEntity field piped to Json",
			source:         `curl -d {{ .CurrentEntity.foo.bar | Json }}`,
			ent:            &entities.Entity{Data: map[string]any{"foo": map[string]any{"bar": "baz"}}},
			args:           nil,
			expectedOutput: `curl -d "baz"`,
			expectError:    false,
		},
		{
			name:           "CurrentEntity nested object piped to Json",
			source:         `curl --json -d {{ .CurrentEntity.payload | Json }}`,
			ent:            &entities.Entity{Data: map[string]any{"payload": map[string]any{"on": true}}},
			args:           nil,
			expectedOutput: `curl --json -d {"on":true}`,
			expectError:    false,
		},
		{
			name:           "Single argument value as Json",
			source:         `echo {{ .Arguments.value | Json }}`,
			ent:            nil,
			args:           map[string]any{"value": "hello"},
			expectedOutput: `echo "hello"`,
			expectError:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func assertJsonOutput(t *testing.T, output, expectedPrefix string, args map[string]string) {
	t.Helper()
	prefix := strings.TrimSuffix(expectedPrefix, " ")
	assert.True(t, strings.HasPrefix(output, prefix), "output %q should start with %q", output, prefix)
	jsonPart := strings.TrimPrefix(output, prefix)
	jsonPart = strings.TrimSpace(jsonPart)
	var decoded map[string]string
	err := json.Unmarshal([]byte(jsonPart), &decoded)
	assert.NoError(t, err)
	for k, v := range args {
		assert.Equal(t, v, decoded[k], "decoded JSON should contain %s=%s", k, v)
	}
	assert.Len(t, decoded, len(args))
}

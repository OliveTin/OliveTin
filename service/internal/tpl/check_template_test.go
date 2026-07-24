package tpl_test

import (
	"testing"

	"github.com/OliveTin/OliveTin/internal/tpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckActionTemplateParses(t *testing.T) {
	assert.NoError(t, tpl.CheckActionTemplateParses(`{{ host.name }}`))
	assert.NoError(t, tpl.CheckActionTemplateParses(`{{ host.hostname }}`))

	err := tpl.CheckActionTemplateParses(`{{ host.name | default "All" }}`)
	require.Error(t, err)
}

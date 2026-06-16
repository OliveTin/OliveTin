package servicehost

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveLogDirectory(t *testing.T) {
	t.Parallel()

	baseDir := filepath.Join(t.TempDir(), "OliveTin")
	absoluteDir := t.TempDir()

	assert.Equal(t, "", resolveLogDirectory("", baseDir))
	assert.Equal(t, absoluteDir, resolveLogDirectory(absoluteDir, baseDir))
	assert.Equal(t, filepath.Join(baseDir, "logs", "service"), resolveLogDirectory("./logs/service", baseDir))
	assert.Equal(t, "logs/service", resolveLogDirectory("logs/service", ""))
}

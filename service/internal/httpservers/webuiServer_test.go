package httpservers

import (
	"os"
	"testing"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGetWebuiDir(t *testing.T) {
	originalDir, err := os.Getwd()
	assert.Equal(t, err, nil)
	os.Chdir("../../") // go test sets the cwd to "httpservers" by default
	defer os.Chdir(originalDir)

	cfg = config.DefaultConfig()

	dir := findWebuiDir()

	assert.Equal(t, "../webui/", dir, "Finding the webui dir")
}

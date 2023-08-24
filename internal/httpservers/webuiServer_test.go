package httpservers

import (
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetWebuiDir(t *testing.T) {
	os.Chdir("../../") // go test sets the cwd to "httpservers" by default

	cfg = config.DefaultConfig()

	dir := findWebuiDir()

	assert.Equal(t, "./webui", dir, "Finding the webui dir")
}

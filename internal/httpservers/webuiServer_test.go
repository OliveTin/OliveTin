package httpservers

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetWebuiDir(t *testing.T) {
	os.Chdir("../../") // go test sets the cwd to "httpservers" by default

	dir := findWebuiDir()

	assert.Equal(t, "./webui", dir, "Finding the webui dir")
}

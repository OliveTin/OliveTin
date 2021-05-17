package webuiServer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetWebuiDir(t *testing.T) {
	dir := findWebuiDir()

	assert.Equal(t, "./webui", dir, "Finding the webui dir")
}

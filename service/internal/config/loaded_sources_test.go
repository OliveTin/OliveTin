package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadedSourcesRecordsUniquePaths(t *testing.T) {
	BeginLoadedSources()
	assert.Empty(t, LoadedSources())

	RecordLoadedSource("/tmp/config.yaml")
	RecordLoadedSource("/tmp/config.d/a.yaml")
	RecordLoadedSource("/tmp/config.yaml")
	RecordLoadedSource("")

	assert.Equal(t, []string{"/tmp/config.yaml", "/tmp/config.d/a.yaml"}, LoadedSources())

	BeginLoadedSources()
	assert.Empty(t, LoadedSources())
}

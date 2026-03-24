package config

import (
	"testing"

	"github.com/dustin/go-humanize"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultFileUploadMaxBytesMatchesHumanize(t *testing.T) {
	n, err := humanize.ParseBytes("10 MB")
	require.NoError(t, err)
	assert.Equal(t, int64(n), int64(DefaultFileUploadMaxBytes))
}

func TestUnmarshalFileUploadsHumanReadableBytes(t *testing.T) {
	yamlText := `
fileUploads:
  maxBytes: "512 kb"
actions:
  - title: x
    shell: echo
    arguments:
      - name: f
        type: file_upload
        maxUploadBytes: "2 MiB"
`
	k := koanf.New(".")
	require.NoError(t, k.Load(rawbytes.Provider([]byte(yamlText)), yaml.Parser()))
	cfg := &Config{}
	require.NoError(t, k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{
		Tag:           "koanf",
		DecoderConfig: newDefaultUnmarshalDecoderConfig(),
	}))

	assert.Equal(t, HumanReadableBytes(512000), cfg.FileUploads.MaxBytes)
	require.Len(t, cfg.Actions, 1)
	require.Len(t, cfg.Actions[0].Arguments, 1)
	assert.Equal(t, HumanReadableBytes(2*1024*1024), cfg.Actions[0].Arguments[0].MaxUploadBytes)
}

func TestUnmarshalFileUploadsMaxBytesNumericStillWorks(t *testing.T) {
	yamlText := `
fileUploads:
  maxBytes: 1048576
`
	k := koanf.New(".")
	require.NoError(t, k.Load(rawbytes.Provider([]byte(yamlText)), yaml.Parser()))
	cfg := &Config{}
	require.NoError(t, k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{
		Tag:           "koanf",
		DecoderConfig: newDefaultUnmarshalDecoderConfig(),
	}))

	assert.Equal(t, HumanReadableBytes(1048576), cfg.FileUploads.MaxBytes)
}

package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dustin/go-humanize"
	mapstructure "github.com/go-viper/mapstructure/v2"
)

// HumanReadableBytes is a byte size from config: a plain number (bytes) or a string
// parsed with github.com/dustin/go-humanize (e.g. "10 MB", "1 GiB", "512 kb").
type HumanReadableBytes int64

// UnmarshalText implements encoding.TextUnmarshaler for YAML string values.
func (h *HumanReadableBytes) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		*h = 0
		return nil
	}
	n, err := humanize.ParseBytes(s)
	if err != nil {
		return fmt.Errorf("parse file size: %w", err)
	}
	*h = HumanReadableBytes(n)
	return nil
}

func humanReadableBytesNumericHook() mapstructure.DecodeHookFunc {
	var zero HumanReadableBytes
	target := reflect.TypeOf(zero)
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if to != target {
			return data, nil
		}
		if v, ok := humanReadableBytesFromDecodedValue(data); ok {
			return v, nil
		}
		return data, nil
	}
}

func humanReadableBytesFromDecodedValue(data any) (HumanReadableBytes, bool) {
	v, ok := ptrHumanReadableBytes(data)
	if ok {
		return v, true
	}
	if v, ok := data.(HumanReadableBytes); ok {
		return v, true
	}
	n, ok := int64FromYAMLNumber(data)
	if !ok {
		return 0, false
	}
	return HumanReadableBytes(n), true
}

func ptrHumanReadableBytes(data any) (HumanReadableBytes, bool) {
	p, ok := data.(*HumanReadableBytes)
	if !ok {
		return 0, false
	}
	if p == nil {
		return 0, true
	}
	return *p, true
}

func int64FromYAMLNumber(data any) (int64, bool) {
	if v, ok := int64FromFloatOrInt(data); ok {
		return v, true
	}
	if v, ok := data.(int64); ok {
		return v, true
	}
	if v, ok := data.(uint64); ok {
		return int64(v), true
	}
	return 0, false
}

func int64FromFloatOrInt(data any) (int64, bool) {
	if v, ok := data.(float64); ok {
		return int64(v), true
	}
	if v, ok := data.(int); ok {
		return int64(v), true
	}
	return 0, false
}

func newDefaultUnmarshalDecoderConfig() *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.TextUnmarshallerHookFunc(),
			humanReadableBytesNumericHook(),
		),
		WeaklyTypedInput: true,
	}
}

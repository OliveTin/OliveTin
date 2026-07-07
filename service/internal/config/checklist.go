package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseChecklistValue parses a checklist argument wire value.
// New values are JSON arrays; legacy comma-separated values are still accepted.
func ParseChecklistValue(value string) ([]string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	if strings.HasPrefix(trimmed, "[") {
		return parseJSONChecklistValue(trimmed)
	}

	return parseLegacyChecklistValue(value)
}

func parseJSONChecklistValue(value string) ([]string, error) {
	var values []string
	if err := json.Unmarshal([]byte(value), &values); err != nil {
		return nil, fmt.Errorf("invalid checklist JSON value: %w", err)
	}

	return values, nil
}

func parseLegacyChecklistValue(value string) ([]string, error) {
	segments := strings.Split(value, ",")
	values := make([]string, 0, len(segments))
	for _, segment := range segments {
		trimmedSegment := strings.TrimSpace(segment)
		if trimmedSegment == "" {
			return nil, fmt.Errorf("checklist value contains an empty segment")
		}

		values = append(values, trimmedSegment)
	}

	return values, nil
}

// FormatChecklistValue serializes selected checklist values for API transport.
func FormatChecklistValue(values []string) string {
	if len(values) == 0 {
		return ""
	}

	encoded, err := json.Marshal(values)
	if err != nil {
		return ""
	}

	return string(encoded)
}

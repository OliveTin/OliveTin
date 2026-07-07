package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseChecklistValue parses a checklist argument wire value.
// Values must be JSON arrays, or a single choice without commas.
func ParseChecklistValue(value string) ([]string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	if strings.HasPrefix(trimmed, "[") {
		return parseJSONChecklistValue(trimmed)
	}

	if strings.Contains(trimmed, ",") {
		return nil, fmt.Errorf("checklist value uses legacy comma-separated format; use a JSON array instead")
	}

	return []string{trimmed}, nil
}

func parseJSONChecklistValue(value string) ([]string, error) {
	var values []string
	if err := json.Unmarshal([]byte(value), &values); err != nil {
		return nil, fmt.Errorf("invalid checklist JSON value: %w", err)
	}

	for _, segment := range values {
		if strings.TrimSpace(segment) == "" {
			return nil, fmt.Errorf("checklist value contains an empty segment")
		}
	}

	return values, nil
}

// FormatChecklistValue serializes selected checklist values for API transport.
func FormatChecklistValue(values []string) (string, error) {
	if len(values) == 0 {
		return "", nil
	}

	encoded, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("encoding checklist value: %w", err)
	}

	return string(encoded), nil
}

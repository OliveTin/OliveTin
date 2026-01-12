package webhooks

import (
	"encoding/json"
	"fmt"

	"github.com/PaesslerAG/jsonpath"
)

type JSONMatcher struct {
	payload interface{}
}

func NewJSONMatcher(payload []byte) (*JSONMatcher, error) {
	var data interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}
	return &JSONMatcher{payload: data}, nil
}

func (m *JSONMatcher) MatchPath(pathExpr string, expectedValue string) (bool, error) {
	value, err := jsonpath.Get(pathExpr, m.payload)
	if err != nil {
		return false, err
	}

	// For string values, compare directly without marshaling
	if strValue, ok := value.(string); ok {
		return strValue == expectedValue, nil
	}

	// For non-string values, marshal to JSON for consistent string representation
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal extracted value: %w", err)
	}

	valueStr := string(jsonBytes)
	return valueStr == expectedValue, nil
}

func (m *JSONMatcher) ExtractValue(pathExpr string) (string, error) {
	value, err := jsonpath.Get(pathExpr, m.payload)
	if err != nil {
		return "", err
	}

	// Marshal to JSON for consistent string representation
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal extracted value: %w", err)
	}

	return string(jsonBytes), nil
}

func (m *JSONMatcher) GetPayload() interface{} {
	return m.payload
}

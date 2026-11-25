package webhooks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theory/jsonpath"
)

type JSONMatcher struct {
	payload any
}

func NewJSONMatcher(payload []byte) (*JSONMatcher, error) {
	var data any
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}
	return &JSONMatcher{payload: data}, nil
}

func (m *JSONMatcher) MatchPath(pathExpr string, expectedValue string) (bool, error) {
	value, err := m.selectValue(pathExpr)
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
	value, err := m.selectValue(pathExpr)
	if err != nil {
		return "", err
	}

	// For string values, return directly without marshaling to avoid adding JSON quotes
	if strValue, ok := value.(string); ok {
		return strValue, nil
	}

	// For non-string values, marshal to JSON for consistent string representation
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal extracted value: %w", err)
	}

	return string(jsonBytes), nil
}

// selectValue returns the single value for a path that can only match one node,
// and a list for a path that can match several (wildcards, descendants). A path
// that matches nothing is an error, which callers use to tell that it is absent.
func (m *JSONMatcher) selectValue(pathExpr string) (any, error) {
	path, err := jsonpath.Parse(withRootSelector(pathExpr))
	if err != nil {
		return nil, err
	}

	nodes := path.Select(m.payload)
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no value found at JSONPath %q", pathExpr)
	}

	if path.Query().Singular() != nil {
		return nodes[0], nil
	}

	return []any(nodes), nil
}

// Earlier versions of OliveTin accepted paths without the leading "$", such as
// "repository.name", so keep accepting them.
func withRootSelector(pathExpr string) string {
	if strings.HasPrefix(pathExpr, "$") {
		return pathExpr
	}

	return "$." + pathExpr
}

func (m *JSONMatcher) GetPayload() any {
	return m.payload
}

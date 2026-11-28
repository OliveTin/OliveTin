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

// HasValue reports whether the path matches at least one value.
func (m *JSONMatcher) HasValue(pathExpr string) (bool, error) {
	_, nodes, err := m.selectNodes(pathExpr)
	if err != nil {
		return false, err
	}
	return len(nodes) > 0, nil
}

// selectValue returns the single value for a path that can only match one node,
// and a list for a path that can match several (wildcards, descendants). A
// singular path that matches nothing is an error, which callers use to tell
// that it is absent. A path that can match several nodes returns a list even
// when it is empty, so that "$.commits[*].id" on a push with no commits is "[]".
func (m *JSONMatcher) selectValue(pathExpr string) (any, error) {
	path, nodes, err := m.selectNodes(pathExpr)
	if err != nil {
		return nil, err
	}

	if path.Query().Singular() == nil {
		return append([]any{}, nodes...), nil
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no value found at JSONPath %q", pathExpr)
	}

	return nodes[0], nil
}

func (m *JSONMatcher) selectNodes(pathExpr string) (*jsonpath.Path, []any, error) {
	path, err := jsonpath.Parse(withRootSelector(pathExpr))
	if err != nil {
		return nil, nil, err
	}

	return path, path.Select(m.payload), nil
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

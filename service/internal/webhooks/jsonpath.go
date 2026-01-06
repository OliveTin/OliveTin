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

	valueStr := fmt.Sprintf("%v", value)
	return valueStr == expectedValue, nil
}

func (m *JSONMatcher) ExtractValue(pathExpr string) (string, error) {
	value, err := jsonpath.Get(pathExpr, m.payload)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", value), nil
}

func (m *JSONMatcher) GetPayload() interface{} {
	return m.payload
}

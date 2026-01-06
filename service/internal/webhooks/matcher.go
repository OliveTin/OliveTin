package webhooks

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/OliveTin/OliveTin/internal/config"
	log "github.com/sirupsen/logrus"
)

type WebhookMatcher struct {
	config    config.WebhookConfig
	req       *http.Request
	body      interface{}
	bodyBytes []byte
}

func NewWebhookMatcher(cfg config.WebhookConfig, r *http.Request, bodyBytes []byte, body interface{}) *WebhookMatcher {
	return &WebhookMatcher{
		config:    cfg,
		req:       r,
		body:      body,
		bodyBytes: bodyBytes,
	}
}

func (m *WebhookMatcher) Matches() bool {
	if !m.matchHeaders() {
		return false
	}

	if !m.matchQuery() {
		return false
	}

	if !m.matchPath() {
		return false
	}

	return true
}

func (m *WebhookMatcher) matchHeaders() bool {
	if len(m.config.MatchHeaders) == 0 {
		return true
	}

	for key, expectedValue := range m.config.MatchHeaders {
		actualValue := m.req.Header.Get(key)
		if !m.compareValues(actualValue, expectedValue) {
			log.WithFields(log.Fields{
				"header":      key,
				"expected":    expectedValue,
				"actual":      actualValue,
			}).Debugf("Header mismatch")
			return false
		}
	}
	return true
}

func (m *WebhookMatcher) matchQuery() bool {
	if len(m.config.MatchQuery) == 0 {
		return true
	}

	query := m.req.URL.Query()
	for key, expectedValue := range m.config.MatchQuery {
		actualValue := query.Get(key)
		if !m.compareValues(actualValue, expectedValue) {
			log.WithFields(log.Fields{
				"query":       key,
				"expected":    expectedValue,
				"actual":      actualValue,
			}).Debugf("Query parameter mismatch")
			return false
		}
	}
	return true
}

func (m *WebhookMatcher) matchPath() bool {
	if m.config.MatchPath == "" {
		return true
	}

	parts := strings.SplitN(m.config.MatchPath, "=", 2)
	jsonPath := parts[0]
	expectedValue := ""
	if len(parts) == 2 {
		expectedValue = parts[1]
	}

	matcher, err := NewJSONMatcher(m.bodyBytes)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Debugf("Failed to create JSON matcher")
		return false
	}

	if expectedValue == "" {
		_, err := matcher.ExtractValue(jsonPath)
		return err == nil
	}

	matches, err := matcher.MatchPath(jsonPath, expectedValue)
	if err != nil {
		log.WithFields(log.Fields{
			"jsonPath": jsonPath,
			"error":    err,
		}).Debugf("Failed to match JSONPath")
		return false
	}
	return matches
}

func (m *WebhookMatcher) compareValues(actual, expected string) bool {
	if strings.HasPrefix(expected, "regex:") {
		pattern := strings.TrimPrefix(expected, "regex:")
		matched, err := regexp.MatchString(pattern, actual)
		if err != nil {
			log.WithFields(log.Fields{
				"pattern": pattern,
				"error":   err,
			}).Warnf("Invalid regex pattern")
			return false
		}
		return matched
	}
	return actual == expected
}

func (m *WebhookMatcher) ExtractArguments() (map[string]string, error) {
	args := make(map[string]string)

	matcher, err := NewJSONMatcher(m.bodyBytes)
	if err != nil {
		return nil, err
	}

	for argName, jsonPath := range m.config.Extract {
		value, err := matcher.ExtractValue(jsonPath)
		if err != nil {
			log.WithFields(log.Fields{
				"argName": argName,
				"jsonPath": jsonPath,
				"error":    err,
			}).Debugf("Failed to extract value")
			continue
		}
		args[argName] = value
	}

	args["webhook_method"] = m.req.Method
	args["webhook_path"] = m.req.URL.Path
	args["webhook_query"] = m.req.URL.RawQuery

	for key, values := range m.req.Header {
		if len(values) > 0 {
			args["webhook_header_"+strings.ToLower(key)] = values[0]
		}
	}

	return args, nil
}

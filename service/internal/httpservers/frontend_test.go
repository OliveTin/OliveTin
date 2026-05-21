package httpservers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactHeaderValuesForLog(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"[redacted]"}, redactHeaderValuesForLog("Authorization", []string{"Bearer secret"}))
	assert.Equal(t, []string{"[redacted]", "[redacted]"}, redactHeaderValuesForLog("Cookie", []string{"a=1", "b=2"}))
	assert.Equal(t, []string{"[redacted]"}, redactHeaderValuesForLog("authorization", []string{"x"}))
	assert.Equal(t, []string{"[redacted]"}, redactHeaderValuesForLog("X-Forwarded-Access-Token", []string{"jwt"}))
	assert.Equal(t, []string{"https"}, redactHeaderValuesForLog("X-Forwarded-Proto", []string{"https"}))
}

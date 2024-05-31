package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetEmojiByShortName(t *testing.T) {
	assert.Equal(t, "&#x1F600;", lookupHTMLIcon("smile", "empty"), "Find an eomji by short name")

	assert.Equal(t, "empty", lookupHTMLIcon("", "empty"), "Find an eomji when the value is empty")

	assert.Equal(t, "notfound", lookupHTMLIcon("notfound", "empty"), "Find an eomji by undefined short name")
}

package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetEmojiByShortName(t *testing.T) {
	assert.Equal(t, "&#x1F600;", lookupHTMLIcon("smile"), "Find an eomji by short name")

	assert.Equal(t, "notfound", lookupHTMLIcon("notfound"), "Find an eomji by undefined short name")
}

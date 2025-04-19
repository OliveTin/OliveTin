package config

import (
	"fmt"
	"strings"
)

var emojis = map[string]string{
	"poop":        "&#x1f4a9;",
	"smile":       "&#x1F600;",
	"ping":        "&#x1f4e1;",
	"backup":      "&#128190;",
	"reboot":      "&#128260;",
	"restart":     "&#128260;",
	"box":         "&#128230;",
	"ashtonished": "&#128562;",
	"clock":       "&#128338;",
	"disk":        "&#128189;",
	"logs":        "&#128269;",
	"light":       "&#128161;",
	"robot":       "&#129302;",
	"ssh":         "&#128272;",
	"theme":       "&#127912;",
}

func lookupHTMLIcon(keyToLookup string, defaultIcon string) string {
	if keyToLookup == "" {
		return defaultIcon
	}

	if emoji, found := emojis[keyToLookup]; found {
		return emoji
	}

	keyToLookup = expandShortPathToFullPath(keyToLookup)

	return keyToLookup
}

func expandShortPathToFullPath(keyToLookup string) string {
	if strings.Contains(keyToLookup, "/") && !strings.HasPrefix(keyToLookup, "<") {
		return fmt.Sprintf("<img src = \"/custom-webui/%s\" />", keyToLookup)
	}

	return keyToLookup
}

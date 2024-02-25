package stringvariables

import (
	"regexp"
	"strings"
	// log "github.com/sirupsen/logrus"
)

var r *regexp.Regexp

func init() {
	r = regexp.MustCompile(`{{ *?([a-zA-Z0-9_]+)\.([a-zA-Z0-9_]+) *?}}`)
}

func ReplaceEntityVars(prefix string, source string) string {
	matches := r.FindAllStringSubmatch(source, -1)

	for _, matches := range matches {
		if len(matches) == 3 {
			property := matches[2]

			val := Get(prefix + "." + property)

			source = strings.Replace(source, matches[0], val, 1)
		}
	}

	return source
}

func RemoveKeysThatStartWith(search string) {
	for k, _ := range Contents {
		if strings.HasPrefix(k, search) {
			delete(Contents, k)
		}
	}
}

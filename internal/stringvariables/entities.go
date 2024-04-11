package stringvariables

import (
	"fmt"
	"regexp"
	"strconv"
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
	for k, _ := range contents {
		if strings.HasPrefix(k, search) {
			delete(contents, k)
		}
	}
}

func GetEntities(entityTitle string) []string {
	var ret []string

	count := GetEntityCount(entityTitle)

	for i := 0; i < count; i++ {
		prefix := GetEntityPrefix(entityTitle, i)

		ret = append(ret, prefix)
	}

	return ret
}

func GetEntityPrefix(entityTitle string, entityIndex int) string {
	return "entities." + entityTitle + "." + fmt.Sprintf("%v", entityIndex)
}

func GetEntityCount(entityTitle string) int {
	count, _ := strconv.Atoi(Get("entities." + entityTitle + ".count"))

	return count
}

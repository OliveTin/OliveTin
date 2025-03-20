package entities

import (
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
	"text/template"
	"regexp"
)

var tpl = template.New("tpl")

func init() {
}

func ParseTemplate(source string) string {
	log.Infof("contents %+v", contents)
	return ParseTemplateWith(source, contents)
}

func migrateLegacyArgumentNames(rawShellCommand string) string {
	r := regexp.MustCompile("{{ *?([a-zA-Z0-9_\\.]+?) *?}}")
	foundArgumentNames := r.FindAllStringSubmatch(rawShellCommand, -1)

	for _, match := range foundArgumentNames {
		argName := match[1]

		if strings.Contains(argName, ".") {
			rawShellCommand = strings.Replace(rawShellCommand, argName, ".Entity." + argName, -1)

			log.WithFields(log.Fields{
				"old": argName,
				"new": ".Entity." + argName,
			}).Warnf("Legacy variable name found, changing to Entity")
			continue
		}

		if !strings.HasPrefix(argName, ".Arguments.") {
			log.WithFields(log.Fields{
				"old": argName,
				"new": ".Arguments." + argName,
			}).Warnf("Legacy variable name found, changing to Argument")

			rawShellCommand = strings.Replace(rawShellCommand, argName, ".Arguments." + argName, -1)
		}
	}

	log.Infof("rawShellCommand %+v", rawShellCommand)

	return rawShellCommand
}


func ParseTemplateWithArgs(source string, ent interface{}, args map[string]string) string {
	source = migrateLegacyArgumentNames(source)

	c := &variableBase{
		OliveTin:  contents.OliveTin,
		Arguments: args,
		Entity:    ent,
	}

	var sb strings.Builder

	t, err := tpl.Parse(source)

	if err != nil {
		log.Warnf("Error parsing template: %v", err)
		return ""
	}

	err = t.Execute(&sb, &c)

	if err != nil {
		log.Warnf("Error executing template: %v", err)
		return ""
	} else {
		return sb.String()
	}

}

func ParseTemplateWith(source string, ent interface{}) string {
	return ParseTemplateWithArgs(source, ent, nil)
}

func GetEntities(entityType string) []interface{} {
	// FIXME hackzone

	keys := []string{}

	for k, _ := range contents.Entities[entityType] {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	vals := []interface{}{}

	for _, v := range keys {
		vals = append(vals, contents.Entities[entityType][v])
	}

	return vals
}

func ParseTemplateBoolWith(source string, ent interface{}) bool {
	source = strings.TrimSpace(source)

	tplBool := ParseTemplateWith(source, ent)

	return tplBool == "true"
}

func ClearEntities(entityType string) {
	delete(contents.Entities, entityType)
}

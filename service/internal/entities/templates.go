package entities

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
)

var tpl = template.New("tpl")

var legacyEntityRegex = regexp.MustCompile(`{{ ([a-zA-Z0-9_]+)\.*?([a-zA-Z0-9_\.]+) }}`)

func migrateLegacyArgumentNames(rawShellCommand string) string {
	foundArgumentNames := legacyEntityRegex.FindAllStringSubmatch(rawShellCommand, -1)

	for _, match := range foundArgumentNames {
		entityName := match[1]
		argName := match[2]

		if strings.Contains(argName, ".") {
			replacement := ".CurrentEntity"

			rawShellCommand = strings.ReplaceAll(rawShellCommand, entityName, replacement)

			log.WithFields(log.Fields{
				"old": entityName,
				"new": replacement,
			}).Warnf("Legacy entity variable name found, changing to CurrentEntity")
			continue
		}

		if !strings.HasPrefix(argName, ".Arguments.") {
			log.WithFields(log.Fields{
				"old": argName,
				"new": ".Arguments." + argName,
			}).Warnf("Legacy variable name found, changing to Argument")

			rawShellCommand = strings.ReplaceAll(rawShellCommand, argName, ".Arguments."+argName)
		}
	}

	return rawShellCommand
}

func ParseTemplateWithArgs(source string, ent *Entity, args map[string]string) string {
	source = migrateLegacyArgumentNames(source)

	ret := ""

	t, err := tpl.Parse(source)

	if err != nil {
		log.WithFields(log.Fields{
			"source": source,
			"err":    err,
		}).Error("Error parsing template")
		return fmt.Sprintf("tpl parse error: %v", err.Error())
	}

	var entdata any

	if ent != nil {
		entdata = ent.Data
	}

	templateVariables := &variableBase{
		OliveTin:      contents.OliveTin,
		Arguments:     args,
		CurrentEntity: entdata,
	}

	var sb strings.Builder
	err = t.Execute(&sb, &templateVariables)

	if err != nil {
		log.WithFields(log.Fields{
			"source":        source,
			"err":           err,
			"currentEntity": ent,
		}).Errorf("Error executing template")
		ret = fmt.Sprintf("tpl exec error: %v", err.Error())
	} else {
		ret = sb.String()
	}

	return ret
}

func ParseTemplateWith(source string, ent *Entity) string {
	return ParseTemplateWithArgs(source, ent, nil)
}

func ParseTemplateBoolWith(source string, ent *Entity) bool {
	source = strings.TrimSpace(source)

	tplBool := ParseTemplateWith(source, ent)

	return tplBool == "true"
}

func ClearEntities(entityType string) {
	delete(contents.Entities, entityType)
}

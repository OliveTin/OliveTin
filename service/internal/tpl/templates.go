package tpl

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/OliveTin/OliveTin/internal/env"
	"github.com/OliveTin/OliveTin/internal/installationinfo"
	log "github.com/sirupsen/logrus"
)

func jsonFunc(v any) (string, error) {
	if v == nil {
		return "null", nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

var tpl = template.New("tpl").
	Option("missingkey=error").
	Funcs(template.FuncMap{"Json": jsonFunc})

type olivetinInfo struct {
	Build   *installationinfo.BuildInfo
	Runtime *installationinfo.RuntimeInfo
}

var legacyArgumentRegex = regexp.MustCompile(`{{\s*([a-zA-Z0-9_]+)\s*}}`)
var legacyEntityPropertiesRegex = regexp.MustCompile(`{{\s*([a-zA-Z0-9_]+)\.([a-zA-Z0-9_\.]+)\s*}}`)

type generalTemplateContext struct {
	OliveTin olivetinInfo
	Env      map[string]string
}

type actionTemplateContext struct {
	CurrentEntity interface{}
	Arguments     map[string]string

	// These are deliberately repeated because embedding structs
	// won't work in text/template.
	OliveTin olivetinInfo
	Env      map[string]string
}

var (
	cachedOliveTinInfo olivetinInfo
	cachedEnvMap       map[string]string
)

func init() {
	cachedOliveTinInfo = olivetinInfo{
		Build:   installationinfo.Build,
		Runtime: installationinfo.Runtime,
	}

	cachedEnvMap = env.BuildEnvMap()
}

func GetNewGeneralTemplateContext() *generalTemplateContext {
	return &generalTemplateContext{
		OliveTin: cachedOliveTinInfo,
		Env:      cachedEnvMap,
	}
}

func migrateLegacyEntityProperties(rawShellCommand string) string {
	foundArgumentNames := legacyEntityPropertiesRegex.FindAllStringSubmatch(rawShellCommand, -1)

	for _, match := range foundArgumentNames {
		entityName := match[1]
		argName := match[2]
		fullMatch := match[0] // The entire matched string like "{{ server.hostname }}"

		if strings.Contains(argName, ".") {
			replacement := "{{ .CurrentEntity." + argName + " }}"

			rawShellCommand = strings.ReplaceAll(rawShellCommand, fullMatch, replacement)

			log.WithFields(log.Fields{
				"old": entityName,
				"new": ".CurrentEntity",
			}).Debugf("Legacy entity variable name found, changing to CurrentEntity")
			continue
		}

		if !strings.HasPrefix(argName, ".Arguments.") {
			replacement := "{{ .CurrentEntity." + argName + " }}"

			rawShellCommand = strings.ReplaceAll(rawShellCommand, fullMatch, replacement)

			log.WithFields(log.Fields{
				"old": argName,
				"new": ".CurrentEntity." + argName,
			}).Debugf("Legacy variable name found, changing to CurrentEntity")
		}
	}

	return rawShellCommand
}

func migrateLegacyArgumentNames(rawShellCommand string) string {
	matches := legacyArgumentRegex.FindAllStringSubmatchIndex(rawShellCommand, -1)

	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		fullMatchStart := match[0]
		fullMatchEnd := match[1]
		argNameStart := match[2]
		argNameEnd := match[3]

		argName := rawShellCommand[argNameStart:argNameEnd]

		log.WithFields(log.Fields{
			"old": argName,
			"new": ".Arguments." + argName,
		}).Debugf("Legacy variable name found, changing to Argument")

		replacement := "{{ .Arguments." + argName + " }}"
		rawShellCommand = rawShellCommand[:fullMatchStart] + replacement + rawShellCommand[fullMatchEnd:]
	}

	return rawShellCommand
}

func ParseTemplateWithActionContext(source string, ent *entities.Entity, args map[string]string) (string, error) {
	source = migrateLegacyArgumentNames(source)
	source = migrateLegacyEntityProperties(source)

	var entdata any

	if ent != nil {
		entdata = ent.Data
	}

	templateVariables := &actionTemplateContext{
		OliveTin: cachedOliveTinInfo,
		Env:      cachedEnvMap,

		Arguments:     args,
		CurrentEntity: entdata,
	}

	result, err := parseTemplate(source, templateVariables)

	if isMissingArgumentError, argName := checkMissingArgumentError(err); isMissingArgumentError {
		return "", fmt.Errorf("required arg not provided: %s", argName)
	}

	if err != nil {
		return "", err
	}

	return result, nil
}

func checkMissingArgumentError(err error) (bool, string) {
	if err == nil {
		return false, ""
	}

	if strings.Contains(err.Error(), "map has no entry for key") {
		re := regexp.MustCompile(`\.Arguments\.(\w+)`)
		match := re.FindStringSubmatch(err.Error())
		if len(match) > 1 {
			return true, match[1]
		}
	}

	return false, ""
}

func parseTemplate(source string, data any) (string, error) {
	t, err := tpl.Parse(source)

	if err != nil {
		return "", err
	}

	var sb strings.Builder
	err = t.Execute(&sb, data)

	if err != nil {
		log.WithFields(log.Fields{
			"source": source,
			"err":    err,
		}).Errorf("Error executing template")

		return "", err
	} else {
		return sb.String(), nil
	}
}

func ParseTemplateOfActionBeforeExec(source string, ent *entities.Entity) string {
	result, err := ParseTemplateWithActionContext(source, ent, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"source": source,
			"err":    err,
		}).Errorf("Error parsing template of action before exec")
		return ""
	}
	return result
}

/*
func ParseTemplateBoolWith(source string, ent *entities.Entity) bool {
	source = strings.TrimSpace(source)

	tplBool := ParseTemplateOfActionBeforeExec(source, ent)

	return tplBool == "true"
}
*/

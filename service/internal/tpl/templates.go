package tpl

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/OliveTin/OliveTin/internal/installationinfo"
	log "github.com/sirupsen/logrus"
)

var tpl = template.New("tpl")

type olivetinInfo struct {
	Build   *installationinfo.BuildInfo
	Runtime *installationinfo.RuntimeInfo
}

var legacyArgumentRegex = regexp.MustCompile(`{{ ([a-zA-Z0-9_]+) }}`)
var legacyEntityPropertiesRegex = regexp.MustCompile(`{{ ([a-zA-Z0-9_]+)\.([a-zA-Z0-9_\.]+) }}`)

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

	cachedEnvMap = buildEnvMap()
}

func GetNewGeneralTemplateContext() *generalTemplateContext {
	return &generalTemplateContext{
		OliveTin: cachedOliveTinInfo,
		Env:      cachedEnvMap,
	}
}

func buildEnvMap() map[string]string {
	envMap := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	return envMap
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
	foundArgumentNames := legacyArgumentRegex.FindAllStringSubmatch(rawShellCommand, -1)

	for _, match := range foundArgumentNames {
		argName := match[1]

		if !strings.HasPrefix(argName, ".Arguments.") {
			log.WithFields(log.Fields{
				"old": argName,
				"new": ".Arguments." + argName,
			}).Debugf("Legacy variable name found, changing to Argument")

			rawShellCommand = strings.ReplaceAll(rawShellCommand, argName, ".Arguments."+argName)
		}
	}

	return rawShellCommand
}

func ParseTemplateWithArgs(source string, ent *entities.Entity, args map[string]string) string {
	source = migrateLegacyArgumentNames(source)
	source = migrateLegacyEntityProperties(source)

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

	templateVariables := &actionTemplateContext{
		OliveTin: cachedOliveTinInfo,
		Env:      cachedEnvMap,

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

func ParseTemplateWith(source string, ent *entities.Entity) string {
	return ParseTemplateWithArgs(source, ent, nil)
}

func ParseTemplateBoolWith(source string, ent *entities.Entity) bool {
	source = strings.TrimSpace(source)

	tplBool := ParseTemplateWith(source, ent)

	return tplBool == "true"
}

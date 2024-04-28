package executor

import (
	config "github.com/OliveTin/OliveTin/internal/config"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
	log "github.com/sirupsen/logrus"

	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	typecheckRegex = map[string]string{
		"very_dangerous_raw_string": "",
		"int":                       "^[\\d]+$",
		"unicode_identifier":        "^[\\w\\/\\\\.\\_ \\d]+$",
		"ascii":                     "^[a-zA-Z0-9]+$",
		"ascii_identifier":          "^[a-zA-Z0-9\\-\\.\\_]+$",
		"ascii_sentence":            "^[a-zA-Z0-9 \\,\\.]+$",
	}
)

func parseActionArguments(rawShellCommand string, values map[string]string, action *config.Action, actionTitle string, entityPrefix string) (string, error) {
	log.WithFields(log.Fields{
		"actionTitle": actionTitle,
		"cmd":         rawShellCommand,
	}).Infof("Action parse args - Before")

	r := regexp.MustCompile("{{ *?([a-zA-Z0-9_]+?) *?}}")
	matches := r.FindAllStringSubmatch(rawShellCommand, -1)

	for _, match := range matches {
		argValue, argProvided := values[match[1]]

		if !argProvided {
			log.Infof("%v", values)
			return "", errors.New("Required arg not provided: " + match[1])
		}

		err := typecheckActionArgument(match[1], argValue, action)

		if err != nil {
			return "", err
		}

		log.WithFields(log.Fields{
			"name":  match[1],
			"value": argValue,
		}).Debugf("Arg assigned")

		rawShellCommand = strings.ReplaceAll(rawShellCommand, match[0], argValue)
	}

	rawShellCommand = sv.ReplaceEntityVars(entityPrefix, rawShellCommand)

	log.WithFields(log.Fields{
		"actionTitle": actionTitle,
		"cmd":         rawShellCommand,
	}).Infof("Action parse args - After")

	return rawShellCommand, nil
}

func typecheckActionArgument(name string, value string, action *config.Action) error {
	arg := action.FindArg(name)

	if arg == nil {
		return errors.New("Action arg not defined: " + name)
	}

	if value == "" {
		return typecheckNull(arg)
	}

	if len(arg.Choices) > 0 {
		return typecheckChoice(value, arg)
	}

	return TypeSafetyCheck(name, value, arg.Type)
}

func typecheckNull(arg *config.ActionArgument) error {
	if arg.RejectNull {
		return errors.New("Null values are not allowed")
	}

	return nil
}

func typecheckChoice(value string, arg *config.ActionArgument) error {
	if arg.Entity != "" {
		return typecheckChoiceEntity(value, arg)
	}

	for _, choice := range arg.Choices {
		if value == choice.Value {
			return nil
		}
	}

	return errors.New("argument value is not one of the predefined choices")
}

func typecheckChoiceEntity(value string, arg *config.ActionArgument) error {
	templateChoice := arg.Choices[0].Value

	for _, ent := range sv.GetEntities(arg.Entity) {
		choice := sv.ReplaceEntityVars(ent, templateChoice)

		if value == choice {
			return nil
		}
	}

	return errors.New("argument value cannot be found in entities")
}

// TypeSafetyCheck checks argument values match a specific type. The types are
// defined in typecheckRegex, and, you guessed it, uses regex to check for allowed
// characters.
func TypeSafetyCheck(name string, value string, argumentType string) error {
	if argumentType == "url" {
		return typeSafetyCheckUrl(name, value)
	}

	if argumentType == "datetime" {
		_, err := time.Parse("2006-01-02T15:04:05", value)

		if err != nil {
			return err
		}

		return nil
	}

	return typeSafetyCheckRegex(name, value, argumentType)
}

func typeSafetyCheckRegex(name string, value string, argumentType string) error {
	pattern := ""

	if strings.HasPrefix(argumentType, "regex:") {
		pattern = strings.Replace(argumentType, "regex:", "", 1)
	} else {
		found := false
		pattern, found = typecheckRegex[argumentType]

		if !found {
			return errors.New("argument type not implemented " + argumentType)
		}
	}

	matches, _ := regexp.MatchString(pattern, value)

	if !matches {
		log.WithFields(log.Fields{
			"name":    name,
			"value":   value,
			"type":    argumentType,
			"pattern": pattern,
		}).Warn("Arg type check safety failure")

		return errors.New("invalid argument, doesn't match " + argumentType)
	}

	return nil
}

func typeSafetyCheckUrl(name string, value string) error {
	_, err := url.ParseRequestURI(value)

	return err
}

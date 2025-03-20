package executor

import (
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	log "github.com/sirupsen/logrus"

	"errors"
	"net/mail"
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

func parseCommandForReplacements(rawShellCommand string, values map[string]string, entity interface{}) (string, map[string]string, error) {
	r := regexp.MustCompile("{{ *?\\.Arguments\\.([a-zA-Z0-9_]+?) *?}}")
	foundArgumentNames := r.FindAllStringSubmatch(rawShellCommand, -1)

	usedArguments := make(map[string]string)

	for _, match := range foundArgumentNames {
		argName := match[1]
		argValue, argProvided := values[argName]

		if !argProvided {
			return "", nil, errors.New("Required arg not provided: " + argName)
		}

		usedArguments[argName] = argValue
	}

	rawShellCommand = entities.ParseTemplateWithArgs(rawShellCommand, entity, values)

	return rawShellCommand, usedArguments, nil
}

func parseActionArguments(rawShellCommand string, values map[string]string, action *config.Action, actionTitle string, entity interface{}) (string, error) {
	log.WithFields(log.Fields{
		"actionTitle": actionTitle,
		"cmd":         rawShellCommand,
	}).Infof("Action parse args - Before")

	rawShellCommand, usedArgs, err := parseCommandForReplacements(rawShellCommand, values, entity)

	if err != nil {
		return "", err
	}

	for argName, argValue := range usedArgs {
		err := typecheckActionArgument(argName, argValue, action)

		if err != nil {
			return "", err
		}

		log.WithFields(log.Fields{
			"name":  argName,
			"value": argValue,
		}).Debugf("Arg assigned")
	}

	rawShellCommand = entities.ParseTemplateWith(rawShellCommand, entity)

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

	for _, ent := range entities.GetEntities(arg.Entity) {
		choice := entities.ParseTemplateWith(templateChoice, ent)

		if value == choice {
			return nil
		}
	}

	return errors.New("argument value cannot be found in entities")
}

// TypeSafetyCheck checks argument values match a specific type. The types are
// defined in typecheckRegex, and, you guessed it, uses regex to check for allowed
// characters.
//
//gocyclo:ignore
func TypeSafetyCheck(name string, value string, argumentType string) error {
	switch argumentType {
	case "password":
		return nil
	case "email":
		return typeSafetyCheckEmail(name, value)
	case "url":
		return typeSafetyCheckUrl(name, value)
	case "datetime":
		return typeSafetyCheckDatetime(name, value)
	}

	return typeSafetyCheckRegex(name, value, argumentType)
}

func typeSafetyCheckEmail(name string, value string) error {
	_, err := mail.ParseAddress(value)

	log.Errorf("Email check: %v, %v", err, value)

	if err != nil {
		return err
	}

	return nil
}

func typeSafetyCheckDatetime(name string, value string) error {
	_, err := time.Parse("2006-01-02T15:04:05", value)

	if err != nil {
		return err
	}

	return nil
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

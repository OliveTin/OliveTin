package executor

import (
	config "github.com/OliveTin/OliveTin/internal/config"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
	log "github.com/sirupsen/logrus"

	"errors"
	"fmt"
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

func parseCommandForReplacements(shellCommand string, values map[string]string) (string, error) {
	r := regexp.MustCompile("{{ *?([a-zA-Z0-9_]+?) *?}}")
	foundArgumentNames := r.FindAllStringSubmatch(shellCommand, -1)

	for _, match := range foundArgumentNames {
		argName := match[1]
		argValue, argProvided := values[argName]

		if !argProvided {
			return "", errors.New("Required arg not provided: " + argName)
		}

		shellCommand = strings.ReplaceAll(shellCommand, match[0], argValue)
	}

	return shellCommand, nil
}

func parseActionArguments(values map[string]string, action *config.Action, entityPrefix string) (string, error) {
	log.WithFields(log.Fields{
		"actionTitle": action.Title,
		"cmd":         action.Shell,
	}).Infof("Action parse args - Before")

	for _, arg := range action.Arguments {
		argName := arg.Name
		argValue := values[argName]

		err := typecheckActionArgument(argName, argValue, action)

		if err != nil {
			return "", err
		}

		log.WithFields(log.Fields{
			"name":  argName,
			"value": argValue,
		}).Debugf("Arg assigned")
	}

	parsedShellCommand, err := parseCommandForReplacements(action.Shell, values)
	parsedShellCommand = sv.ReplaceEntityVars(entityPrefix, parsedShellCommand)
	redactedShellCommand := redactShellCommand(parsedShellCommand, action.Arguments, values)

	if err != nil {
		return "", err
	}

	log.WithFields(log.Fields{
		"actionTitle": action.Title,
		"cmd":         redactedShellCommand,
	}).Infof("Action parse args - After")

	return parsedShellCommand, nil
}

func redactShellCommand(shellCommand string, arguments []config.ActionArgument, argumentValues map[string]string) string {
	for _, arg := range arguments {
		if arg.Type == "password" {
			argValue, exists := argumentValues[arg.Name]

			if !exists {
				log.Warnf("Redact shell command: Argument %s not found in values", arg.Name)
				continue
			}

			shellCommand = strings.ReplaceAll(shellCommand, argValue, "<redacted>")
		}
	}

	return shellCommand
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

// TypeSafetyCheck checks argument values match a specific type. The types are
// defined in typecheckRegex, and, you guessed it, uses regex to check for allowed
// characters.
//
//gocyclo:ignore
func TypeSafetyCheck(name string, value string, argumentType string) error {
	switch argumentType {
	case "password":
		return nil
	case "raw_string_multiline":
		return nil
	case "email":
		return typeSafetyCheckEmail(value)
	case "url":
		return typeSafetyCheckUrl(value)
	case "datetime":
		return typeSafetyCheckDatetime(value)
	}

	return typeSafetyCheckRegex(name, value, argumentType)
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

func typeSafetyCheckEmail(value string) error {
	_, err := mail.ParseAddress(value)

	log.Errorf("Email check: %v, %v", err, value)

	if err != nil {
		return err
	}

	return nil
}

func typeSafetyCheckDatetime(value string) error {
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

		return errors.New(fmt.Sprintf("invalid argument %v, doesn't match %v", name, argumentType))
	}

	return nil
}

func typeSafetyCheckUrl(value string) error {
	_, err := url.ParseRequestURI(value)

	return err
}

package executor

import (
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	log "github.com/sirupsen/logrus"

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
		"int":                       `^\d+$`,
		"unicode_identifier":        `^[\w\-\.\_\d]+$`,
		"ascii":                     `^[a-zA-Z0-9]+$`,
		"ascii_identifier":          `^[a-zA-Z0-9\-\._]+$`,
		"ascii_sentence":            `^[a-zA-Z0-9\-\._, ]+$`,
	}
)

func parseCommandForReplacements(shellCommand string, values map[string]string, entity any) (string, error) {
	r := regexp.MustCompile(`{{ *?\.Arguments\.([a-zA-Z0-9_]+?) *?}}`)
	foundArgumentNames := r.FindAllStringSubmatch(shellCommand, -1)

	for _, match := range foundArgumentNames {
		argName := match[1]
		argValue, argProvided := values[argName]

		if !argProvided {
			return "", fmt.Errorf("required arg not provided: %v", argName)
		}

		shellCommand = strings.ReplaceAll(shellCommand, match[0], argValue)
	}

	return shellCommand, nil
}

func parseActionArguments(rawShellCommand string, values map[string]string, action *config.Action, entity *entities.Entity) (string, error) {
	log.WithFields(log.Fields{
		"actionTitle": action.Title,
		"cmd":         action.Shell,
	}).Infof("Action parse args - Before")

	rawShellCommand, err := parseCommandForReplacements(rawShellCommand, values, entity)

	for _, arg := range action.Arguments {
		argName := arg.Name
		argValue := values[argName]

		err := typecheckActionArgument(&arg, argValue, action)

		if err != nil {
			return "", err
		}

		log.WithFields(log.Fields{
			"name":  argName,
			"value": argValue,
		}).Debugf("Arg assigned")
	}

	parsedShellCommand := entities.ParseTemplateWith(rawShellCommand, entity)
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

//gocyclo:ignore
func redactShellCommand(shellCommand string, arguments []config.ActionArgument, argumentValues map[string]string) string {
	for _, arg := range arguments {
		if arg.Type == "password" {
			argValue, exists := argumentValues[arg.Name]

			if !exists {
				log.Warnf("Redact shell command: Argument %s not found in values", arg.Name)
				continue
			}

			if argValue == "" {
				continue
			}

			shellCommand = strings.ReplaceAll(shellCommand, argValue, "<redacted>")
		}
	}

	return shellCommand
}

func typecheckActionArgument(arg *config.ActionArgument, value string, action *config.Action) error {
	if arg.Type == "confirmation" {
		return nil
	}

	if arg.Name == "" {
		return fmt.Errorf("argument name cannot be empty")
	}

	return typecheckActionArgumentFound(value, action, arg)
}

func typecheckActionArgumentFound(value string, action *config.Action, arg *config.ActionArgument) error {
	if value == "" {
		return typecheckNull(arg)
	}

	if len(arg.Choices) > 0 {
		return typecheckChoice(value, arg)
	}

	return TypeSafetyCheck(arg.Name, value, arg.Type)
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
		return fmt.Errorf("null values are not allowed")
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

	return fmt.Errorf("argument value is not one of the predefined choices")
}

func typecheckChoiceEntity(value string, arg *config.ActionArgument) error {
	templateChoice := arg.Choices[0].Value

	for _, ent := range entities.GetEntityInstances(arg.Entity) {
		choice := entities.ParseTemplateWith(templateChoice, ent)

		if value == choice {
			return nil
		}
	}

	return fmt.Errorf("argument value cannot be found in entities")
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
			return fmt.Errorf("argument type not implemented %v for arg: %v", argumentType, name)
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

		return fmt.Errorf("invalid argument %v, doesn't match %v", name, argumentType)
	}

	return nil
}

func typeSafetyCheckUrl(value string) error {
	_, err := url.ParseRequestURI(value)

	return err
}

func mangleInvalidArgumentValues(req *ExecutionRequest) {
	for _, arg := range req.Action.Arguments {
		if arg.Type == "datetime" {
			mangleInvalidDatetimeValues(req, &arg)
		}

		mangleCheckboxValues(req, &arg)
	}
}

func mangleCheckboxValues(req *ExecutionRequest, arg *config.ActionArgument) {
	if arg.Type != "checkbox" {
		return
	}

	log.Infof("Checking checkbox values for argument %s in action %s", arg.Name, req.Action.Title)

	for i, _ := range arg.Choices {
		choice := &arg.Choices[i]

		if req.Arguments[arg.Name] == choice.Title {
			log.WithFields(log.Fields{
				"arg":         arg.Name,
				"oldValue":    req.Arguments[arg.Name],
				"newValue":    choice.Value,
				"actionTitle": req.Action.Title,
			}).Infof("Mangled checkbox value")

			req.Arguments[arg.Name] = choice.Value
		}
	}
}

func mangleInvalidDatetimeValues(req *ExecutionRequest, arg *config.ActionArgument) {
	value, exists := req.Arguments[arg.Name]

	if !exists || value == "" {
		return
	}

	timestamp, err := time.Parse("2006-01-02T15:04", value)

	if err == nil {
		log.WithFields(log.Fields{
			"arg":         arg.Name,
			"value":       value,
			"actionTitle": req.Action.Title,
		}).Warnf("Mangled invalid datetime value without seconds to :00 seconds, this issue is commonly caused by Android browsers.")

		req.Arguments[arg.Name] = timestamp.Format("2006-01-02T15:04:05")
	}
}

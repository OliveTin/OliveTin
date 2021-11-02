package config

import (
	log "github.com/sirupsen/logrus"
)

func Sanitize(cfg *Config) {
	sanitizeLogLevel(cfg)

	//log.Infof("cfg %p", cfg)

	for idx, _ := range cfg.Actions {
		sanitizeAction(&cfg.Actions[idx])
	}
}

func sanitizeLogLevel(cfg *Config) {
	if logLevel, err := log.ParseLevel(cfg.LogLevel); err == nil {
		log.Info("Setting log level to ", logLevel)
		log.SetLevel(logLevel)
	}
}

func sanitizeAction(action *Action) {
	if action.Timeout < 3 {
		action.Timeout = 3
	}

	action.Icon = lookupHTMLIcon(action.Icon)

	for idx, _ := range action.Arguments {
		sanitizeActionArgument(&action.Arguments[idx])
	}
}

func sanitizeActionArgument(arg *ActionArgument) {
	if arg.Title == "" {
		arg.Title = arg.Name
	}

	for idx, choice := range arg.Choices {
		if choice.Title == "" {
			arg.Choices[idx].Title = choice.Value
		}
	}

	sanitizeActionArgumentNoType(arg)

	// TODO Validate the default against the type checker, but this creates a
	// import loop
}

func sanitizeActionArgumentNoType(arg *ActionArgument) {
	if len(arg.Choices) == 0 && arg.Type == "" {
		log.WithFields(log.Fields{
			"arg": arg.Name,
		}).Warn("Argument type isn't set, will default to 'ascii' but this may not be safe. You should set a type specifically.")
		arg.Type = "ascii"
	}
}

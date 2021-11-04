package config

import (
	log "github.com/sirupsen/logrus"
)

// Sanitize will look for common configuration issues, and fix them. For example,
// populating undefined fields - name -> title, etc.
func (cfg *Config) Sanitize() {
	cfg.sanitizeLogLevel()

	// log.Infof("cfg %p", cfg)

	for idx := range cfg.Actions {
		cfg.Actions[idx].sanitize()
	}
}

func (cfg *Config) sanitizeLogLevel() {
	if logLevel, err := log.ParseLevel(cfg.LogLevel); err == nil {
		log.Info("Setting log level to ", logLevel)
		log.SetLevel(logLevel)
	}
}

func (action *Action) sanitize() {
	if action.Timeout < 3 {
		action.Timeout = 3
	}

	action.Icon = lookupHTMLIcon(action.Icon)

	for idx := range action.Arguments {
		action.Arguments[idx].sanitize()
	}
}

func (arg *ActionArgument) sanitize() {
	if arg.Title == "" {
		arg.Title = arg.Name
	}

	for idx, choice := range arg.Choices {
		if choice.Title == "" {
			arg.Choices[idx].Title = choice.Value
		}
	}

	arg.sanitizeNoType()

	// TODO Validate the default against the type checker, but this creates a
	// import loop
}

func (arg *ActionArgument) sanitizeNoType() {
	if len(arg.Choices) == 0 && arg.Type == "" {
		log.WithFields(log.Fields{
			"arg": arg.Name,
		}).Warn("Argument type isn't set, will default to 'ascii' but this may not be safe. You should set a type specifically.")
		arg.Type = "ascii"
	}
}

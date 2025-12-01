package config

import (
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Sanitize will look for common configuration issues, and fix them. For example,
// populating undefined fields - name -> title, etc.
func (cfg *Config) Sanitize() {
	cfg.sanitizeLogLevel()
	cfg.sanitizeAuthRequireGuestsToLogin()
	cfg.sanitizeLogHistoryPageSize()

	// log.Infof("cfg %p", cfg)

	for idx := range cfg.Actions {
		cfg.Actions[idx].sanitize(cfg)
	}

	cfg.sanitizeDashboardsForInlineActions()
}

func (cfg *Config) sanitizeDashboardsForInlineActions() {
	for _, dashboard := range cfg.Dashboards {
		cfg.sanitizeDashboardComponentForInlineActions(dashboard)
	}
}
func (cfg *Config) sanitizeDashboardComponentForInlineActions(component *DashboardComponent) {
	if component == nil {
		return
	}

	cfg.sanitizeInlineAction(component)
	cfg.sanitizeChildDashboardComponents(component)
}

func (cfg *Config) sanitizeInlineAction(component *DashboardComponent) {
	if component.InlineAction == nil {
		return
	}

	if component.InlineAction.Title == "" {
		component.InlineAction.Title = component.Title
	}

	component.InlineAction.sanitize(cfg)

	if cfg.inlineActionExists(component.InlineAction) {
		return
	}

	cfg.Actions = append(cfg.Actions, component.InlineAction)
}

func (cfg *Config) inlineActionExists(action *Action) bool {
	if cfg.inlineActionPointerExists(action) {
		return true
	}

	if cfg.inlineActionIDExists(action) {
		return true
	}

	return false
}

func (cfg *Config) inlineActionPointerExists(action *Action) bool {
	for _, existingAction := range cfg.Actions {
		if existingAction == action {
			return true
		}
	}

	return false
}

func (cfg *Config) inlineActionIDExists(action *Action) bool {
	if action.ID == "" {
		return false
	}

	for _, existingAction := range cfg.Actions {
		if existingAction.ID == action.ID {
			return true
		}
	}

	return false
}

func (cfg *Config) sanitizeChildDashboardComponents(component *DashboardComponent) {
	for _, child := range component.Contents {
		cfg.sanitizeDashboardComponentForInlineActions(child)
	}
}

func (cfg *Config) sanitizeLogLevel() {
	if logLevel, err := log.ParseLevel(cfg.LogLevel); err == nil {
		log.Info("Setting log level to ", logLevel)
		log.SetLevel(logLevel)
	}
}

func (action *Action) sanitize(cfg *Config) {
	if action.Timeout < 3 {
		action.Timeout = 3
	}

	action.ID = getActionID(action)
	action.Icon = lookupHTMLIcon(action.Icon, cfg.DefaultIconForActions)
	action.PopupOnStart = sanitizePopupOnStart(action.PopupOnStart, cfg)

	if action.MaxConcurrent < 1 {
		action.MaxConcurrent = 1
	}

	for idx := range action.Arguments {
		action.Arguments[idx].sanitize()
	}
}

func (cfg *Config) sanitizeAuthRequireGuestsToLogin() {
	if cfg.AuthRequireGuestsToLogin {
		log.Infof("AuthRequireGuestsToLogin is enabled. All defaultPermissions will be set to false")

		cfg.DefaultPermissions.View = false
		cfg.DefaultPermissions.Exec = false
		cfg.DefaultPermissions.Logs = false
	}
}

func (cfg *Config) sanitizeLogHistoryPageSize() {
	if cfg.LogHistoryPageSize < 10 {
		log.Warnf("LogsHistoryLimit is too low, setting it to 10")
		cfg.LogHistoryPageSize = 10
	} else if cfg.LogHistoryPageSize > 100 {
		log.Warnf("LogsHistoryLimit is high, you can do this, but expect browser lag.")
	}
}

func getActionID(action *Action) string {
	if action.ID == "" {
		return uuid.NewString()
	}

	if strings.Contains(action.ID, "{{") {
		log.Fatalf("Action IDs cannot contain variables")
	}

	return action.ID
}

//gocyclo:ignore
func sanitizePopupOnStart(raw string, cfg *Config) string {
	switch raw {
	case "execution-dialog":
		return raw
	case "execution-dialog-output-html":
		return raw
	case "execution-dialog-stdout-only":
		return raw
	case "execution-button":
		return raw
	default:
		return cfg.DefaultPopupOnStart
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

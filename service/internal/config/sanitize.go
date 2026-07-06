package config

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/OliveTin/OliveTin/internal/env"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Sanitize will look for common configuration issues, and fix them. For example,
// populating undefined fields - name -> title, etc.
func (cfg *Config) Sanitize() {
	cfg.sanitizeLogLevel()
	cfg.sanitizeAuthRequireGuestsToLogin()
	cfg.sanitizeLogHistoryPageSize()
	cfg.sanitizeLocalUsers()
	cfg.sanitizeSecurityHeaders()
	cfg.sanitizeOnClickDefaults()

	// log.Infof("cfg %p", cfg)

	for idx := range cfg.Actions {
		cfg.Actions[idx].sanitize(cfg)
	}

	cfg.sanitizeDashboardsForInlineActions()

	cfg.sanitizeActionGroups()
	cfg.sanitizeActionGroupReferences()

	if err := cfg.validateReservedActionArgumentNames(); err != nil {
		log.Fatalf("%v", err)
	}

	if err := cfg.validateChecklistChoiceValues(); err != nil {
		log.Fatalf("%v", err)
	}
}

func (cfg *Config) validateReservedActionArgumentNames() error {
	for _, action := range cfg.Actions {
		if err := action.validateReservedArgumentNames(); err != nil {
			return err
		}
	}

	return nil
}

func (action *Action) validateReservedArgumentNames() error {
	if action == nil {
		return nil
	}

	for _, arg := range action.Arguments {
		if strings.HasPrefix(arg.Name, ReservedArgumentNamePrefix) {
			return fmt.Errorf("action %q argument %q uses reserved prefix %q", action.Title, arg.Name, ReservedArgumentNamePrefix)
		}
	}

	return nil
}

func (cfg *Config) validateChecklistChoiceValues() error {
	for _, action := range cfg.Actions {
		if err := action.validateChecklistChoiceValues(); err != nil {
			return err
		}
	}

	return nil
}

func (action *Action) validateChecklistChoiceValues() error {
	if action == nil {
		return nil
	}

	for _, arg := range action.Arguments {
		if err := validateChecklistChoicesForArgument(action.Title, arg); err != nil {
			return err
		}
	}

	return nil
}

func validateChecklistChoicesForArgument(actionTitle string, arg ActionArgument) error {
	if arg.Type != "checklist" {
		return nil
	}

	for _, choice := range arg.Choices {
		if strings.Contains(choice.Value, ",") {
			return fmt.Errorf(
				`action %q argument %q choice value %q must not contain commas`,
				actionTitle,
				arg.Name,
				choice.Value,
			)
		}
	}

	return nil
}

func (cfg *Config) sanitizeDashboardsForInlineActions() {
	for _, dashboard := range cfg.Dashboards {
		cfg.sanitizeDashboardComponentForInlineActions(dashboard)
	}
}
func (cfg *Config) sanitizeDashboardComponentForInlineActions(component *DashboardComponent) {
	visited := make(map[*DashboardComponent]bool)
	cfg.sanitizeDashboardComponentForInlineActionsHelper(component, visited)
}

func (cfg *Config) sanitizeDashboardComponentForInlineActionsHelper(component *DashboardComponent, visited map[*DashboardComponent]bool) {
	if component == nil {
		return
	}

	if visited[component] {
		return
	}

	visited[component] = true

	cfg.sanitizeInlineAction(component)
	cfg.sanitizeChildDashboardComponents(component, visited)
}

func (cfg *Config) sanitizeInlineAction(component *DashboardComponent) {
	if component.InlineAction == nil {
		return
	}

	sanitizeInlineActionTitles(component)

	if component.Entity != "" && component.InlineAction.Entity == "" {
		component.InlineAction.Entity = component.Entity
	}

	component.InlineAction.sanitize(cfg)

	cfg.addInlineActionIfNotExists(component.InlineAction)
}

func (cfg *Config) addInlineActionIfNotExists(action *Action) {
	if cfg.inlineActionExists(action) {
		return
	}

	cfg.Actions = append(cfg.Actions, action)
}

func sanitizeInlineActionTitles(component *DashboardComponent) {
	if component.InlineAction.Title == "" {
		component.InlineAction.Title = component.Title
	}

	if component.Title == "" {
		component.Title = component.InlineAction.Title
	}
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

func (cfg *Config) sanitizeChildDashboardComponents(component *DashboardComponent, visited map[*DashboardComponent]bool) {
	for _, child := range component.Contents {
		if child.Entity == "" {
			child.Entity = component.Entity
		}

		cfg.sanitizeDashboardComponentForInlineActionsHelper(child, visited)
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
	migrateActionOnClick(action)
	action.sanitizeJustification()
	action.OnClick = sanitizeOnClick(action.OnClick, cfg)
	action.PopupOnStart = action.OnClick

	if action.MaxConcurrent < 1 {
		action.MaxConcurrent = 1
	}

	action.Groups = dedupeStrings(action.Groups)

	for idx := range action.Arguments {
		action.Arguments[idx].sanitize()
	}
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))

	for _, value := range values {
		out = appendUniqueString(out, seen, value)
	}

	return out
}

func appendUniqueString(out []string, seen map[string]struct{}, value string) []string {
	if value == "" {
		return out
	}

	if _, found := seen[value]; found {
		return out
	}

	seen[value] = struct{}{}

	return append(out, value)
}

const defaultActionGroupQueueSize = 5

func (cfg *Config) sanitizeActionGroups() {
	for _, group := range cfg.ActionGroups {
		if group == nil {
			continue
		}

		if group.QueueSize <= 0 {
			group.QueueSize = defaultActionGroupQueueSize
		}

		group.Icon = lookupHTMLIcon(group.Icon, cfg.DefaultIconForActions)
	}
}

func (cfg *Config) sanitizeActionGroupReferences() {
	for _, action := range cfg.Actions {
		for _, groupName := range action.Groups {
			cfg.warnInvalidActionGroupReference(action, groupName)
		}
	}
}

func (cfg *Config) warnInvalidActionGroupReference(action *Action, groupName string) {
	group, found := cfg.ActionGroups[groupName]
	if !found {
		log.WithFields(log.Fields{
			"actionTitle": action.Title,
			"groupName":   groupName,
		}).Warn("Action references unknown action group")
		return
	}

	if group == nil || group.MaxConcurrent < 1 {
		log.WithFields(log.Fields{
			"actionTitle": action.Title,
			"groupName":   groupName,
		}).Warn("Action references action group that will not be enforced at runtime")
	}
}

func (cfg *Config) sanitizeAuthRequireGuestsToLogin() {
	if cfg.AuthRequireGuestsToLogin {
		log.Infof("AuthRequireGuestsToLogin is enabled. All defaultPermissions will be set to false")

		cfg.DefaultPermissions.View = false
		cfg.DefaultPermissions.Exec = false
		cfg.DefaultPermissions.Logs = false
		cfg.DefaultPermissions.Kill = false
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

func (cfg *Config) sanitizeLocalUsers() {
	for _, user := range cfg.AuthLocalUsers.Users {
		expandLocalUserEnvTemplates(user)
	}

	if err := validateUniqueLocalUserAPIKeys(cfg.AuthLocalUsers.Users); err != nil {
		log.Fatalf("%v", err)
	}
}

func expandLocalUserEnvTemplates(user *LocalUser) {
	if user == nil {
		return
	}

	if user.Password != "" {
		user.Password = expandEnvTemplate(user.Password)
	}

	if user.ApiKey != "" {
		user.ApiKey = expandEnvTemplate(user.ApiKey)
	}
}

// validateUniqueLocalUserAPIKeys returns an error when two local users share the same non-empty apiKey.
func validateUniqueLocalUserAPIKeys(users []*LocalUser) error {
	seen := make(map[string]string)

	for _, user := range users {
		if err := recordUniqueLocalUserAPIKey(seen, user); err != nil {
			return err
		}
	}

	return nil
}

func recordUniqueLocalUserAPIKey(seen map[string]string, user *LocalUser) error {
	if user == nil || user.ApiKey == "" {
		return nil
	}

	if prior, ok := seen[user.ApiKey]; ok {
		return fmt.Errorf("duplicate authLocalUsers apiKey for users %q and %q", prior, user.Username)
	}

	seen[user.ApiKey] = user.Username

	return nil
}

func (cfg *Config) sanitizeSecurityHeaders() {
	cfg.sanitizeSecurityHeadersCSP()
	cfg.sanitizeSecurityHeadersXFrameOptions()
}

func (cfg *Config) sanitizeSecurityHeadersCSP() {
	if !cfg.Security.HeaderContentSecurityPolicy || cfg.Security.ContentSecurityPolicy != "" {
		return
	}
	cfg.Security.ContentSecurityPolicy = ContentSecurityPolicyDefault
}

func (cfg *Config) sanitizeSecurityHeadersXFrameOptions() {
	if !cfg.Security.HeaderXFrameOptions || cfg.Security.XFrameOptions != "" {
		return
	}
	cfg.Security.XFrameOptions = "DENY"
}

// expandEnvTemplate expands {{ .Env.VAR }} in config strings using the process environment.
func expandEnvTemplate(source string) string {
	t, err := template.New("envTemplate").Option("missingkey=error").Parse(source)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Debug("Env template parse failed, using literal")
		return source
	}
	var b strings.Builder
	if err := t.Execute(&b, map[string]interface{}{"Env": env.BuildEnvMap()}); err != nil {
		log.WithFields(log.Fields{"error": err}).Debug("Env template execute failed, using literal")
		return source
	}
	return b.String()
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
func sanitizeOnClick(raw string, cfg *Config) string {
	switch raw {
	case "execution-dialog":
		return raw
	case "execution-dialog-output-html":
		return raw
	case "execution-dialog-stdout-only":
		return raw
	case "execution-button":
		return raw
	case "history":
		return raw
	default:
		return cfg.DefaultOnClick
	}
}

func migrateActionOnClick(action *Action) {
	if action.OnClick == "" && action.PopupOnStart != "" {
		action.OnClick = action.PopupOnStart
	}
}

func (action *Action) sanitizeJustification() {
	switch action.Justification {
	case "false":
		action.Justification = ""
	case "true":
		action.Justification = JustificationRequiredNoTemplate
	}
}

func shouldMigrateDefaultOnClickFromPopup(onClick, popupOnStart string) bool {
	if popupOnStart == "" {
		return false
	}
	if onClick == "" {
		return true
	}
	return onClick == "nothing" && popupOnStart != "nothing"
}

func (cfg *Config) migrateDefaultOnClickFromLegacyPopup() {
	if !shouldMigrateDefaultOnClickFromPopup(cfg.DefaultOnClick, cfg.DefaultPopupOnStart) {
		return
	}
	cfg.DefaultOnClick = cfg.DefaultPopupOnStart
}

func (cfg *Config) sanitizeOnClickDefaults() {
	cfg.migrateDefaultOnClickFromLegacyPopup()

	if cfg.DefaultOnClick == "" {
		cfg.DefaultOnClick = "nothing"
	}

	cfg.DefaultPopupOnStart = cfg.DefaultOnClick
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
	arg.sanitizeChecklist()

	// Default value validation runs in executor at config load (validateArgumentDefaults).
}

func (arg *ActionArgument) sanitizeChecklist() {
	if arg.Type != "checklist" {
		return
	}

	arg.warnMissingChecklistChoices()
	arg.warnInvalidChecklistEntityTemplate()
}

func (arg *ActionArgument) warnMissingChecklistChoices() {
	if len(arg.Choices) == 0 {
		log.WithFields(log.Fields{
			"arg": arg.Name,
		}).Warn("Checklist argument has no choices defined")
	}
}

func (arg *ActionArgument) warnInvalidChecklistEntityTemplate() {
	if arg.Entity == "" || len(arg.Choices) == 1 {
		return
	}

	log.WithFields(log.Fields{
		"arg":    arg.Name,
		"entity": arg.Entity,
	}).Warn("Checklist argument with entity should define exactly one choice template")
}

func (arg *ActionArgument) sanitizeNoType() {
	if len(arg.Choices) == 0 && arg.Type == "" {
		log.WithFields(log.Fields{
			"arg": arg.Name,
		}).Warn("Argument type isn't set, will default to 'ascii' but this may not be safe. You should set a type specifically.")
		arg.Type = "ascii"
	}
}

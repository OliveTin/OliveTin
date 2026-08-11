package configissues

const (
	SeverityWarning = "warning"
	SeverityError   = "error"

	CodeActionGroupUnknown     = "action_group_unknown"
	CodeActionGroupUnenforced  = "action_group_unenforced"
	CodeArgTypeUnset           = "arg_type_unset"
	CodeChecklistNoChoices     = "checklist_no_choices"
	CodeEntityArgumentChoices  = "entity_argument_choices"
	CodeEnvUnset               = "env_unset"
	CodeIncludeMissing         = "include_missing"
	CodeTemplateParse          = "template_parse"
	CodeArgDefaultInvalid      = "arg_default_invalid"
	CodeEntityFile             = "entity_file"
	CodeEntityEmpty            = "entity_empty"
	CodeEntityTypeUnconfigured = "entity_type_unconfigured"
	CodeCronInvalid            = "cron_invalid"
	CodeCronEntityBinding      = "cron_entity_binding"
	CodeWatcherPath            = "watcher_path"
	CodeAclUnknown             = "acl_unknown"
)

// Issue is a configuration warning or error surfaced on Diagnostics.
type Issue struct {
	Severity     string
	Code         string
	Message      string
	ActionID     string
	ActionTitle  string
	ArgumentName string
	Source       string
	ConfigFile   string
}

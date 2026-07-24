package configcheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/OliveTin/OliveTin/internal/tpl"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

// Rebuild clears and rebuilds the issue list from the current config, entity
// state, sticky load-time reports, and any extra issues provided by callers
// (for example literal argument default validation from the executor).
func Rebuild(cfg *config.Config, extra ...configissues.Issue) {
	if cfg == nil {
		configissues.Replace(extra)
		return
	}

	collected := make([]configissues.Issue, 0)
	collected = append(collected, configissues.CopySticky()...)
	collected = append(collected, collectActionGroupIssues(cfg)...)
	collected = append(collected, collectArgumentIssues(cfg)...)
	collected = append(collected, collectIncludeIssues(cfg)...)
	collected = append(collected, collectTemplateParseIssues(cfg)...)
	collected = append(collected, collectEntityFileIssues(cfg)...)
	collected = append(collected, collectEntityEmptyIssues(cfg)...)
	collected = append(collected, collectCronIssues(cfg)...)
	collected = append(collected, collectWatcherPathIssues(cfg)...)
	collected = append(collected, extra...)
	configissues.Replace(collected)
}

func collectActionGroupIssues(cfg *config.Config) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	for _, action := range cfg.Actions {
		if action == nil {
			continue
		}
		for _, groupName := range action.Groups {
			out = append(out, actionGroupIssue(cfg, action, groupName)...)
		}
	}
	return out
}

func actionGroupIssue(cfg *config.Config, action *config.Action, groupName string) []configissues.Issue {
	group, found := cfg.ActionGroups[groupName]
	if !found {
		return []configissues.Issue{actionIssue(action, configissues.SeverityWarning, configissues.CodeActionGroupUnknown,
			fmt.Sprintf("Action references unknown action group %q", groupName), groupName, "")}
	}

	if group == nil || group.MaxConcurrent < 1 {
		return []configissues.Issue{actionIssue(action, configissues.SeverityWarning, configissues.CodeActionGroupUnenforced,
			fmt.Sprintf("Action references action group %q that will not be enforced at runtime", groupName), groupName, "")}
	}

	return nil
}

func actionIssue(action *config.Action, severity, code, message, source, argName string) configissues.Issue {
	return configissues.Issue{
		Severity:     severity,
		Code:         code,
		Message:      message,
		ActionID:     action.ID,
		ActionTitle:  action.Title,
		ArgumentName: argName,
		Source:       source,
		ConfigFile:   action.SourceFile,
	}
}

func collectArgumentIssues(cfg *config.Config) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	for _, action := range cfg.Actions {
		if action == nil {
			continue
		}
		for i := range action.Arguments {
			out = append(out, argumentIssues(action, &action.Arguments[i])...)
		}
	}
	return out
}

func argumentIssues(action *config.Action, arg *config.ActionArgument) []configissues.Issue {
	if arg.Type != "checklist" {
		return nil
	}

	out := make([]configissues.Issue, 0)
	out = append(out, checklistNoChoicesIssue(action, arg)...)
	out = append(out, checklistEntityChoicesIssue(action, arg)...)
	return out
}

func checklistNoChoicesIssue(action *config.Action, arg *config.ActionArgument) []configissues.Issue {
	if len(arg.Choices) > 0 {
		return nil
	}

	return []configissues.Issue{actionIssue(action, configissues.SeverityWarning, configissues.CodeChecklistNoChoices,
		"Checklist argument has no choices defined", "", arg.Name)}
}

func checklistEntityChoicesIssue(action *config.Action, arg *config.ActionArgument) []configissues.Issue {
	if arg.Entity == "" || len(arg.Choices) <= 1 {
		return nil
	}

	return []configissues.Issue{actionIssue(action, configissues.SeverityWarning, configissues.CodeChecklistEntityChoices,
		"Checklist argument with entity should define exactly one choice template", arg.Entity, arg.Name)}
}

func collectIncludeIssues(cfg *config.Config) []configissues.Issue {
	if cfg.Include == "" {
		return nil
	}

	includePath := cfg.Include
	if !filepath.IsAbs(includePath) {
		includePath = filepath.Join(cfg.GetDir(), cfg.Include)
	}
	info, err := os.Stat(includePath)
	if err != nil {
		return []configissues.Issue{{
			Severity:   configissues.SeverityWarning,
			Code:       configissues.CodeIncludeMissing,
			Message:    fmt.Sprintf("Include directory not found: %s", includePath),
			Source:     includePath,
			ConfigFile: includePath,
		}}
	}

	if !info.IsDir() {
		return []configissues.Issue{{
			Severity:   configissues.SeverityWarning,
			Code:       configissues.CodeIncludeMissing,
			Message:    fmt.Sprintf("Include path is not a directory: %s", includePath),
			Source:     includePath,
			ConfigFile: includePath,
		}}
	}

	return nil
}

func collectTemplateParseIssues(cfg *config.Config) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	for _, action := range cfg.Actions {
		if action == nil {
			continue
		}
		for i := range action.Arguments {
			arg := &action.Arguments[i]
			out = append(out, templateIssuesForArg(action, arg)...)
		}
	}
	return out
}

func templateIssuesForArg(action *config.Action, arg *config.ActionArgument) []configissues.Issue {
	out := make([]configissues.Issue, 0)

	if looksLikeTemplate(arg.Default) {
		if err := tpl.CheckActionTemplateParses(arg.Default); err != nil {
			out = append(out, actionIssue(action, configissues.SeverityError, configissues.CodeTemplateParse,
				fmt.Sprintf("Argument default template failed to parse: %v", err), arg.Default, arg.Name))
		}
	}

	for _, choice := range arg.Choices {
		out = append(out, templateIssueForChoice(action, arg, choice.Title, "title")...)
		out = append(out, templateIssueForChoice(action, arg, choice.Value, "value")...)
	}

	return out
}

func templateIssueForChoice(action *config.Action, arg *config.ActionArgument, value, field string) []configissues.Issue {
	if !looksLikeTemplate(value) {
		return nil
	}

	if err := tpl.CheckActionTemplateParses(value); err != nil {
		return []configissues.Issue{actionIssue(action, configissues.SeverityError, configissues.CodeTemplateParse,
			fmt.Sprintf("Argument choice %s template failed to parse: %v", field, err), value, arg.Name)}
	}

	return nil
}

func looksLikeTemplate(value string) bool {
	return strings.Contains(value, "{{")
}

func collectEntityFileIssues(cfg *config.Config) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	baseDir := entities.ResolveEntitiesBaseDir(cfg.GetDir())

	for _, ef := range cfg.Entities {
		out = append(out, entityFileIssuesFor(ef, baseDir)...)
	}

	return out
}

func entityFileIssuesFor(ef *config.EntityFile, baseDir string) []configissues.Issue {
	if ef == nil || ef.File == "" {
		return nil
	}

	path := resolveEntityPath(ef.File, baseDir)
	issue := entityFileIssue(ef, path)
	if issue == nil {
		return nil
	}
	return []configissues.Issue{*issue}
}

func resolveEntityPath(file, baseDir string) string {
	if filepath.IsAbs(file) {
		return file
	}
	return filepath.Join(baseDir, file)
}

func entityFileIssue(ef *config.EntityFile, path string) *configissues.Issue {
	data, err := os.ReadFile(path)
	if err != nil {
		return entityReadIssue(ef, path, err)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return &configissues.Issue{
			Severity:   configissues.SeverityWarning,
			Code:       configissues.CodeEntityFile,
			Message:    fmt.Sprintf("Entity %q file is empty", ef.Name),
			Source:     path,
			ConfigFile: ef.SourceFile,
		}
	}

	if strings.HasSuffix(path, ".json") {
		return entityJSONParseIssue(ef, path, data)
	}

	return entityYAMLParseIssue(ef, path, data)
}

func entityReadIssue(ef *config.EntityFile, path string, err error) *configissues.Issue {
	return &configissues.Issue{
		Severity:   configissues.SeverityError,
		Code:       configissues.CodeEntityFile,
		Message:    fmt.Sprintf("Entity %q file could not be read: %v", ef.Name, err),
		Source:     path,
		ConfigFile: ef.SourceFile,
	}
}

func entityJSONParseIssue(ef *config.EntityFile, path string, data []byte) *configissues.Issue {
	decoder := json.NewDecoder(bytes.NewReader(data))
	for decoder.More() {
		d := make(map[string]any)
		if err := decoder.Decode(&d); err != nil {
			return &configissues.Issue{
				Severity:   configissues.SeverityError,
				Code:       configissues.CodeEntityFile,
				Message:    fmt.Sprintf("Entity %q file could not be parsed as JSON: %v", ef.Name, err),
				Source:     path,
				ConfigFile: ef.SourceFile,
			}
		}
	}
	return nil
}

func entityYAMLParseIssue(ef *config.EntityFile, path string, data []byte) *configissues.Issue {
	var parsed []map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return &configissues.Issue{
			Severity:   configissues.SeverityError,
			Code:       configissues.CodeEntityFile,
			Message:    fmt.Sprintf("Entity %q file could not be parsed as YAML: %v", ef.Name, err),
			Source:     path,
			ConfigFile: ef.SourceFile,
		}
	}
	return nil
}

func collectEntityEmptyIssues(cfg *config.Config) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	for _, action := range cfg.Actions {
		out = append(out, entityEmptyIssueFor(cfg, action)...)
	}
	return out
}

func entityEmptyIssueFor(cfg *config.Config, action *config.Action) []configissues.Issue {
	if !shouldReportEntityEmpty(cfg, action) {
		return nil
	}

	return []configissues.Issue{actionIssue(action, configissues.SeverityWarning, configissues.CodeEntityEmpty,
		fmt.Sprintf("Entity-bound action has no instances of entity %q", action.Entity), action.Entity, "")}
}

func shouldReportEntityEmpty(cfg *config.Config, action *config.Action) bool {
	if action == nil || action.Entity == "" {
		return false
	}
	if len(entities.GetEntityInstances(action.Entity)) > 0 {
		return false
	}
	return entityEmptyLoadGatePassed(cfg, action.Entity)
}

func entityEmptyLoadGatePassed(cfg *config.Config, entityName string) bool {
	// Avoid startup false positives: entity files load after the first action-map rebuild.
	if entityTypeConfigured(cfg, entityName) && !entities.HasEntityLoadAttempted(entityName) {
		return false
	}
	return true
}

func entityTypeConfigured(cfg *config.Config, entityName string) bool {
	for _, ef := range cfg.Entities {
		if ef != nil && ef.Name == entityName {
			return true
		}
	}
	return false
}

func collectCronIssues(cfg *config.Config) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	parser := cronParser(cfg)

	for _, action := range cfg.Actions {
		out = append(out, cronIssuesForAction(action, parser)...)
	}

	return out
}

func cronIssuesForAction(action *config.Action, parser cron.Parser) []configissues.Issue {
	if action == nil {
		return nil
	}

	out := make([]configissues.Issue, 0)
	out = append(out, cronEntityBindingIssue(action)...)
	out = append(out, cronScheduleIssues(action, parser)...)
	return out
}

func cronEntityBindingIssue(action *config.Action) []configissues.Issue {
	if action.Entity == "" || len(action.ExecOnCron) == 0 {
		return nil
	}

	return []configissues.Issue{actionIssue(action, configissues.SeverityWarning, configissues.CodeCronEntityBinding,
		"Entity-bound action uses execOnCron; cron runs without an entity binding and will not execute", action.Entity, "")}
}

func cronScheduleIssues(action *config.Action, parser cron.Parser) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	for _, cronline := range action.ExecOnCron {
		out = append(out, cronScheduleIssue(action, parser, cronline)...)
	}
	return out
}

func cronScheduleIssue(action *config.Action, parser cron.Parser, cronline string) []configissues.Issue {
	_, err := parser.Parse(cronline)
	if err == nil {
		return nil
	}

	return []configissues.Issue{actionIssue(action, configissues.SeverityError, configissues.CodeCronInvalid,
		fmt.Sprintf("Invalid cron schedule %q: %v", cronline, err), cronline, "")}
}

func cronParser(cfg *config.Config) cron.Parser {
	if cfg.CronSupportForSeconds {
		return cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	}
	return cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
}

func collectWatcherPathIssues(cfg *config.Config) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	for _, action := range cfg.Actions {
		out = append(out, watcherPathIssuesForAction(action)...)
	}
	out = append(out, watcherPathIssuesForEntities(cfg)...)
	return out
}

func watcherPathIssuesForAction(action *config.Action) []configissues.Issue {
	if action == nil {
		return nil
	}

	out := make([]configissues.Issue, 0)
	out = append(out, watcherCreatedDirIssues(action)...)
	out = append(out, watcherChangedDirIssues(action)...)
	out = append(out, watcherCalendarFileIssues(action)...)
	return out
}

func watcherCreatedDirIssues(action *config.Action) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	for _, dir := range action.ExecOnFileCreatedInDir {
		out = append(out, watcherDirIssue(action, dir)...)
	}
	return out
}

func watcherChangedDirIssues(action *config.Action) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	for _, dir := range action.ExecOnFileChangedInDir {
		out = append(out, watcherDirIssue(action, dir)...)
	}
	return out
}

func watcherCalendarFileIssues(action *config.Action) []configissues.Issue {
	return watcherFileIssue(action, action.ExecOnCalendarFile)
}

func watcherPathIssuesForEntities(cfg *config.Config) []configissues.Issue {
	out := make([]configissues.Issue, 0)
	baseDir := entities.ResolveEntitiesBaseDir(cfg.GetDir())

	for _, ef := range cfg.Entities {
		if ef == nil || ef.File == "" {
			continue
		}
		path := resolveEntityPath(ef.File, baseDir)
		out = append(out, missingWatchDirIssue(path, filepath.Dir(path), ef.SourceFile, "", "")...)
	}
	return out
}

func watcherDirIssue(action *config.Action, dir string) []configissues.Issue {
	if dir == "" {
		return nil
	}
	return missingWatchDirIssue(dir, dir, action.SourceFile, action.ID, action.Title)
}

func watcherFileIssue(action *config.Action, filePath string) []configissues.Issue {
	if filePath == "" {
		return nil
	}
	return missingWatchDirIssue(filePath, filepath.Dir(filePath), action.SourceFile, action.ID, action.Title)
}

func missingWatchDirIssue(displayPath, dirToStat, configFile, actionID, actionTitle string) []configissues.Issue {
	_, err := os.Stat(dirToStat)
	if err == nil {
		return nil
	}

	return []configissues.Issue{{
		Severity:    configissues.SeverityError,
		Code:        configissues.CodeWatcherPath,
		Message:     fmt.Sprintf("Could not create watcher for %q: %v", displayPath, err),
		ActionID:    actionID,
		ActionTitle: actionTitle,
		Source:      displayPath,
		ConfigFile:  configFile,
	}}
}

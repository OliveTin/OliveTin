package executor

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/configcheck"
	"github.com/OliveTin/OliveTin/internal/configissues"
	"github.com/OliveTin/OliveTin/internal/entities"
)

func (e *Executor) FindBindingByID(id string) *ActionBinding {
	e.MapActionBindingsLock.RLock()
	pair, found := e.MapActionBindings[id]
	e.MapActionBindingsLock.RUnlock()

	if !found {
		return nil
	}

	return pair
}

func (e *Executor) FindBindingWithNoEntity(action *config.Action) *ActionBinding {
	e.MapActionBindingsLock.RLock()

	defer e.MapActionBindingsLock.RUnlock()

	for _, binding := range e.MapActionBindings {
		if binding.Action == action && binding.Entity == nil {
			return binding
		}
	}

	return nil
}

type RebuildActionMapRequest struct {
	Cfg              *config.Config
	dashboardTargets *dashboardTargetIndex
}

func validateArgumentDefaults(cfg *config.Config) []configissues.Issue {
	if cfg == nil {
		return nil
	}
	out := make([]configissues.Issue, 0)
	for _, action := range cfg.Actions {
		out = append(out, validateActionArgumentDefaults(action)...)
	}
	return out
}

func validateActionArgumentDefaults(action *config.Action) []configissues.Issue {
	if action == nil {
		return nil
	}
	out := make([]configissues.Issue, 0)
	for i := range action.Arguments {
		if issue := validateArgumentDefault(action, &action.Arguments[i]); issue != nil {
			out = append(out, *issue)
		}
	}
	return out
}

func validateArgumentDefault(action *config.Action, arg *config.ActionArgument) *configissues.Issue {
	if arg.Default == "" {
		return nil
	}
	if strings.Contains(arg.Default, "{{") {
		return nil
	}
	if err := ValidateArgument(arg, arg.Default, action); err != nil {
		return &configissues.Issue{
			Severity:     configissues.SeverityWarning,
			Code:         configissues.CodeArgDefaultInvalid,
			Message:      fmt.Sprintf("Argument default value failed validation: %v", err),
			ActionID:     action.ID,
			ActionTitle:  action.Title,
			ArgumentName: arg.Name,
			Source:       arg.Default,
			ConfigFile:   action.SourceFile,
		}
	}
	return nil
}

func (e *Executor) RebuildActionMap() {
	defaultIssues := validateArgumentDefaults(e.Cfg)

	e.MapActionBindingsLock.Lock()

	clear(e.MapActionBindings)

	req := &RebuildActionMapRequest{
		Cfg:              e.Cfg,
		dashboardTargets: buildDashboardTargetIndex(e.Cfg),
	}

	for configOrder, action := range e.Cfg.Actions {
		if action.Entity != "" {
			registerActionsFromEntities(e, configOrder, action.Entity, action, req)
		} else {
			registerAction(e, configOrder, action, req)
		}
	}

	e.MapActionBindingsLock.Unlock()

	configcheck.Rebuild(e.Cfg, defaultIssues...)

	for _, l := range e.copyListeners() {
		l.OnActionMapRebuilt()
	}
}

func registerAction(e *Executor, configOrder int, action *config.Action, req *RebuildActionMapRequest) {
	bindingId := generateActionBindingId(action, "")

	e.MapActionBindings[bindingId] = &ActionBinding{
		ID:           bindingId,
		Action:       action,
		Entity:       nil,
		ConfigOrder:  configOrder,
		OnDashboards: resolveOnDashboards(req.dashboardTargets, action.Title, ""),
	}
}

func registerActionsFromEntities(e *Executor, configOrder int, entityTitle string, tpl *config.Action, req *RebuildActionMapRequest) {
	for _, ent := range entities.GetEntityInstancesOrdered(entityTitle) {
		registerActionFromEntity(e, configOrder, tpl, ent, req)
	}
}

func registerActionFromEntity(e *Executor, configOrder int, tpl *config.Action, ent *entities.Entity, req *RebuildActionMapRequest) {
	virtualActionId := generateActionBindingId(tpl, ent.UniqueKey)

	e.MapActionBindings[virtualActionId] = &ActionBinding{
		ID:           virtualActionId,
		Action:       tpl,
		Entity:       ent,
		ConfigOrder:  configOrder,
		OnDashboards: resolveOnDashboards(req.dashboardTargets, tpl.Title, ent.UniqueKey),
	}
}

func generateActionBindingId(action *config.Action, entityPrefix string) string {
	if action.ID != "" && entityPrefix == "" {
		return action.ID
	}

	h := sha256.New()

	if entityPrefix == "" {
		h.Write([]byte(action.Title))
	} else {
		// Include the entity data to make each entity instance unique
		h.Write([]byte(action.Title + "." + entityPrefix))
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

package executor

import (
	"crypto/sha256"
	"fmt"
	"slices"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	log "github.com/sirupsen/logrus"
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
	Cfg                   *config.Config
	DashboardActionTitles []string
}

func (e *Executor) RebuildActionMap() {
	e.MapActionBindingsLock.Lock()

	clear(e.MapActionBindings)

	req := &RebuildActionMapRequest{
		Cfg:                   e.Cfg,
		DashboardActionTitles: make([]string, 0),
	}

	findDashboardActionTitles(req)

	log.WithFields(log.Fields{
		"titles": req.DashboardActionTitles,
	}).Trace("dashboardActionTitles")

	for configOrder, action := range e.Cfg.Actions {
		if action.Entity != "" {
			registerActionsFromEntities(e, configOrder, action.Entity, action, req)
		} else {
			registerAction(e, configOrder, action, req)
		}
	}

	e.MapActionBindingsLock.Unlock()

	for _, l := range e.listeners {
		l.OnActionMapRebuilt()
	}
}

func findDashboardActionTitles(req *RebuildActionMapRequest) {
	for _, dashboard := range req.Cfg.Dashboards {
		recurseDashboardForActionTitles(dashboard, req)
	}
}

//gocyclo:ignore
func recurseDashboardForActionTitles(component *config.DashboardComponent, req *RebuildActionMapRequest) {
	for _, sub := range component.Contents {
		if sub.InlineAction != nil {
			title := sub.Title
			if title == "" {
				title = sub.InlineAction.Title
			}
			if title != "" {
				req.DashboardActionTitles = append(req.DashboardActionTitles, title)
			}
		} else if sub.Type == "link" || sub.Type == "" {
			req.DashboardActionTitles = append(req.DashboardActionTitles, sub.Title)
		}

		if len(sub.Contents) > 0 {
			recurseDashboardForActionTitles(sub, req)
		}
	}
}

func registerAction(e *Executor, configOrder int, action *config.Action, req *RebuildActionMapRequest) {
	bindingId := generateActionBindingId(action, "")

	e.MapActionBindings[bindingId] = &ActionBinding{
		ID:            bindingId,
		Action:        action,
		Entity:        nil,
		ConfigOrder:   configOrder,
		IsOnDashboard: slices.Contains(req.DashboardActionTitles, action.Title),
	}
}

func registerActionsFromEntities(e *Executor, configOrder int, entityTitle string, tpl *config.Action, req *RebuildActionMapRequest) {
	for _, ent := range entities.GetEntityInstances(entityTitle) {
		registerActionFromEntity(e, configOrder, tpl, ent, req)
	}
}

func registerActionFromEntity(e *Executor, configOrder int, tpl *config.Action, ent *entities.Entity, req *RebuildActionMapRequest) {
	virtualActionId := generateActionBindingId(tpl, ent.UniqueKey)

	e.MapActionBindings[virtualActionId] = &ActionBinding{
		ID:            virtualActionId,
		Action:        tpl,
		Entity:        ent,
		ConfigOrder:   configOrder,
		IsOnDashboard: slices.Contains(req.DashboardActionTitles, tpl.Title),
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

package executor

import (
	"crypto/sha256"
	"fmt"
	"slices"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	log "github.com/sirupsen/logrus"
)

func (e *Executor) FindActionByBindingID(id string) *config.Action {
	binding := e.FindBindingByID(id)

	if binding == nil {
		return nil
	}

	return binding.Action
}

func (e *Executor) FindBindingByID(id string) *ActionBinding {
	e.MapActionIdToBindingLock.RLock()
	pair, found := e.MapActionIdToBinding[id]
	e.MapActionIdToBindingLock.RUnlock()

	if !found {
		return nil
	}

	return pair
}

type RebuildActionMapRequest struct {
	Cfg                   *config.Config
	DashboardActionTitles []string
}

func (e *Executor) RebuildActionMap() {
	e.MapActionIdToBindingLock.Lock()

	clear(e.MapActionIdToBinding)

	req := &RebuildActionMapRequest{
		Cfg:                   e.Cfg,
		DashboardActionTitles: make([]string, 0),
	}

	findDashboardActionTitles(req)

	log.Infof("dashboardActionTitles: %v", req.DashboardActionTitles)

	for configOrder, action := range e.Cfg.Actions {
		if action.Entity != "" {
			registerActionsFromEntities(e, configOrder, action.Entity, action, req)
		} else {
			registerAction(e, configOrder, action, req)
		}
	}

	e.MapActionIdToBindingLock.Unlock()

	for _, l := range e.listeners {
		l.OnActionMapRebuilt()
	}
}

func findDashboardActionTitles(req *RebuildActionMapRequest) {
	for _, dashboard := range req.Cfg.Dashboards {
		recurseDashboardForActionTitles(dashboard, req)
	}
}

func recurseDashboardForActionTitles(component *config.DashboardComponent, req *RebuildActionMapRequest) {
	for _, sub := range component.Contents {
		if sub.Type == "link" || sub.Type == "" {
			req.DashboardActionTitles = append(req.DashboardActionTitles, sub.Title)
		}

		if len(sub.Contents) > 0 {
			recurseDashboardForActionTitles(sub, req)
		}
	}
}

func registerAction(e *Executor, configOrder int, action *config.Action, req *RebuildActionMapRequest) {
	actionId := hashActionToID(action, "")

	e.MapActionIdToBinding[actionId] = &ActionBinding{
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
	virtualActionId := hashActionToID(tpl, "ent")

	e.MapActionIdToBinding[virtualActionId] = &ActionBinding{
		Action:        tpl,
		Entity:        ent,
		ConfigOrder:   configOrder,
		IsOnDashboard: slices.Contains(req.DashboardActionTitles, tpl.Title),
	}
}

func hashActionToID(action *config.Action, entityPrefix string) string {
	if action.ID != "" && entityPrefix == "" {
		return action.ID
	}

	h := sha256.New()

	if entityPrefix == "" {
		h.Write([]byte(action.Title))
	} else {
		h.Write([]byte(action.ID + "." + entityPrefix))
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

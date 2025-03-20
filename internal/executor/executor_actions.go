package executor

import (
	"crypto/sha256"
	"fmt"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (e *Executor) FindActionBindingByID(id string) *config.Action {
	e.MapActionIdToBindingLock.RLock()
	pair, found := e.MapActionIdToBinding[id]
	e.MapActionIdToBindingLock.RUnlock()

	if found {
		log.Infof("findActionBinding %v, %v", id, pair.Action.ID)
		return pair.Action
	}

	return nil
}

func (e *Executor) RebuildActionMap() {
	e.MapActionIdToBindingLock.Lock()

	clear(e.MapActionIdToBinding)

	for configOrder, action := range e.Cfg.Actions {
		if action.Entity != "" {
			registerActionsFromEntities(e, configOrder, action.Entity, action)
		} else {
			registerAction(e, configOrder, action)
		}
	}

	e.MapActionIdToBindingLock.Unlock()

	for _, l := range e.listeners {
		l.OnActionMapRebuilt()
	}
}

func registerAction(e *Executor, configOrder int, action *config.Action) {
	actionId := hashActionToID(action, "")

	e.MapActionIdToBinding[actionId] = &ActionBinding{
		Action:      action,
		Entity:      nil,
		ConfigOrder: configOrder,
	}
}

func registerActionsFromEntities(e *Executor, configOrder int, entityTitle string, tpl *config.Action) {
	for _, ent := range entities.GetEntities(entityTitle) {
		registerActionFromEntity(e, configOrder, tpl, ent)
	}
}

func registerActionFromEntity(e *Executor, configOrder int, tpl *config.Action, entity interface{}) {
	virtualActionId := hashActionToID(tpl, "ent")

	e.MapActionIdToBinding[virtualActionId] = &ActionBinding{
		Action:      tpl,
		Entity:      entity,
		ConfigOrder: configOrder,
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
		h.Write([]byte(uuid.NewString()))
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

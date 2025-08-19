package entities

/**
 * The ephemeralvariablemap is used "only" for variable substitution in config
 * titles, shell arguments, etc, in the foorm of {{ key }}, like Jinja2.
 *
 * OliveTin itself really only ever "writes" to this map, mostly by loading
 * EntityFiles, and the only form of "reading" is for the variable substitution
 * in configs.
 */

import (
	"strings"
	"sync"

	"github.com/OliveTin/OliveTin/internal/installationinfo"
)

type entityInstancesByKey map[string]*Entity

type entitiesByClass map[string]entityInstancesByKey

type variableBase struct {
	OliveTin installationInfo
	Entities entitiesByClass

	CurrentEntity interface{}
	Arguments     map[string]string
}

type installationInfo struct {
	Build   *installationinfo.BuildInfo
	Runtime *installationinfo.RuntimeInfo
}

var (
	contents *variableBase
	rwmutex  = sync.RWMutex{}
)

func init() {
	rwmutex.Lock()

	contents = &variableBase{
		OliveTin: installationInfo{
			Build:   installationinfo.Build,
			Runtime: installationinfo.Runtime,
		},
		Entities: make(entitiesByClass, 0),
	}

	rwmutex.Unlock()
}

func GetAll() *variableBase {
	rwmutex.RLock()
	defer rwmutex.RUnlock()

	return contents
}

func GetEntities() entitiesByClass {
	return contents.Entities
}

func GetEntityInstances(entityName string) entityInstancesByKey {
	if entities, ok := contents.Entities[entityName]; ok {
		return entities
	}

	return nil
}

func AddEntity(entityName string, entityKey string, data any) {
	rwmutex.Lock()

	if _, ok := contents.Entities[entityName]; !ok {
		contents.Entities[entityName] = make(entityInstancesByKey, 0)
	}

	contents.Entities[entityName][entityKey] = &Entity {
		Data: data,
		UniqueKey: entityKey,
		Title:      findEntityTitle(data),
	}

	rwmutex.Unlock()
}

func findEntityTitle(data any) string {
    if mapData, ok := data.(map[string]any); ok {
		keys := make(map[string]string)

		for k := range mapData {
			lookupKey := strings.ToLower(k)
			keys[lookupKey] = k
		}

		for _, key := range []string{"title", "name", "id"} {
			if lookupKey, exists := keys[strings.ToLower(key)]; exists {
				if value, ok := mapData[lookupKey]; ok {
					if valueStr, ok := value.(string); ok {
						return valueStr
					}
				}
			}
		}
	}

	return "Untitled Entity"
}

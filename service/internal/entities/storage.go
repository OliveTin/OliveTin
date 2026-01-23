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
	"os"
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
	Env           map[string]string
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

	envMap := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	contents = &variableBase{
		OliveTin: installationInfo{
			Build:   installationinfo.Build,
			Runtime: installationinfo.Runtime,
		},
		Entities: make(entitiesByClass, 0),
		Env:      envMap,
	}

	rwmutex.Unlock()
}

func GetAll() *variableBase {
	rwmutex.RLock()
	defer rwmutex.RUnlock()

	return contents
}

func GetEntities() entitiesByClass {
	rwmutex.RLock()

	copiedEntities := make(entitiesByClass, len(contents.Entities))

	for entityName, entityInstances := range contents.Entities {
		copiedInstances := make(entityInstancesByKey, len(entityInstances))

		for key, entity := range entityInstances {
			copiedInstances[key] = entity
		}
		copiedEntities[entityName] = copiedInstances
	}

	rwmutex.RUnlock()

	return copiedEntities
}

func GetEntityInstances(entityName string) entityInstancesByKey {
	rwmutex.RLock()
	defer rwmutex.RUnlock()

	if entities, ok := contents.Entities[entityName]; ok {
		copiedInstances := make(entityInstancesByKey, len(entities))

		for key, entity := range entities {
			copiedInstances[key] = entity
		}
		return copiedInstances
	}

	return make(entityInstancesByKey, 0)
}

func AddEntity(entityName string, entityKey string, data any) {
	rwmutex.Lock()

	if _, ok := contents.Entities[entityName]; !ok {
		contents.Entities[entityName] = make(entityInstancesByKey, 0)
	}

	contents.Entities[entityName][entityKey] = &Entity{
		Data:      data,
		UniqueKey: entityKey,
		Title:     findEntityTitle(data),
	}

	rwmutex.Unlock()
}

//gocyclo:ignore
func findEntityTitle(data any) string {
	if mapData, ok := data.(map[string]any); ok {
		keys := make(map[string]string)

		for k := range mapData {
			lookupKey := strings.ToLower(k)
			keys[lookupKey] = k
		}

		for _, key := range []string{"title", "name", "id", "hostname", "host", "label"} {
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

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
	"sort"
	"strconv"
	"strings"
	"sync"
)

type entityInstancesByKey map[string]*Entity

type EntitiesByClass map[string]entityInstancesByKey

var (
	rwmutex  = sync.RWMutex{}
	entities EntitiesByClass
)

func init() {
	rwmutex.Lock()
	entities = make(EntitiesByClass, 0)
	rwmutex.Unlock()
}

func GetEntities() EntitiesByClass {
	rwmutex.RLock()

	copiedEntities := make(EntitiesByClass, len(entities))

	for entityName, entityInstances := range entities {
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

	if entities, ok := entities[entityName]; ok {
		copiedInstances := make(entityInstancesByKey, len(entities))

		for key, entity := range entities {
			copiedInstances[key] = entity
		}
		return copiedInstances
	}

	return make(entityInstancesByKey, 0)
}

func GetEntityInstancesOrdered(entityName string) []*Entity {
	instances := GetEntityInstances(entityName)
	if len(instances) == 0 {
		return nil
	}

	keys := make([]string, 0, len(instances))
	for key := range instances {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return compareEntityKeys(keys[i], keys[j]) < 0
	})

	result := make([]*Entity, 0, len(keys))
	for _, key := range keys {
		result = append(result, instances[key])
	}
	return result
}

//gocyclo:ignore
func compareEntityKeys(a, b string) int {
	ai, errA := strconv.ParseInt(a, 10, 64)
	bi, errB := strconv.ParseInt(b, 10, 64)
	if errA == nil && errB == nil {
		if ai < bi {
			return -1
		}
		if ai > bi {
			return 1
		}
		return 0
	}
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func AddEntity(entityName string, entityKey string, data any) {
	rwmutex.Lock()

	if _, ok := entities[entityName]; !ok {
		entities[entityName] = make(entityInstancesByKey, 0)
	}

	entities[entityName][entityKey] = &Entity{
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

func ClearEntitiesOfType(entityType string) {
	rwmutex.Lock()
	defer rwmutex.Unlock()

	delete(entities, entityType)
}

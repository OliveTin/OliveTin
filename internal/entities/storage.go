/**
 * The ephemeralvariablemap is used "only" for variable substitution in config
 * titles, shell arguments, etc, in the foorm of {{ key }}, like Jinja2.
 *
 * OliveTin itself really only ever "writes" to this map, mostly by loading
 * EntityFiles, and the only form of "reading" is for the variable substitution
 * in configs.
 */

package entities

import (
	"github.com/OliveTin/OliveTin/internal/installationinfo"
	"sync"
)

type variableBase struct {
	OliveTin installationInfo
	Entities map[string]map[string]interface{}

	Entity    interface{}
	Arguments map[string]string
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
		Entities: make(map[string]map[string]interface{}),
	}

	rwmutex.Unlock()
}

func AddEntity(entityName string, entityKey string, entity map[string]interface{}) {
	rwmutex.Lock()

	if _, ok := contents.Entities[entityName]; !ok {
		contents.Entities[entityName] = make(map[string]interface{})
	}

	contents.Entities[entityName][entityKey] = entity

	rwmutex.Unlock()
}

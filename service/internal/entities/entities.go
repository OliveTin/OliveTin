package entities

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/filehelper"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	EntityChangedSender chan bool
	listeners           []func()
)

type Entity struct {
	Data      any
	UniqueKey string
	Title     string
}

func AddListener(l func()) {
	listeners = append(listeners, l)
}

func SetupEntityFileWatchers(cfg *config.Config) {
	baseDir := ResolveEntitiesBaseDir(cfg.GetDir())
	for i := range cfg.Entities { // #337 - iterate by key, not by value
		ef := cfg.Entities[i]
		watchAndLoadEntity(baseDir, ef)
	}
}

// ResolveEntitiesBaseDir returns the directory used to resolve relative entity file paths.
func ResolveEntitiesBaseDir(configDir string) string {
	return resolveEntitiesBaseDir(configDir)
}

//gocyclo:ignore
func resolveEntitiesBaseDir(configDir string) string {
	absConfigDir, err := filepath.Abs(configDir)

	if err != nil {
		log.Errorf("Error getting absolute path for %s: %v", configDir, err)
		return configDir
	}

	if strings.Contains(absConfigDir, "integration-tests") {
		return configDir
	}

	devVar := filepath.Join(configDir, "var")

	if _, err := os.Stat(devVar); err == nil {
		return devVar
	}
	return absConfigDir
}

func watchAndLoadEntity(baseDir string, ef *config.EntityFile) {
	p := ef.File
	if !filepath.IsAbs(p) {
		p = filepath.Join(baseDir, p)
		log.WithFields(log.Fields{"entityFile": p}).Debugf("Adding config dir to entity file path")
	}
	go filehelper.WatchFileWrite(p, func(filename string) { loadEntityFile(p, ef.Name) }, filehelper.WatchMeta{
		ConfigFile: ef.SourceFile,
	})
	loadEntityFile(p, ef.Name)
}

func loadEntityFile(filename string, entityname string) {
	defer func() {
		MarkEntityLoadAttempted(entityname)
		notifyEntityListeners()
	}()

	if strings.HasSuffix(filename, ".json") {
		loadEntityFileJson(filename, entityname)
		return
	}
	loadEntityFileYaml(filename, entityname)
}

func notifyEntityListeners() {
	for _, l := range listeners {
		l()
	}
}

func loadEntityFileJson(filename string, entityname string) {
	log.WithFields(log.Fields{
		"file": filename,
		"name": entityname,
	}).Infof("Loading entity file with JSON format")

	jfile, err := os.ReadFile(filename)

	if err != nil {
		log.Errorf("ReadIn: %v", err)
		return
	}

	data := make([]map[string]any, 0)

	decoder := json.NewDecoder(bytes.NewReader(jfile))

	for decoder.More() {
		d := make(map[string]any)

		err := decoder.Decode(&d)

		if err != nil {
			log.Errorf("%v", err)
			return
		}

		data = append(data, d)
	}

	replaceEntitiesFromFile(entityname, data)
}

func loadEntityFileYaml(filename string, entityname string) {
	log.WithFields(log.Fields{
		"file": filename,
		"name": entityname,
	}).Infof("Loading entity file with YAML format")

	yfile, err := os.ReadFile(filename)

	if err != nil {
		log.Errorf("ReadIn: %v", err)
		return
	}

	var data []map[string]any

	err = yaml.Unmarshal(yfile, &data)

	if err != nil {
		log.Errorf("Unmarshal: %v", err)
		return
	}

	replaceEntitiesFromFile(entityname, data)
}

func replaceEntitiesFromFile(entityname string, data []map[string]any) {
	rwmutex.Lock()
	defer rwmutex.Unlock()

	delete(entities, entityname)

	if len(data) == 0 {
		return
	}

	entities[entityname] = make(entityInstancesByKey, 0)
	for i, mapp := range data {
		entityKey := fmt.Sprintf("%d", i)
		entities[entityname][entityKey] = &Entity{
			Data:      mapp,
			UniqueKey: entityKey,
			Title:     findEntityTitle(mapp),
		}
	}
}

/*
//gocyclo:ignore
func serializeValueToSv(prefix string, value any) {
	if m, ok := value.(map[string]any); ok { // if value is a map we need to flatten it
		serializeMapToSv(prefix, m)
	} else if s, ok := value.([]any); ok { // if value is a slice we need to flatten it
		serializeSliceToSv(prefix, s)
	} else if f, ok := value.(float64); ok {
		if canConvertToInt64(f) {
			s := int64(f)
			sv.Set(prefix, fmt.Sprintf("%d", s))
		} else {
			sv.Set(prefix, fmt.Sprintf("%f", f))
		}
	} else {
		sv.Set(prefix, fmt.Sprintf("%v", value))
	}
}

func canConvertToInt64(f float64) bool {
	return f >= math.MinInt64 && f <= math.MaxInt64 && f == math.Trunc(f)
}

func serializeMapToSv(prefix string, m map[string]any) {
	for k, v := range m {
		serializeValueToSv(prefix+"."+k, v)
	}
}

func serializeSliceToSv(prefix string, s []any) {
	sv.Set(prefix+".count", fmt.Sprintf("%v", len(s)))

	for i, v := range s {
		serializeValueToSv(prefix+"."+fmt.Sprintf("%v", i), v)
	}
}
*/

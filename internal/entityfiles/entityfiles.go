package entityfiles

import (
	"bytes"
	"encoding/json"
	"fmt"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/filehelper"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

var (
	EntityChangedSender chan bool
	listeners           []func()
)

func AddListener(l func()) {
	listeners = append(listeners, l)
}

func SetupEntityFileWatchers(cfg *config.Config) {
	configDir := cfg.GetDir()

	configDirVar := filepath.Join(configDir, "var") // for development purposes

	if _, err := os.Stat(configDirVar); err == nil {
		configDir = configDirVar
	}

	for _, ef := range cfg.Entities {
		p := ef.File

		if !filepath.IsAbs(p) {
			p = filepath.Join(configDir, p)

			log.WithFields(log.Fields{
				"entityFile": p,
			}).Debugf("Adding config dir to entity file path")
		}

		go filehelper.WatchFileWrite(p, func(filename string) {
			loadEntityFile(p, ef.Name)
		})

		loadEntityFile(p, ef.Name)
	}
}

func loadEntityFile(filename string, entityname string) {
	if strings.HasSuffix(filename, ".json") {
		loadEntityFileJson(filename, entityname)
	} else {
		loadEntityFileYaml(filename, entityname)
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

	data := make([]map[string]string, 0)

	decoder := json.NewDecoder(bytes.NewReader(jfile))

	for decoder.More() {
		d := make(map[string]string)

		err := decoder.Decode(&d)

		if err != nil {
			log.Errorf("%v", err)
			return
		}

		data = append(data, d)
	}

	updateEvmFromFile(entityname, data)
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

	data := make([]map[string]string, 1)

	err = yaml.Unmarshal(yfile, &data)

	if err != nil {
		log.Errorf("Unmarshal: %v", err)
	}

	updateEvmFromFile(entityname, data)
}

func updateEvmFromFile(entityname string, data []map[string]string) {
	count := len(data)

	sv.RemoveKeysThatStartWith("entities." + entityname)

	sv.SetEntityCount(entityname, count)

	for i, mapp := range data {
		prefix := "entities." + entityname + "." + fmt.Sprintf("%v", i)

		for k, v := range mapp {
			sv.Set(prefix+"."+k, v)
		}
	}

	for _, l := range listeners {
		l()
	}
}

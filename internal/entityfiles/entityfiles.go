package entityfiles

import (
	"bytes"
	"encoding/json"
	"fmt"
	config "github.com/OliveTin/OliveTin/internal/config"
	sv "github.com/OliveTin/OliveTin/internal/stringvariables"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"strings"
)

func SetupEntityFileWatchers(cfg *config.Config) {
	for _, ef := range cfg.Entities {
		go watch(ef.File, ef.Name)
		loadEntityFile(ef.File, ef.Name)
	}
}

func watch(file string, entityname string) {
	log.WithFields(log.Fields{
		"file": file,
		"name": entityname,
	}).Infof("Watching entity file")

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Errorf("Could not watch entity file: %v", err)
		return
	}

	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			processEvent(watcher, file, entityname)
		}
	}()

	err = watcher.Add(file)

	if err != nil {
		log.WithFields(log.Fields{
			"file": file,
		}).Errorf("Could not create entity watcher: %v", err)
	}

	<-done
}

func processEvent(watcher *fsnotify.Watcher, filename string, entityname string) {
	select {
	case event, ok := <-watcher.Events:
		if !ok {
			return
		}

		loadEntityFileIfWritten(&event, filename, entityname)

		return
	case err := <-watcher.Errors:
		log.Errorf("Error in fsnotify: %v", err)
		return
	}
}

func loadEntityFileIfWritten(event *fsnotify.Event, filename string, entityname string) {
	if event.Has(fsnotify.Remove) {
		log.WithFields(log.Fields{
			"file": filename,
		}).Warnf("Entity file deleted! Will no longer be able to watch for changes!")
	}

	if event.Has(fsnotify.Write) {
		loadEntityFile(filename, entityname)
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

	jfile, err := ioutil.ReadFile(filename)

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

	yfile, err := ioutil.ReadFile(filename)

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

	sv.Contents["entities."+entityname+".count"] = fmt.Sprintf("%v", count)

	for i, mapp := range data {
		prefix := "entities." + entityname + "." + fmt.Sprintf("%v", i)

		for k, v := range mapp {
			sv.Contents[prefix+"."+k] = v
		}
	}
}

package config

import (
	"github.com/knadh/koanf/v2"
	log "github.com/sirupsen/logrus"
)

const sourceFileKey = "x-olivetin-source-file"

// stampSourceOnMaps sets the OliveTin source-file marker on each map in a
// koanf slice value (actions or entities).
func stampSourceOnMaps(raw any, sourceFile string) any {
	if sourceFile == "" {
		return raw
	}

	items, ok := raw.([]interface{})
	if !ok {
		return raw
	}

	for _, item := range items {
		stampSourceOnMap(item, sourceFile)
	}
	return items
}

func stampSourceOnMap(item any, sourceFile string) {
	m, ok := item.(map[string]interface{})
	if !ok {
		return
	}
	if existing, ok := m[sourceFileKey].(string); ok && existing != "" {
		return
	}
	m[sourceFileKey] = sourceFile
}

func stampLoadedConfigSources(k *koanf.Koanf, configPath string) {
	stampConfigKey(k, "actions", configPath)
	stampConfigKey(k, "entities", configPath)
}

func stampConfigKey(k *koanf.Koanf, key, configPath string) {
	if err := k.Set(key, stampSourceOnMaps(k.Get(key), configPath)); err != nil {
		log.WithFields(log.Fields{
			"key":        key,
			"configPath": configPath,
		}).Errorf("Failed to persist source stamps: %v", err)
	}
}

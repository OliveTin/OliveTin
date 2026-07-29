package config

import (
	"github.com/knadh/koanf/v2"
	log "github.com/sirupsen/logrus"
)

const sourceFileKey = "x-olivetin-source-file"

// stampSourceOnMaps sets the OliveTin source-file marker on each map in a
// koanf slice value (actions or entities). Always overwrites so user-provided
// x-olivetin-source-file values cannot spoof the real config path.
func stampSourceOnMaps(raw any, sourceFile string) any {
	if sourceFile == "" {
		return raw
	}

	items, ok := raw.([]any)
	if !ok {
		return raw
	}

	for _, item := range items {
		stampSourceOnMap(item, sourceFile)
	}
	return items
}

func stampSourceOnMap(item any, sourceFile string) {
	m, ok := item.(map[string]any)
	if !ok {
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

// applyStampedSourceFiles copies stamped source paths from koanf maps onto
// unmarshaled actions/entities. SourceFile uses koanf:"-" so YAML cannot set it.
func applyStampedSourceFiles(k *koanf.Koanf, cfg *Config) {
	actionPaths := stampedSourcePaths(k.Get("actions"))
	for i, action := range cfg.Actions {
		if action != nil {
			action.SourceFile = sourcePathAt(actionPaths, i)
		}
	}

	entityPaths := stampedSourcePaths(k.Get("entities"))
	for i, entity := range cfg.Entities {
		if entity != nil {
			entity.SourceFile = sourcePathAt(entityPaths, i)
		}
	}
}

func sourcePathAt(paths []string, index int) string {
	if index >= len(paths) {
		return ""
	}
	return paths[index]
}

func stampedSourcePaths(raw any) []string {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}

	paths := make([]string, len(items))
	for i, item := range items {
		paths[i] = stampedSourceFromMap(item)
	}
	return paths
}

func stampedSourceFromMap(item any) string {
	m, ok := item.(map[string]any)
	if !ok {
		return ""
	}
	path, _ := m[sourceFileKey].(string)
	return path
}

package config

import "github.com/knadh/koanf/v2"

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
	_ = k.Set("actions", stampSourceOnMaps(k.Get("actions"), configPath))
	_ = k.Set("entities", stampSourceOnMaps(k.Get("entities"), configPath))
}

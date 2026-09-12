package config

import "sync"

var (
	loadedSourcesMu sync.Mutex
	loadedSources   []string
)

// BeginLoadedSources clears the list of config files recorded for the current load.
func BeginLoadedSources() {
	loadedSourcesMu.Lock()
	defer loadedSourcesMu.Unlock()
	loadedSources = nil
}

// RecordLoadedSource records a config file that was successfully read during parsing.
func RecordLoadedSource(path string) {
	if path == "" {
		return
	}

	loadedSourcesMu.Lock()
	defer loadedSourcesMu.Unlock()

	for _, existing := range loadedSources {
		if existing == path {
			return
		}
	}

	loadedSources = append(loadedSources, path)
}

// LoadedSources returns a copy of config files read during the latest config load.
func LoadedSources() []string {
	loadedSourcesMu.Lock()
	defer loadedSourcesMu.Unlock()

	out := make([]string, len(loadedSources))
	copy(out, loadedSources)
	return out
}

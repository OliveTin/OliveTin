package executor

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// LoadLogsFromDisk loads persisted logs from YAML files on disk and restores them to the executor.
// This should be called during startup if saveLogs is configured.
func (e *Executor) LoadLogsFromDisk() {
	resultsDir := e.Cfg.SaveLogs.ResultsDirectory
	if resultsDir == "" {
		return
	}

	entries, skippedCount := e.readLogDirectory(resultsDir)
	if entries == nil {
		return
	}

	loadedLogs, skippedCount := e.parseLogFiles(resultsDir, entries, skippedCount)

	sort.Slice(loadedLogs, func(i, j int) bool {
		return loadedLogs[i].DatetimeStarted.Before(loadedLogs[j].DatetimeStarted)
	})

	skippedCount = e.restoreLogsToExecutor(loadedLogs, skippedCount)

	log.WithFields(log.Fields{
		"loaded":  len(loadedLogs),
		"skipped": skippedCount,
	}).Info("Finished loading persisted logs from disk")
}

// readLogDirectory reads the log directory and returns entries, or nil if the directory doesn't exist or can't be read.
func (e *Executor) readLogDirectory(resultsDir string) ([]os.DirEntry, int) {
	if _, err := os.Stat(resultsDir); os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"directory": resultsDir,
		}).Debug("Logs directory does not exist, skipping log loading")
		return nil, 0
	}

	log.WithFields(log.Fields{
		"directory": resultsDir,
	}).Info("Loading persisted logs from disk")

	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		log.WithFields(log.Fields{
			"directory": resultsDir,
			"error":     err,
		}).Warnf("Failed to read logs directory")
		return nil, 0
	}

	return entries, 0
}

// parseLogFiles parses YAML log files from the directory entries.
func (e *Executor) parseLogFiles(resultsDir string, entries []os.DirEntry, skippedCount int) ([]*InternalLogEntry, int) {
	loadedLogs := make([]*InternalLogEntry, 0)

	for _, entry := range entries {
		if !e.shouldProcessLogEntry(entry) {
			continue
		}

		logEntry, newSkippedCount := e.processLogFileEntry(resultsDir, entry.Name())
		skippedCount += newSkippedCount
		if logEntry != nil {
			loadedLogs = append(loadedLogs, logEntry)
		}
	}

	return loadedLogs, skippedCount
}

// shouldProcessLogEntry checks if a directory entry should be processed as a log file.
func (e *Executor) shouldProcessLogEntry(entry os.DirEntry) bool {
	return !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml")
}

// processLogFileEntry processes a single log file entry and returns the log entry or nil if it should be skipped.
func (e *Executor) processLogFileEntry(resultsDir, filename string) (*InternalLogEntry, int) {
	logEntry, ok := e.loadLogFileFromPath(resultsDir, filename)
	if !ok {
		return nil, 1
	}

	if logEntry.ExecutionTrackingID == "" {
		log.WithFields(log.Fields{
			"file": filepath.Join(resultsDir, filename),
		}).Warnf("Log file missing execution tracking ID, skipping")
		return nil, 1
	}

	e.restoreBindingForLogEntry(logEntry, filepath.Join(resultsDir, filename))
	return logEntry, 0
}

// loadLogFileFromPath loads and unmarshals a single log file.
func (e *Executor) loadLogFileFromPath(resultsDir, filename string) (*InternalLogEntry, bool) {
	filepath := filepath.Join(resultsDir, filename)
	data, err := os.ReadFile(filepath)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  filepath,
			"error": err,
		}).Warnf("Failed to read log file")
		return nil, false
	}

	var logEntry InternalLogEntry
	if err := yaml.Unmarshal(data, &logEntry); err != nil {
		log.WithFields(log.Fields{
			"file":  filepath,
			"error": err,
		}).Warnf("Failed to unmarshal log file")
		return nil, false
	}

	return &logEntry, true
}

// restoreBindingForLogEntry attempts to restore the binding for a log entry if it's missing or invalid.
func (e *Executor) restoreBindingForLogEntry(logEntry *InternalLogEntry, filepath string) {
	if e.hasValidBinding(logEntry) || logEntry.ActionConfigTitle == "" {
		return
	}

	binding := e.findBindingByActionTitle(logEntry.ActionConfigTitle, logEntry.EntityPrefix)
	if binding != nil {
		logEntry.Binding = binding
		return
	}

	e.logBindingNotFound(logEntry, filepath)
	logEntry.Binding = nil
}

// hasValidBinding checks if a log entry has a valid binding.
func (e *Executor) hasValidBinding(logEntry *InternalLogEntry) bool {
	return logEntry.Binding != nil && logEntry.Binding.Action != nil
}

// logBindingNotFound logs a debug message when a binding cannot be found for a log entry.
func (e *Executor) logBindingNotFound(logEntry *InternalLogEntry, filepath string) {
	log.WithFields(log.Fields{
		"file":         filepath,
		"actionTitle":  logEntry.ActionConfigTitle,
		"entityPrefix": logEntry.EntityPrefix,
		"trackingId":   logEntry.ExecutionTrackingID,
	}).Debug("Could not find binding for log entry, loading without binding")
}

// restoreLogsToExecutor restores loaded logs to the executor's internal structures.
func (e *Executor) restoreLogsToExecutor(loadedLogs []*InternalLogEntry, skippedCount int) int {
	e.logmutex.Lock()
	defer e.logmutex.Unlock()

	for _, logEntry := range loadedLogs {
		if _, exists := e.logs[logEntry.ExecutionTrackingID]; exists {
			log.WithFields(log.Fields{
				"trackingId": logEntry.ExecutionTrackingID,
			}).Debug("Log entry already exists, skipping")
			skippedCount++
			continue
		}

		logEntry.Index = int64(len(e.logsTrackingIdsByDate))
		e.logs[logEntry.ExecutionTrackingID] = logEntry
		e.logsTrackingIdsByDate = append(e.logsTrackingIdsByDate, logEntry.ExecutionTrackingID)

		if logEntry.Binding != nil {
			e.addLogToBindingMap(logEntry)
		}
	}

	return skippedCount
}

// addLogToBindingMap adds a log entry to the LogsByBindingId map.
func (e *Executor) addLogToBindingMap(logEntry *InternalLogEntry) {
	if _, containsKey := e.LogsByBindingId[logEntry.Binding.ID]; !containsKey {
		e.LogsByBindingId[logEntry.Binding.ID] = make([]*InternalLogEntry, 0)
	}
	e.LogsByBindingId[logEntry.Binding.ID] = append(e.LogsByBindingId[logEntry.Binding.ID], logEntry)
}

// findBindingByActionTitle attempts to find a binding by matching the action config title and entity prefix.
func (e *Executor) findBindingByActionTitle(actionConfigTitle string, entityPrefix string) *ActionBinding {
	e.MapActionBindingsLock.RLock()
	defer e.MapActionBindingsLock.RUnlock()

	for _, binding := range e.MapActionBindings {
		if binding.Action.Title == actionConfigTitle && e.matchesEntityPrefix(binding, entityPrefix) {
			return binding
		}
	}

	return nil
}

// matchesEntityPrefix checks if a binding matches the given entity prefix.
func (e *Executor) matchesEntityPrefix(binding *ActionBinding, entityPrefix string) bool {
	if entityPrefix == "" {
		return binding.Entity == nil
	}
	return binding.Entity != nil && binding.Entity.UniqueKey == entityPrefix
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultLogFile  = "flakey-test-runs.log"
	defaultJSONLFile = "flakey-test-runs.jsonl"
)

type testEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Output  string    `json:"Output"`
	Elapsed float64   `json:"Elapsed"`
}

type testFailure struct {
	Package string `json:"package"`
	Test    string `json:"test"`
	Output  string `json:"output"`
}

type runSummary struct {
	Passes   int `json:"passes"`
	Failures int `json:"failures"`
	Skipped  int `json:"skipped"`
}

type jsonlRecord struct {
	Run            int           `json:"run"`
	Timestamp      string        `json:"timestamp"`
	ExitCode       int           `json:"exitCode"`
	DurationMs     int64         `json:"durationMs"`
	Passes         int           `json:"passes"`
	Failures       int           `json:"failures"`
	Skipped        int           `json:"skipped"`
	FailureDetails []testFailure `json:"failureDetails"`
}

func initLog() {
	logFormat := os.Getenv("OLIVETIN_LOG_FORMAT")

	if logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
			ForceQuote:    true,
		})
	}

	log.SetLevel(log.InfoLevel)
}

func serviceRoot() string {
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "..", "..")
		if stat, statErr := os.Stat(filepath.Join(candidate, "go.mod")); statErr == nil && !stat.IsDir() {
			return candidate
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	if stat, statErr := os.Stat(filepath.Join(wd, "go.mod")); statErr == nil && !stat.IsDir() {
		return wd
	}

	return filepath.Join(wd, "..")
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func formatFailure(failure testFailure) string {
	lines := []string{
		fmt.Sprintf("FAILURE: %s", failure.Test),
		fmt.Sprintf("  package: %s", failure.Package),
	}

	output := strings.TrimSpace(failure.Output)
	if output == "" {
		output = "unknown"
	}
	lines = append(lines, fmt.Sprintf("  output: %s", output))

	return strings.Join(lines, "\n")
}

func formatRunCounts(summary runSummary) string {
	return fmt.Sprintf("%d pass %d fail %d skip", summary.Passes, summary.Failures, summary.Skipped)
}

func appendRunLog(logFile, jsonlFile string, run, exitCode int, summary runSummary, failures []testFailure, durationMs int64) error {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	passed := exitCode == 0
	result := "PASS"
	if !passed {
		result = "FAIL"
	}

	block := []string{
		fmt.Sprintf(
			"=== RUN %d | %s | %s | %d pass %d fail %d skip | %.1fs ===",
			run,
			timestamp,
			result,
			summary.Passes,
			summary.Failures,
			summary.Skipped,
			float64(durationMs)/1000,
		),
	}

	for _, failure := range failures {
		block = append(block, formatFailure(failure))
	}

	block = append(block, "")
	if err := appendFile(logFile, strings.Join(block, "\n")+"\n"); err != nil {
		return err
	}

	record := jsonlRecord{
		Run:            run,
		Timestamp:      timestamp,
		ExitCode:       exitCode,
		DurationMs:     durationMs,
		Passes:         summary.Passes,
		Failures:       summary.Failures,
		Skipped:        summary.Skipped,
		FailureDetails: failures,
	}

	encoded, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return appendFile(jsonlFile, string(encoded)+"\n")
}

func appendFile(path, content string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

func runTestsOnce(rootDir string) (int, runSummary, []testFailure, error) {
	cmd := exec.Command("go", "test", "./...", "-count=1", "-json")
	cmd.Dir = rootDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 1, runSummary{}, nil, err
	}

	if err := cmd.Start(); err != nil {
		return 1, runSummary{}, nil, err
	}

	summary := runSummary{}
	failures := make([]testFailure, 0)
	failureOutput := make(map[string]*strings.Builder)

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var event testEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}

		switch event.Action {
		case "pass":
			if event.Test != "" {
				summary.Passes++
			}
		case "fail":
			if event.Test != "" {
				summary.Failures++
				key := event.Package + "\x00" + event.Test
				failures = append(failures, testFailure{
					Package: event.Package,
					Test:    event.Test,
					Output:  "",
				})
				failureOutput[key] = &strings.Builder{}
			}
		case "skip":
			if event.Test != "" {
				summary.Skipped++
			}
		case "output":
			if event.Test == "" {
				continue
			}
			key := event.Package + "\x00" + event.Test
			if builder, ok := failureOutput[key]; ok {
				builder.WriteString(event.Output)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 1, summary, failures, err
	}

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			for index := range failures {
				key := failures[index].Package + "\x00" + failures[index].Test
				if builder, ok := failureOutput[key]; ok {
					failures[index].Output = strings.TrimSpace(builder.String())
				}
			}
			return exitErr.ExitCode(), summary, failures, nil
		}
		return 1, summary, failures, err
	}

	return 0, summary, failures, nil
}

func main() {
	initLog()

	rootDir := serviceRoot()
	logFile := envOrDefault("FLAKEY_LOG_FILE", filepath.Join(rootDir, defaultLogFile))
	jsonlFile := envOrDefault("FLAKEY_JSONL_FILE", filepath.Join(rootDir, defaultJSONLFile))

	header := strings.Join([]string{
		fmt.Sprintf("# Flaky test run log started %s", time.Now().UTC().Format(time.RFC3339)),
		fmt.Sprintf("# Log file: %s", logFile),
		fmt.Sprintf("# JSONL file: %s", jsonlFile),
		"",
	}, "\n") + "\n"

	if err := os.WriteFile(logFile, []byte(header), 0o644); err != nil {
		log.WithError(err).Fatal("failed to initialize log file")
	}
	if err := os.WriteFile(jsonlFile, nil, 0o644); err != nil {
		log.WithError(err).Fatal("failed to initialize jsonl file")
	}

	log.Infof("Logging flaky test runs to %s", logFile)
	log.Infof("Structured run data: %s", jsonlFile)

	run := 0
	for {
		run++

		start := time.Now()
		exitCode, summary, failures, err := runTestsOnce(rootDir)
		durationMs := time.Since(start).Milliseconds()

		if err != nil {
			log.WithError(err).Errorf("Run %d failed to execute: %s", run, formatRunCounts(summary))
			_ = appendRunLog(logFile, jsonlFile, run, 1, summary, failures, durationMs)
			os.Exit(1)
		}

		if logErr := appendRunLog(logFile, jsonlFile, run, exitCode, summary, failures, durationMs); logErr != nil {
			log.WithError(logErr).Fatal("failed to append run log")
		}

		result := "PASS"
		if exitCode != 0 {
			result = "FAIL"
		}
		log.Infof("Run %d: %s | %s (%.1fs) — logged", run, result, formatRunCounts(summary), float64(durationMs)/1000)

		if exitCode != 0 {
			log.Errorf("Failure on run %d, stopping. See %s", run, logFile)
			os.Exit(exitCode)
		}
	}
}

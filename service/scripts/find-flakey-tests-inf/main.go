package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultLogFile   = "flakey-test-runs.log"
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

type testRunState struct {
	summary       runSummary
	failures      []testFailure
	failureOutput map[string]*strings.Builder
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

func hasGoMod(dir string) bool {
	stat, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil && !stat.IsDir()
}

func rootFromExecutable() (string, bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", false
	}

	candidate := filepath.Join(filepath.Dir(exe), "..", "..")
	if hasGoMod(candidate) {
		return candidate, true
	}

	return "", false
}

func rootFromWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	if hasGoMod(wd) {
		return wd
	}

	return filepath.Join(wd, "..")
}

func serviceRoot() string {
	if root, ok := rootFromExecutable(); ok {
		return root
	}

	return rootFromWorkingDir()
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

func runResultLabel(exitCode int) string {
	if exitCode == 0 {
		return "PASS"
	}
	return "FAIL"
}

func buildRunLogBlock(run int, timestamp string, exitCode int, summary runSummary, failures []testFailure, durationMs int64) string {
	block := []string{
		fmt.Sprintf(
			"=== RUN %d | %s | %s | %d pass %d fail %d skip | %.1fs ===",
			run,
			timestamp,
			runResultLabel(exitCode),
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
	return strings.Join(block, "\n") + "\n"
}

func appendJSONLRecord(jsonlFile string, run int, timestamp string, exitCode int, durationMs int64, summary runSummary, failures []testFailure) error {
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

func appendRunLog(logFile, jsonlFile string, run, exitCode int, summary runSummary, failures []testFailure, durationMs int64) error {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	if err := appendFile(logFile, buildRunLogBlock(run, timestamp, exitCode, summary, failures, durationMs)); err != nil {
		return err
	}

	return appendJSONLRecord(jsonlFile, run, timestamp, exitCode, durationMs, summary, failures)
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

func newTestRunState() *testRunState {
	return &testRunState{
		failures:      make([]testFailure, 0),
		failureOutput: make(map[string]*strings.Builder),
	}
}

func failureKey(pkg, test string) string {
	return pkg + "\x00" + test
}

func (state *testRunState) handlePass(event testEvent) {
	if event.Test == "" {
		return
	}
	state.summary.Passes++
}

func (state *testRunState) handleSkip(event testEvent) {
	if event.Test == "" {
		return
	}
	state.summary.Skipped++
}

func (state *testRunState) handleFail(event testEvent) {
	if event.Test == "" {
		return
	}

	state.summary.Failures++
	key := failureKey(event.Package, event.Test)
	output := ""
	if builder, ok := state.failureOutput[key]; ok {
		output = strings.TrimSpace(builder.String())
	} else {
		state.failureOutput[key] = &strings.Builder{}
	}

	state.failures = append(state.failures, testFailure{
		Package: event.Package,
		Test:    event.Test,
		Output:  output,
	})
}

func (state *testRunState) handleOutput(event testEvent) {
	if event.Test == "" {
		return
	}

	key := failureKey(event.Package, event.Test)
	builder, ok := state.failureOutput[key]
	if !ok {
		builder = &strings.Builder{}
		state.failureOutput[key] = builder
	}
	builder.WriteString(event.Output)
}

type testEventHandler func(*testRunState, testEvent)

var testEventHandlers = map[string]testEventHandler{
	"pass":   (*testRunState).handlePass,
	"fail":   (*testRunState).handleFail,
	"skip":   (*testRunState).handleSkip,
	"output": (*testRunState).handleOutput,
}

func (state *testRunState) processEvent(event testEvent) {
	handler, ok := testEventHandlers[event.Action]
	if !ok {
		return
	}
	handler(state, event)
}

func (state *testRunState) finalizeFailureOutputs() {
	for index := range state.failures {
		key := failureKey(state.failures[index].Package, state.failures[index].Test)
		if builder, ok := state.failureOutput[key]; ok {
			state.failures[index].Output = strings.TrimSpace(builder.String())
		}
	}
}

func scanTestEvents(stdout io.Reader, state *testRunState) error {
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var event testEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		state.processEvent(event)
	}
	return scanner.Err()
}

func finishTestCommand(cmd *exec.Cmd, state *testRunState) (int, runSummary, []testFailure, error) {
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			state.finalizeFailureOutputs()
			return exitErr.ExitCode(), state.summary, state.failures, nil
		}
		return 1, state.summary, state.failures, err
	}
	return 0, state.summary, state.failures, nil
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

	state := newTestRunState()
	if err := scanTestEvents(stdout, state); err != nil {
		return 1, state.summary, state.failures, err
	}

	return finishTestCommand(cmd, state)
}

func buildLogHeader(logFile, jsonlFile string) string {
	return strings.Join([]string{
		fmt.Sprintf("# Flaky test run log started %s", time.Now().UTC().Format(time.RFC3339)),
		fmt.Sprintf("# Log file: %s", logFile),
		fmt.Sprintf("# JSONL file: %s", jsonlFile),
		"",
	}, "\n") + "\n"
}

func initRunFiles(logFile, jsonlFile string) error {
	if err := os.WriteFile(logFile, []byte(buildLogHeader(logFile, jsonlFile)), 0o644); err != nil {
		return err
	}
	return os.WriteFile(jsonlFile, nil, 0o644)
}

func logRunResult(run, exitCode int, summary runSummary, durationMs int64) {
	log.Infof(
		"Run %d: %s | %s (%.1fs) — logged",
		run,
		runResultLabel(exitCode),
		formatRunCounts(summary),
		float64(durationMs)/1000,
	)
}

func executeRun(run int, rootDir, logFile, jsonlFile string) (exitCode int, stop bool) {
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

	logRunResult(run, exitCode, summary, durationMs)
	return exitCode, exitCode != 0
}

func main() {
	initLog()

	rootDir := serviceRoot()
	logFile := envOrDefault("FLAKEY_LOG_FILE", filepath.Join(rootDir, defaultLogFile))
	jsonlFile := envOrDefault("FLAKEY_JSONL_FILE", filepath.Join(rootDir, defaultJSONLFile))

	if err := initRunFiles(logFile, jsonlFile); err != nil {
		log.WithError(err).Fatal("failed to initialize output files")
	}

	log.Infof("Logging flaky test runs to %s", logFile)
	log.Infof("Structured run data: %s", jsonlFile)

	run := 0
	for {
		run++
		exitCode, stop := executeRun(run, rootDir, logFile, jsonlFile)
		if stop {
			log.Errorf("Failure on run %d, stopping. See %s", run, logFile)
			os.Exit(exitCode)
		}
	}
}

package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/OliveTin/OliveTin/internal/auth"
	authpublic "github.com/OliveTin/OliveTin/internal/auth/authpublic"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
)

func testingExecutor() (*Executor, *config.Config) {
	cfg := config.DefaultConfig()

	e := DefaultExecutor(cfg)

	a1 := &config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ person }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "person",
				Type: "ascii",
			},
		},
	}

	cfg.Actions = append(cfg.Actions, a1)
	cfg.Sanitize()

	return e, cfg
}

func TestCreateExecutorAndExec(t *testing.T) {
	e, cfg := testingExecutor()

	req := ExecutionRequest{
		AuthenticatedUser: &authpublic.AuthenticatedUser{Username: "MrTickle"},
		Cfg:               cfg,
		Arguments: map[string]string{
			"person": "yourself",
		},
	}

	// Ensure bindings are available and set the binding to the only configured action
	e.RebuildActionMap()
	if len(cfg.Actions) > 0 {
		req.Binding = e.FindBindingWithNoEntity(cfg.Actions[0])
	}

	assert.NotNil(t, e, "Create an executor")

	wg, _ := e.ExecRequest(&req)
	wg.Wait()

	assert.Equal(t, int32(0), req.logEntry.ExitCode, "Exit code is zero")
}

func TestStepRequestActionPopulateLogEntryResolvesEntityTemplates(t *testing.T) {
	req := &ExecutionRequest{
		logEntry: &InternalLogEntry{},
		Binding: &ActionBinding{
			Action: &config.Action{
				Title: "Do something with {{ project.name }}",
				Icon:  "{{ project.icon }}",
			},
			Entity: &entities.Entity{
				Data: map[string]any{
					"name": "foo",
					"icon": "🐰",
				},
				UniqueKey: "foo-key",
			},
		},
	}

	stepRequestActionPopulateLogEntry(req)

	assert.Equal(t, "Do something with foo", req.logEntry.ActionTitle)
	assert.Equal(t, "🐰", req.logEntry.ActionIcon)
	assert.Equal(t, "Do something with {{ project.name }}", req.logEntry.ActionConfigTitle)
	assert.Equal(t, "foo-key", req.logEntry.EntityPrefix)
}

func TestExecNonExistant(t *testing.T) {
	e, cfg := testingExecutor()

	req := ExecutionRequest{
		//		Binding:  e.FindBindingWithNoEntity("waffles"),
		logEntry: &InternalLogEntry{},
		Cfg:      cfg,
	}

	wg, _ := e.ExecRequest(&req)
	wg.Wait()

	assert.Equal(t, int32(-1337), req.logEntry.ExitCode, "Log entry is set to an internal error code")
	assert.Equal(t, "&#x1f4a9;", req.logEntry.ActionIcon, "Log entry icon is a poop (not found)")
}

func TestArgumentNameCamelCase(t *testing.T) {
	req := newExecRequest()
	req.Binding.Action = &config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ personName }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "personName",
				Type: "ascii",
			},
		},
	}

	req.Arguments = map[string]string{
		"personName": "Fred",
	}

	out, err := parseActionArguments(req)

	assert.Equal(t, "echo 'Tickling Fred'", out)
	assert.Nil(t, err)
}

func TestArgumentNameSnakeCase(t *testing.T) {
	req := newExecRequest()
	req.Binding.Action = &config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ person_name }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "person_name",
				Type: "ascii",
			},
		},
	}

	req.Arguments = map[string]string{
		"person_name": "Fred",
	}

	out, err := parseActionArguments(req)

	assert.Equal(t, "echo 'Tickling Fred'", out)
	assert.Nil(t, err)
}

func TestGetLogsEmpty(t *testing.T) {
	e, cfg := testingExecutor()

	assert.Equal(t, int64(10), cfg.LogHistoryPageSize, "Logs page size should be 10")

	logs, paging := e.GetLogTrackingIds(0, 10)

	assert.NotNil(t, logs, "Logs should not be nil")
	assert.Equal(t, 0, len(logs), "No logs yet")
	assert.Equal(t, int64(0), paging.CountRemaining, "There should be no remaining logs")
}

func TestGetLogsLessThanPageSize(t *testing.T) {
	e, cfg := testingExecutor()

	cfg.Actions = append(cfg.Actions, &config.Action{
		Title: "blat",
		Shell: "date",
	})
	cfg.Sanitize()

	// Rebuild action map to include newly added action
	e.RebuildActionMap()

	assert.Equal(t, int64(10), cfg.LogHistoryPageSize, "Logs page size should be 10")

	logEntries, paging := e.GetLogTrackingIds(0, 10)

	assert.Equal(t, 0, len(logEntries), "There should be 0 logs")
	assert.Zero(t, paging.CountRemaining, "There should be no remaining logs")

	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)

	logEntries, paging = e.GetLogTrackingIds(0, 10)

	assert.Equal(t, 7, len(logEntries), "There should be 7 logs")
	assert.Zero(t, paging.CountRemaining, "There should be no remaining logs")

	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)

	logEntries, paging = e.GetLogTrackingIds(0, 10)

	assert.Equal(t, 10, len(logEntries), "There should be 10 logs")
	assert.Equal(t, int64(2), paging.CountRemaining, "There should be 1 remaining logs")
}

func execNewReqAndWait(e *Executor, title string, cfg *config.Config) {
	req := &ExecutionRequest{
		//		ActionTitle: title,
		Cfg: cfg,
	}

	// Ensure we have a binding for the requested title
	e.RebuildActionMap()
	var action *config.Action
	for _, a := range cfg.Actions {
		if a.Title == title {
			action = a
			break
		}
	}
	if action != nil {
		req.Binding = e.FindBindingWithNoEntity(action)
	}

	wg, _ := e.ExecRequest(req)
	wg.Wait()
}

func TestGetPagingIndexes(t *testing.T) {
	assert.Zero(t, getPagingStartIndex(5, 0), "Testing start index from empty list")
	assert.Equal(t, int64(4), getPagingStartIndex(5, 10), "Testing start index from mid point")
	assert.Equal(t, int64(9), getPagingStartIndex(-1, 10), "Testing start index with negative offset")
	assert.Equal(t, int64(0), getPagingStartIndex(15, 10), "Testing start index with large offset")
	assert.Equal(t, int64(9), getPagingStartIndex(0, 10), "Testing start index with zero count")
}

func TestUnsetRequiredArgument(t *testing.T) {
	req := newExecRequest()
	req.Binding.Action = &config.Action{
		Title: "Print your name",
		Shell: "echo 'Your name is: {{ name }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "name",
				Type: "ascii",
			},
		},
	}

	req.Arguments = map[string]string{}

	out, err := parseActionArguments(req)

	assert.Equal(t, "", out)
	assert.NotNil(t, err)
}

func TestUnusedArgumentStillPassesTypeSafetyCheck(t *testing.T) {
	req := newExecRequest()
	req.Binding.Action = &config.Action{
		Title: "Print your name",
		Shell: "echo 'Your name is: {{ name }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "name",
				Type: "ascii",
			},
			{
				Name: "age",
				Type: "int",
			},
		},
	}

	req.Arguments = map[string]string{
		"name": "Fred",
		"age":  "Not an integer",
	}

	out, err := parseActionArguments(req)

	assert.Equal(t, "", out)
	assert.NotNil(t, err)
}

// https://github.com/OliveTin/OliveTin/issues/564
func TestMangleInvalidArgumentValues(t *testing.T) {
	e, cfg := testingExecutor()

	a1 := &config.Action{
		Title: "Validate my date without seconds because I am from an Android phone",
		Shell: "echo 'The date is: {{ date }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "date",
				Type: "datetime",
			},
		},
	}

	cfg.Actions = append(cfg.Actions, a1)
	cfg.Sanitize()

	// Build bindings for newly added action
	e.RebuildActionMap()

	req := ExecutionRequest{
		//		Action:            a1,
		AuthenticatedUser: auth.UserFromSystem(cfg, "testuser"),
		Cfg:               cfg,
		Arguments: map[string]string{
			"date": "1990-01-10T12:00", // Invalid format, should be without seconds
		},
	}

	// Set binding to our appended action
	req.Binding = e.FindBindingWithNoEntity(a1)

	wg, _ := e.ExecRequest(&req)
	wg.Wait()

	assert.NotNil(t, req.logEntry, "Log entry should not be nil")
	assert.Equal(t, req.logEntry.Output, "The date is: 1990-01-10T12:00:00\n", "Date should be mangled to a valid format")

}

func TestWebhookRejectsShellExecution(t *testing.T) {
	cfg := config.DefaultConfig()
	e := DefaultExecutor(cfg)
	a1 := &config.Action{
		Title: "Webhook Shell Reject",
		Shell: "echo '{{ msg }}'",
		Arguments: []config.ActionArgument{
			{Name: "msg", Type: "ascii"},
		},
	}
	cfg.Actions = append(cfg.Actions, a1)
	cfg.Sanitize()
	e.RebuildActionMap()

	req := ExecutionRequest{
		Tags:              []string{"webhook"},
		AuthenticatedUser: auth.UserFromSystem(cfg, "webhook"),
		Cfg:               cfg,
		Arguments:         map[string]string{"msg": "hello"},
		Binding:           e.FindBindingWithNoEntity(a1),
	}

	wg, _ := e.ExecRequest(&req)
	wg.Wait()

	assert.NotNil(t, req.logEntry)
	assert.Equal(t, int32(-1337), req.logEntry.ExitCode)
	assert.Contains(t, req.logEntry.Output, "webhooks cannot use Shell execution")
}

func TestWebhookAllowsExecExecution(t *testing.T) {
	cfg := config.DefaultConfig()
	e := DefaultExecutor(cfg)
	a1 := &config.Action{
		Title: "Webhook Exec OK",
		Exec:  []string{"echo", "{{ msg }}"},
		Arguments: []config.ActionArgument{
			{Name: "msg", Type: "ascii"},
		},
	}
	cfg.Actions = append(cfg.Actions, a1)
	cfg.Sanitize()
	e.RebuildActionMap()

	req := ExecutionRequest{
		Tags:              []string{"webhook"},
		AuthenticatedUser: auth.UserFromSystem(cfg, "webhook"),
		Cfg:               cfg,
		Arguments:         map[string]string{"msg": "hello"},
		Binding:           e.FindBindingWithNoEntity(a1),
	}

	wg, _ := e.ExecRequest(&req)
	wg.Wait()

	assert.NotNil(t, req.logEntry)
	assert.Equal(t, int32(0), req.logEntry.ExitCode)
	assert.Contains(t, req.logEntry.Output, "hello")
}

func TestWebhookRejectsShellAfterCompleted(t *testing.T) {
	cfg := config.DefaultConfig()
	e := DefaultExecutor(cfg)
	a1 := &config.Action{
		Title:               "Webhook After Shell Reject",
		Exec:                []string{"echo", "{{ msg }}"},
		ShellAfterCompleted: "echo after",
		Arguments: []config.ActionArgument{
			{Name: "msg", Type: "ascii"},
		},
	}
	cfg.Actions = append(cfg.Actions, a1)
	cfg.Sanitize()
	e.RebuildActionMap()

	req := ExecutionRequest{
		Tags:              []string{"webhook"},
		AuthenticatedUser: auth.UserFromSystem(cfg, "webhook"),
		Cfg:               cfg,
		Arguments:         map[string]string{"msg": "hello"},
		Binding:           e.FindBindingWithNoEntity(a1),
	}

	wg, _ := e.ExecRequest(&req)
	wg.Wait()

	assert.NotNil(t, req.logEntry)
	assert.Contains(t, req.logEntry.Output, "webhooks cannot use shellAfterCompleted")
}

func TestShellAfterCompletedUsesOutputEnvSafely(t *testing.T) {
	cfg := config.DefaultConfig()
	e := DefaultExecutor(cfg)
	injectedPath := filepath.Join(t.TempDir(), "olivetin-injected")
	expectedMainOutput := "'; touch " + injectedPath + "; echo '"
	a1 := &config.Action{
		Title:               "After completion escape",
		Shell:               "printf %s \"" + expectedMainOutput + "\"",
		ShellAfterCompleted: "printf %s {{ output }}",
	}
	cfg.Actions = append(cfg.Actions, a1)
	cfg.Sanitize()
	e.RebuildActionMap()

	req := ExecutionRequest{
		AuthenticatedUser: auth.UserFromSystem(cfg, "cron"),
		Cfg:               cfg,
		Binding:           e.FindBindingWithNoEntity(a1),
	}

	wg, _ := e.ExecRequest(&req)
	wg.Wait()

	assert.NotNil(t, req.logEntry)
	assert.Equal(t, int32(0), req.logEntry.ExitCode)
	assert.True(t, strings.HasPrefix(req.logEntry.Output, expectedMainOutput))
	assert.Contains(t, req.logEntry.Output, "OliveTin::shellAfterCompleted stdout\n"+expectedMainOutput)
	_, err := os.Stat(injectedPath)
	assert.True(t, os.IsNotExist(err), "shellAfterCompleted must not execute injected commands from output")
}

func TestFilterToDefinedArgumentsOnly(t *testing.T) {
	req := newExecRequest()
	req.Binding.Action = &config.Action{
		Title: "Filter test",
		Shell: "echo '{{ name }}'",
		Arguments: []config.ActionArgument{
			{Name: "name", Type: "ascii"},
		},
	}
	req.Arguments = map[string]string{
		"name":            "Alice",
		"webhook_path":    "/malicious/$(id)",
		"extra_undefined": "ignored",
	}

	filterToDefinedArgumentsOnly(req)

	assert.Equal(t, "Alice", req.Arguments["name"])
	assert.Empty(t, req.Arguments["webhook_path"])
	assert.Empty(t, req.Arguments["extra_undefined"])
}

func TestFilterToDefinedArgumentsDropsReservedPrefixArgs(t *testing.T) {
	req := newExecRequest()
	req.Binding.Action = &config.Action{
		Title:     "Filter test",
		Shell:     "echo test",
		Arguments: []config.ActionArgument{},
	}
	req.Arguments = map[string]string{
		"ot_executionTrackingId": "track-123",
		"ot_username":            "webhook",
	}

	filterToDefinedArgumentsOnly(req)

	assert.Empty(t, req.Arguments["ot_executionTrackingId"])
	assert.Empty(t, req.Arguments["ot_username"])
}

func TestStepParseArgsInjectsSystemArgsAfterFiltering(t *testing.T) {
	req := newExecRequest()
	req.TrackingID = "server-track-456"
	req.AuthenticatedUser = &authpublic.AuthenticatedUser{Username: "alice"}
	req.Binding.Action = &config.Action{
		Title: "Filter then inject",
		Shell: "echo test",
		Arguments: []config.ActionArgument{
			{Name: "name", Type: "ascii"},
		},
	}
	req.Arguments = map[string]string{
		"name":                   "Alice",
		"ot_executionTrackingId": "attacker-track",
		"ot_username":            "mallory",
		"ot_custom":              "polluted",
	}

	assert.True(t, stepParseArgs(req))
	assert.Equal(t, "Alice", req.Arguments["name"])
	assert.Equal(t, "server-track-456", req.Arguments["ot_executionTrackingId"])
	assert.Equal(t, "alice", req.Arguments["ot_username"])
	assert.Empty(t, req.Arguments["ot_custom"])
}

func TestStepParseArgsDropsReservedPrefixArgsFromEnvironment(t *testing.T) {
	req := newExecRequest()
	req.TrackingID = "server-track-456"
	req.AuthenticatedUser = &authpublic.AuthenticatedUser{Username: "alice@example.com"}
	req.Binding.Action = &config.Action{
		Title:     "No reserved prefix pollution",
		Shell:     "echo test",
		Arguments: []config.ActionArgument{},
	}
	req.Arguments = map[string]string{
		"ot_custom": "polluted",
	}

	assert.True(t, stepParseArgs(req))
	env := buildEnv(req.Arguments)

	assert.False(t, containsEnvPrefix(env, "OT_CUSTOM="))
	assert.True(t, containsEnvPrefix(env, "OT_USERNAME=alice@example.com"))
	assert.True(t, containsEnvPrefix(env, "OT_EXECUTIONTRACKINGID=server-track-456"))
}

func TestSystemArgumentDefinitionsAreReservedAndShellSafe(t *testing.T) {
	unsafeTypes := map[string]struct{}{
		"email":                     {},
		"password":                  {},
		"raw_string_multiline":      {},
		"url":                       {},
		"very_dangerous_raw_string": {},
	}
	seen := map[string]struct{}{}

	for _, arg := range systemArgumentDefinitions {
		assert.True(t, strings.HasPrefix(arg.Name, config.ReservedArgumentNamePrefix))
		assert.NotEmpty(t, arg.Type)
		assert.True(t, arg.RejectNull)

		_, duplicate := seen[arg.Name]
		assert.False(t, duplicate, "duplicate system argument definition %q", arg.Name)
		seen[arg.Name] = struct{}{}

		_, unsafe := unsafeTypes[arg.Type]
		assert.False(t, unsafe, "system argument %q uses unsafe type %q", arg.Name, arg.Type)
	}
}

func TestValidatedSystemArgsMatchesSystemArgumentDefinitions(t *testing.T) {
	req := newExecRequest()
	req.TrackingID = "server-track-456"
	req.AuthenticatedUser = &authpublic.AuthenticatedUser{Username: "alice@example.com"}

	args, err := validatedSystemArgs(req)

	assert.Nil(t, err)
	assert.Len(t, args, len(systemArgumentDefinitions))
	for _, arg := range systemArgumentDefinitions {
		assert.Contains(t, args, arg.Name)
	}
}

func TestBuildShellAfterArgsOnlyAddsExpectedNonSystemArgs(t *testing.T) {
	req := newExecRequest()
	req.logEntry = &InternalLogEntry{
		Output:   "hello",
		ExitCode: 7,
	}
	req.TrackingID = "server-track-456"
	req.AuthenticatedUser = &authpublic.AuthenticatedUser{Username: "alice@example.com"}
	req.Binding.Action = &config.Action{ShellAfterCompleted: "echo test"}

	args, err := buildShellAfterArgs(req)

	assert.Nil(t, err)
	assert.Len(t, args, len(systemArgumentDefinitions)+2)
	assert.Contains(t, args, "output")
	assert.Contains(t, args, "exitCode")
	for _, arg := range systemArgumentDefinitions {
		assert.Contains(t, args, arg.Name)
	}
}

func TestStepParseArgsAllowsEmailUsernameSystemArg(t *testing.T) {
	req := newExecRequest()
	req.logEntry = &InternalLogEntry{}
	req.TrackingID = "server-track-456"
	req.AuthenticatedUser = &authpublic.AuthenticatedUser{Username: "alice@example.com"}
	req.Binding.Action = &config.Action{
		Title:     "Email username",
		Shell:     "echo test",
		Arguments: []config.ActionArgument{},
	}

	assert.True(t, stepParseArgs(req))
	assert.Equal(t, "alice@example.com", req.Arguments["ot_username"])
}

func TestStepParseArgsFailsWhenUsernameSystemArgIsInvalid(t *testing.T) {
	req := newExecRequest()
	req.logEntry = &InternalLogEntry{}
	req.TrackingID = "server-track-456"
	req.AuthenticatedUser = &authpublic.AuthenticatedUser{Username: "alice;id"}
	req.Binding.Action = &config.Action{
		Title:     "Invalid system arg",
		Shell:     "echo test",
		Arguments: []config.ActionArgument{},
	}

	assert.False(t, stepParseArgs(req))
	assert.Contains(t, req.logEntry.Output, `system argument "ot_username" failed validation`)
	assert.Empty(t, req.Arguments["ot_username"])
}

func TestStepParseArgsFailsWhenTrackingIDSystemArgIsInvalid(t *testing.T) {
	req := newExecRequest()
	req.logEntry = &InternalLogEntry{}
	req.TrackingID = "track/../../bad"
	req.AuthenticatedUser = &authpublic.AuthenticatedUser{Username: "alice"}
	req.Binding.Action = &config.Action{
		Title:     "Invalid tracking ID",
		Shell:     "echo test",
		Arguments: []config.ActionArgument{},
	}

	assert.False(t, stepParseArgs(req))
	assert.Contains(t, req.logEntry.Output, `system argument "ot_executionTrackingId" failed validation`)
	assert.Empty(t, req.Arguments["ot_executionTrackingId"])
}

func TestBuildShellAfterArgsUsesValidatedSystemArgs(t *testing.T) {
	req := newExecRequest()
	req.logEntry = &InternalLogEntry{
		Output:   "hello",
		ExitCode: 7,
	}
	req.TrackingID = "server-track-456"
	req.AuthenticatedUser = &authpublic.AuthenticatedUser{Username: "alice@example.com"}
	req.Binding.Action = &config.Action{
		Title:               "Shell after",
		ShellAfterCompleted: "echo test",
	}

	args, err := buildShellAfterArgs(req)

	assert.Nil(t, err)
	assert.Equal(t, "alice@example.com", args["ot_username"])
	assert.Equal(t, "server-track-456", args["ot_executionTrackingId"])
	assert.Equal(t, "hello", args["output"])
	assert.Equal(t, "7", args["exitCode"])
}

func TestBuildShellAfterArgsFailsWhenSystemArgIsInvalid(t *testing.T) {
	req := newExecRequest()
	req.logEntry = &InternalLogEntry{}
	req.TrackingID = "server-track-456"
	req.AuthenticatedUser = &authpublic.AuthenticatedUser{Username: "alice;id"}
	req.Binding.Action = &config.Action{
		Title:               "Shell after invalid username",
		ShellAfterCompleted: "echo test",
	}

	args, err := buildShellAfterArgs(req)

	assert.Nil(t, args)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), `system argument "ot_username" failed validation`)
}

func containsEnvPrefix(env []string, prefix string) bool {
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return true
		}
	}

	return false
}

func TestTriggerExecutesTriggeredAction(t *testing.T) {
	cfg := config.DefaultConfig()
	e := DefaultExecutor(cfg)
	helloAction := &config.Action{
		Title: "Hello world",
		Shell: "echo 'Hello World!'",
	}
	triggerAction := &config.Action{
		Title:    "Simple action that triggers another action",
		Shell:    "echo 'Hi'",
		Triggers: []string{"Hello world"},
	}
	cfg.Actions = append(cfg.Actions, helloAction, triggerAction)
	cfg.Sanitize()
	e.RebuildActionMap()

	finishedTitles := make(chan string, 4)
	collector := &executionFinishedCollector{ch: finishedTitles}
	e.AddListener(collector)

	req := &ExecutionRequest{
		AuthenticatedUser: auth.UserFromSystem(cfg, "testuser"),
		Cfg:               cfg,
		Binding:           e.FindBindingWithNoEntity(triggerAction),
	}
	wg, _ := e.ExecRequest(req)
	wg.Wait()

	var got []string
	for i := 0; i < 2; i++ {
		select {
		case title := <-finishedTitles:
			got = append(got, title)
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for execution %d; got %v", i+1, got)
		}
	}
	assert.Contains(t, got, "Hello world", "triggered action must run")
	assert.Contains(t, got, "Simple action that triggers another action", "triggering action must run")
}

func TestTriggerUnknownActionTitleSkipsWithoutPanic(t *testing.T) {
	cfg := config.DefaultConfig()
	e := DefaultExecutor(cfg)
	triggerAction := &config.Action{
		Title:    "Action with bad trigger",
		Shell:    "echo 'ok'",
		Triggers: []string{"Nonexistent action"},
	}
	cfg.Actions = append(cfg.Actions, triggerAction)
	cfg.Sanitize()
	e.RebuildActionMap()

	finishedTitles := make(chan string, 4)
	collector := &executionFinishedCollector{ch: finishedTitles}
	e.AddListener(collector)

	req := &ExecutionRequest{
		AuthenticatedUser: auth.UserFromSystem(cfg, "testuser"),
		Cfg:               cfg,
		Binding:           e.FindBindingWithNoEntity(triggerAction),
	}
	wg, _ := e.ExecRequest(req)
	wg.Wait()

	var got []string
	select {
	case title := <-finishedTitles:
		got = append(got, title)
	case <-time.After(500 * time.Millisecond):
	}
	assert.Len(t, got, 1, "only the triggering action runs; unknown trigger is skipped")

	if len(got) > 0 {
		assert.Equal(t, "Action with bad trigger", got[0])
	}
}

type executionFinishedCollector struct {
	ch chan string
}

func (c *executionFinishedCollector) OnExecutionStarted(_ *InternalLogEntry) {}

func (c *executionFinishedCollector) OnExecutionFinished(entry *InternalLogEntry) {
	c.ch <- entry.ActionTitle
}

func (c *executionFinishedCollector) OnOutputChunk(_ []byte, _ string) {}

func (c *executionFinishedCollector) OnActionMapRebuilt() {}

func TestSanitizeLogFilename(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"Echo Test", "Echo Test"},
		{"Create/update Monthly Report", "Create_update Monthly Report"},
		{`path\with\backslashes`, "path_with_backslashes"},
		{`a:b*c?d"e<f>g|h`, "a_b_c_d_e_f_g_h"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, sanitizeLogFilename(tt.title), "title=%q", tt.title)
	}
}

func TestStepSaveLogSanitizesSlashInTitle(t *testing.T) {
	resultsDir := t.TempDir()
	outputDir := t.TempDir()
	started := time.Unix(1714333384, 0)
	trackingID := "5e2dc9e5-b6b3-445b-bff9-c2082b0bbbb2"
	title := "Create/update Monthly Report"

	req := &ExecutionRequest{
		Cfg: &config.Config{
			SaveLogs: config.SaveLogsConfig{
				ResultsDirectory: resultsDir,
				OutputDirectory:  outputDir,
			},
		},
		Binding: &ActionBinding{
			Action: &config.Action{},
		},
		logEntry: &InternalLogEntry{
			ActionTitle:         title,
			DatetimeStarted:     started,
			ExecutionTrackingID: trackingID,
			Output:              "report ok",
		},
	}

	assert.True(t, stepSaveLog(req))

	expectedBase := "Create_update Monthly Report.1714333384." + trackingID
	resultsPath := filepath.Join(resultsDir, expectedBase+".yaml")
	outputPath := filepath.Join(outputDir, expectedBase+".log")

	assert.FileExists(t, resultsPath)
	assert.FileExists(t, outputPath)

	resultsEntries, err := os.ReadDir(resultsDir)
	assert.NoError(t, err)
	assert.Len(t, resultsEntries, 1, "results file must be flat under resultsDirectory, not a subdirectory")

	data, err := os.ReadFile(resultsPath)
	assert.NoError(t, err)
	assert.Contains(t, string(data), title, "YAML content keeps the original action title")

	output, err := os.ReadFile(outputPath)
	assert.NoError(t, err)
	assert.Equal(t, "report ok", string(output))
}

func TestStepSaveLogKeepsSafeTitleFilename(t *testing.T) {
	resultsDir := t.TempDir()
	started := time.Unix(1714333384, 0)
	trackingID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

	req := &ExecutionRequest{
		Cfg: &config.Config{
			SaveLogs: config.SaveLogsConfig{
				ResultsDirectory: resultsDir,
			},
		},
		Binding: &ActionBinding{
			Action: &config.Action{},
		},
		logEntry: &InternalLogEntry{
			ActionTitle:         "Echo Test",
			DatetimeStarted:     started,
			ExecutionTrackingID: trackingID,
		},
	}

	assert.True(t, stepSaveLog(req))

	expectedPath := filepath.Join(resultsDir, "Echo Test.1714333384."+trackingID+".yaml")
	assert.FileExists(t, expectedPath)
}

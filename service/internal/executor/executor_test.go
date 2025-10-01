package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
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
		AuthenticatedUser: &acl.AuthenticatedUser{Username: "Mr Tickle"},
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
	a1 := &config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ personName }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "personName",
				Type: "ascii",
			},
		},
	}

	values := map[string]string{
		"personName": "Fred",
	}

	out, err := parseActionArguments(values, a1, nil)

	assert.Equal(t, "echo 'Tickling Fred'", out)
	assert.Nil(t, err)
}

func TestArgumentNameSnakeCase(t *testing.T) {
	a1 := &config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ person_name }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "person_name",
				Type: "ascii",
			},
		},
	}

	values := map[string]string{
		"person_name": "Fred",
	}

	out, err := parseActionArguments(values, a1, nil)

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
	a1 := &config.Action{
		Title: "Print your name",
		Shell: "echo 'Your name is: {{ name }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "name",
				Type: "ascii",
			},
		},
	}

	values := map[string]string{}

	out, err := parseActionArguments(values, a1, nil)

	assert.Equal(t, "", out)
	assert.NotNil(t, err)
}

func TestUnusedArgumentStillPassesTypeSafetyCheck(t *testing.T) {
	a1 := &config.Action{
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

	values := map[string]string{
		"name": "Fred",
		"age":  "Not an integer",
	}

	out, err := parseActionArguments(values, a1, nil)

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
		AuthenticatedUser: acl.UserFromSystem(cfg, "testuser"),
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

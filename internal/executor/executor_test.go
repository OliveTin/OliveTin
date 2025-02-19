package executor

import (
	"github.com/stretchr/testify/assert"
	"testing"

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
		ActionTitle:       "Do some tickles",
		AuthenticatedUser: &acl.AuthenticatedUser{Username: "Mr Tickle"},
		Cfg:               cfg,
		Arguments: map[string]string{
			"person": "yourself",
		},
	}

	assert.NotNil(t, e, "Create an executor")

	wg, _ := e.ExecRequest(&req)
	wg.Wait()

	assert.Equal(t, int32(0), req.logEntry.ExitCode, "Exit code is zero")
}

func TestExecNonExistant(t *testing.T) {
	e, cfg := testingExecutor()

	req := ExecutionRequest{
		ActionTitle: "Waffles",
		logEntry:    &InternalLogEntry{},
		Cfg:         cfg,
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

	out, err := parseActionArguments(a1.Shell, values, a1, a1.Title, "")

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

	out, err := parseActionArguments(a1.Shell, values, a1, a1.Title, "")

	assert.Equal(t, "echo 'Tickling Fred'", out)
	assert.Nil(t, err)
}

func TestGetLogsEmpty(t *testing.T) {
	e, cfg := testingExecutor()

	assert.Equal(t, int64(10), cfg.LogHistoryPageSize, "Logs page size should be 10")

	logs, remaining := e.GetLogTrackingIds(0, 10)

	assert.NotNil(t, logs, "Logs should not be nil")
	assert.Equal(t, 0, len(logs), "No logs yet")
	assert.Equal(t, int64(0), remaining, "There should be no remaining logs")
}

func TestGetLogsLessThanPageSize(t *testing.T) {
	e, cfg := testingExecutor()

	cfg.Actions = append(cfg.Actions, &config.Action{
		Title: "blat",
		Shell: "date",
	})

	assert.Equal(t, int64(10), cfg.LogHistoryPageSize, "Logs page size should be 10")

	logEntries, remaining := e.GetLogTrackingIds(0, 10)

	assert.Equal(t, 0, len(logEntries), "There should be 0 logs")
	assert.Zero(t, remaining, "There should be no remaining logs")

	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)

	logEntries, remaining = e.GetLogTrackingIds(0, 10)

	assert.Equal(t, 7, len(logEntries), "There should be 7 logs")
	assert.Zero(t, remaining, "There should be no remaining logs")

	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)
	execNewReqAndWait(e, "blat", cfg)

	logEntries, remaining = e.GetLogTrackingIds(0, 10)

	assert.Equal(t, 10, len(logEntries), "There should be 10 logs")
	assert.Equal(t, int64(2), remaining, "There should be 1 remaining logs")
}

func execNewReqAndWait(e *Executor, title string, cfg *config.Config) {
	req := &ExecutionRequest{
		ActionTitle: title,
		Cfg:         cfg,
	}

	wg, _ := e.ExecRequest(req)
	wg.Wait()
}

func TestGetPagingIndexes(t *testing.T) {
	assert.Zero(t, getPagingStartIndex(5, 0, 5), "Testing start index from empty list")
	assert.Equal(t, int64(4), getPagingStartIndex(5, 10, 5), "Testing start index from mid point")
	assert.Equal(t, int64(9), getPagingStartIndex(-1, 10, 5), "Testing start index with negative offset")
	assert.Equal(t, int64(0), getPagingStartIndex(15, 10, 5), "Testing start index with large offset")
	assert.Equal(t, int64(9), getPagingStartIndex(0, 10, 0), "Testing start index with zero count")
}

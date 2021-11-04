package executor

import (
	"github.com/stretchr/testify/assert"
	"testing"

	acl "github.com/jamesread/OliveTin/internal/acl"
	config "github.com/jamesread/OliveTin/internal/config"
)

func TestSanitizeUnsafe(t *testing.T) {
	assert.Nil(t, TypeSafetyCheck("", "_zomg_ c:/ haxxor ' bobby tables && rm -rf ", "very_dangerous_raw_string"))
}

func testingExecutor() (*Executor, *config.Config) {
	e := DefaultExecutor()

	cfg := config.DefaultConfig()

	a1 := config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ person }}'",
		Arguments: []config.ActionArgument{
			config.ActionArgument{
				Name: "person",
				Type: "ascii",
			},
		},
	}

	cfg.Actions = append(cfg.Actions, a1)
	config.Sanitize(cfg)

	return e, cfg
}

func TestCreateExecutorAndExec(t *testing.T) {
	e, cfg := testingExecutor()

	req := ExecutionRequest{
		ActionName: "Do some tickles",
		User:       &acl.User{Username: "Mr Tickle"},
		Cfg:        cfg,
		Arguments: map[string]string{
			"person": "yourself",
		},
	}

	e.ExecRequest(&req)

	assert.NotNil(t, e, "Create an executor")

	assert.NotNil(t, e.ExecRequest(&req), "Execute a request")
	assert.Equal(t, int32(0), req.logEntry.ExitCode, "Exit code is zero")
}

func TestExecNonExistant(t *testing.T) {
	e, cfg := testingExecutor()

	req := ExecutionRequest{
		ActionName: "Waffles",
		logEntry:   &InternalLogEntry{},
		Cfg:        cfg,
	}

	e.ExecRequest(&req)

	assert.Equal(t, int32(-1337), req.logEntry.ExitCode, "Log entry is set to an internal error code")
	assert.Equal(t, "", req.logEntry.ActionIcon, "Log entry icon wasnt found")
}

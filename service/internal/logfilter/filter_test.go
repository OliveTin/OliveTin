package logfilter

import (
	"testing"

	"github.com/expr-lang/expr/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileNegatedSearchTerm(t *testing.T) {
	program, err := Compile("!Update")
	require.NoError(t, err)

	assert.False(t, mustMatch(t, program, Record{Action: "Run Update script"}))
	assert.True(t, mustMatch(t, program, Record{Action: "Ping host"}))
}

func TestCompileStatusNotEqual(t *testing.T) {
	program, err := Compile("Status != Completed")
	require.NoError(t, err)

	assert.True(t, mustMatch(t, program, Record{Status: "Blocked"}))
	assert.False(t, mustMatch(t, program, Record{Status: "Completed"}))
}

func TestCompileContainsAndBooleanWords(t *testing.T) {
	program, err := Compile(`Status == Completed and Action contains backup`)
	require.NoError(t, err)

	assert.True(t, mustMatch(t, program, Record{Status: "Completed", Action: "Nightly backup"}))
	assert.False(t, mustMatch(t, program, Record{Status: "Blocked", Action: "Nightly backup"}))
}

func TestCompileRejectsOverlongExpression(t *testing.T) {
	_, err := Compile(string(make([]byte, maxFilterLength+1)))
	require.Error(t, err)
}

func TestCompileRejectsUnknownField(t *testing.T) {
	_, err := Compile(`SecretField == "x"`)
	require.Error(t, err)
}

func mustMatch(t *testing.T, program *vm.Program, record Record) bool {
	t.Helper()
	matched, err := Matches(program, record)
	require.NoError(t, err)
	return matched
}

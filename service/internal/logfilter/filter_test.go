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

func TestCompileNormalizesFieldNamesAndValueTypes(t *testing.T) {
	cases := []struct {
		expression string
		record     Record
		want       bool
	}{
		{"status == completed", Record{Status: "completed"}, true},
		{"ExitCode == 0", Record{ExitCode: 0}, true},
		{"exitcode == 0", Record{ExitCode: 0}, true},
		{"Blocked == true", Record{Blocked: true}, true},
		{"blocked == false", Record{Blocked: false}, true},
	}
	for _, tc := range cases {
		t.Run(tc.expression, func(t *testing.T) {
			program, err := Compile(tc.expression)
			require.NoError(t, err)
			assert.Equal(t, tc.want, mustMatch(t, program, tc.record))
		})
	}
}

func TestCompileRejectsOverlongExpression(t *testing.T) {
	_, err := Compile(string(make([]byte, maxFilterLength+1)))
	require.Error(t, err)
}

func TestCompileRejectsUnknownField(t *testing.T) {
	_, err := Compile(`SecretField == "x"`)
	require.Error(t, err)
}

func TestIncludesReturnsErrorsForMalformedArguments(t *testing.T) {
	_, err := includes("only-one")
	require.Error(t, err)

	_, err = includes(123, "needle")
	require.Error(t, err)
	require.Contains(t, err.Error(), "haystack")

	_, err = includes("hay", 123)
	require.Error(t, err)
	require.Contains(t, err.Error(), "needle")
}

func TestHasTagReturnsErrorsForMalformedArguments(t *testing.T) {
	_, err := hasTag("only-one")
	require.Error(t, err)

	_, err = hasTag("not-tags", "x")
	require.Error(t, err)
	require.Contains(t, err.Error(), "tags")

	_, err = hasTag([]string{"a"}, 123)
	require.Error(t, err)
	require.Contains(t, err.Error(), "needle")
}

func TestMatchesSurfacesIncludesTypeErrors(t *testing.T) {
	program, err := Compile(`Action contains 123`)
	require.NoError(t, err)

	_, err = Matches(program, Record{Action: "test 123"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "needle")
}

func mustMatch(t *testing.T, program *vm.Program, record Record) bool {
	t.Helper()
	matched, err := Matches(program, record)
	require.NoError(t, err)
	return matched
}

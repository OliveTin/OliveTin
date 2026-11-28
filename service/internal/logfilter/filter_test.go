package logfilter

import (
	"testing"
	"time"

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

func TestCompileRejectsExpensiveExpressions(t *testing.T) {
	expensive := []string{
		"count(1..999, count(1..999, len(Output)+#>0) == 7) == 7",
		"len(1..999999) > 0",
		"(1 in 1..999999)",
		"all(Tags, {# == \"x\"})",
		"map(Tags, # + Output) != nil",
		"Output matches \"(a+)+$\"",
		"Output + Output == \"\"",
		"let x = Output; x == \"\"",
		"Tags[0] == \"webhook\"",
		"upper(Output) == \"X\"",
	}

	for _, expression := range expensive {
		program, err := Compile(expression)
		require.Error(t, err, "expression must be rejected: %s", expression)
		assert.Nil(t, program, "expression must not compile: %s", expression)
	}
}

func TestCompileAllowsSupportedExpressions(t *testing.T) {
	record := Record{Status: "Completed", Action: "Backup", User: "admin", Output: "done", Tags: []string{"webhook"}, ExitCode: 0}

	supported := map[string]bool{
		"Backup":                                     true,
		"!Backup":                                    false,
		"Status == \"Completed\"":                    true,
		"Status != \"Completed\"":                    false,
		"action contains back and user == admin":     true,
		"(ExitCode > 0) || !Running":                 true,
		"hasTag(Tags, \"webhook\")":                  true,
		"includes(Output, \"done\") && !Blocked":     true,
		"(Action startsWith \"Back\")":               true,
		"(\"webhook\" in Tags)":                      true,
		"(Status in [\"Completed\", \"Timed out\"])": true,
		"ExitCode == -1":                             false,
	}

	for expression, expected := range supported {
		program, err := Compile(expression)
		require.NoError(t, err, "expression must compile: %s", expression)

		matched, err := Matches(program, record)
		require.NoError(t, err, expression)
		assert.Equal(t, expected, matched, expression)
	}
}

func nestedChainedComparison(depth int) string {
	term := "0"
	for range depth {
		term = "0<(" + term + ")<0"
	}
	return term
}

func TestCompileRejectsNestedChainedComparisonsQuickly(t *testing.T) {
	expressions := []string{
		"unknown(" + nestedChainedComparison(40) + ")",
		nestedChainedComparison(40),
	}

	for _, expression := range expressions {
		start := time.Now()
		program, err := Compile(expression)
		require.Error(t, err)
		assert.Nil(t, program)
		assert.Less(t, time.Since(start), time.Second)
	}
}

package executor

import (
	config "github.com/OliveTin/OliveTin/internal/config"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSanitizeUnsafe(t *testing.T) {
	assert.Nil(t, TypeSafetyCheck("", "_zomg_ c:/ haxxor ' bobby tables && rm -rf ", "very_dangerous_raw_string"))
}

func TestSanitizeUnimplemented(t *testing.T) {
	err := TypeSafetyCheck("", "I am a happy little argument", "greeting_type")

	assert.NotNil(t, err, "Test an argument type that does not exist")
}

func TestArgumentNameNumbers(t *testing.T) {
	a1 := config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ person1name }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "person1name",
				Type: "ascii",
			},
		},
	}

	values := map[string]string{
		"person1name": "Fred",
	}

	out, err := parseActionArguments(a1.Shell, values, &a1, a1.Title, "")

	assert.Equal(t, "echo 'Tickling Fred'", out)
	assert.Nil(t, err)
}

func TestArgumentNotProvided(t *testing.T) {
	a1 := config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ personName }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "person",
				Type: "ascii",
			},
		},
	}

	values := map[string]string{}

	out, err := parseActionArguments(a1.Shell, values, &a1, a1.Title, "")

	assert.Equal(t, "", out)
	assert.Equal(t, err.Error(), "Required arg not provided: personName")
}

func TestTypeSafetyCheckUrl(t *testing.T) {
	assert.Nil(t, TypeSafetyCheck("test1", "http://google.com", "url"), "Test URL: google.com")
	assert.Nil(t, TypeSafetyCheck("test2", "http://technowax.net:80?foo=bar", "url"), "Test URL: technowax.net with query arguments")
	assert.Nil(t, TypeSafetyCheck("test3", "http://localhost:80?foo=bar", "url"), "Test URL: localhost with query arguments")
	assert.NotNil(t, TypeSafetyCheck("test4", "http://lo  host:80", "url"), "Test a badly formed URL")
	assert.NotNil(t, TypeSafetyCheck("test5", "12345", "url"), "Test a badly formed URL")
	assert.NotNil(t, TypeSafetyCheck("test6", "_!23;", "url"), "Test a badly formed URL")
}

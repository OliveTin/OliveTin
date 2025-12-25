package executor

import (
	"fmt"
	"strings"

	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/entities"
	log "github.com/sirupsen/logrus"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeUnsafe(t *testing.T) {
	assert.Nil(t, TypeSafetyCheck("", "_zomg_ c:/ haxxor ' bobby tables && rm -rf ", "very_dangerous_raw_string"))
}

func TestSanitizeUnimplemented(t *testing.T) {
	err := TypeSafetyCheck("", "I am a happy little argument", "greeting_type")

	assert.NotNil(t, err, "Test an argument type that does not exist")
}

func TestValidateArgumentCheckboxDefaultValues(t *testing.T) {
	arg := config.ActionArgument{
		Name: "confirm",
		Type: "checkbox",
	}
	action := config.Action{
		Title: "Test checkbox default values",
	}

	// Default checkbox values without choices should accept "1" and "0"
	err := ValidateArgument(&arg, "1", &action)
	assert.Nil(t, err, "Expected checkbox value \"1\" to be accepted without choices")

	err = ValidateArgument(&arg, "0", &action)
	assert.Nil(t, err, "Expected checkbox value \"0\" to be accepted without choices")
}

func TestMangleCheckboxValueWithChoices(t *testing.T) {
	log.SetLevel(log.PanicLevel)

	arg := config.ActionArgument{
		Name: "confirm",
		Type: "checkbox",
		Choices: []config.ActionArgumentChoice{
			{Title: "Enabled", Value: "on"},
			{Title: "Disabled", Value: "off"},
		},
	}

	// When the incoming value matches a choice title, it should be mapped to the choice value
	out := mangleCheckboxValue(&arg, "Enabled", "Test action")
	assert.Equal(t, "on", out, "Expected checkbox title to be mangled to its value")

	out = mangleCheckboxValue(&arg, "Disabled", "Test action")
	assert.Equal(t, "off", out, "Expected checkbox title to be mangled to its value")

	// When there is no matching title, the value should be returned unchanged
	out = mangleCheckboxValue(&arg, "something-else", "Test action")
	assert.Equal(t, "something-else", out, "Expected non-matching value to be returned unchanged")
}

func TestMangleArgumentValueCheckbox(t *testing.T) {
	log.SetLevel(log.PanicLevel)

	arg := config.ActionArgument{
		Name: "confirm",
		Type: "checkbox",
		Choices: []config.ActionArgumentChoice{
			{Title: "Yes", Value: "true-value"},
			{Title: "No", Value: "false-value"},
		},
	}

	out := MangleArgumentValue(&arg, "Yes", "Test action")
	assert.Equal(t, "true-value", out, "Expected MangleArgumentValue to delegate to mangleCheckboxValue for checkbox types")

	out = MangleArgumentValue(&arg, "No", "Test action")
	assert.Equal(t, "false-value", out)

	// For non-matching values, it should return the original value
	out = MangleArgumentValue(&arg, "maybe", "Test action")
	assert.Equal(t, "maybe", out)
}

func TestValidateArgumentCheckboxWithChoices(t *testing.T) {
	log.SetLevel(log.PanicLevel)

	arg := config.ActionArgument{
		Name: "confirm",
		Type: "checkbox",
		Choices: []config.ActionArgumentChoice{
			{Title: "Enabled", Value: "on"},
			{Title: "Disabled", Value: "off"},
		},
	}
	action := config.Action{
		Title: "Test checkbox with choices",
	}

	// Titles should be accepted once mangled to their values
	err := ValidateArgument(&arg, "Enabled", &action)
	assert.Nil(t, err, "Expected checkbox title \"Enabled\" to be accepted after mangling to choice value")

	err = ValidateArgument(&arg, "Disabled", &action)
	assert.Nil(t, err, "Expected checkbox title \"Disabled\" to be accepted after mangling to choice value")

	// Unknown titles should be rejected because they do not match any choice value
	err = ValidateArgument(&arg, "Maybe", &action)
	assert.NotNil(t, err, "Expected unknown checkbox title to be rejected against choices")
}

func TestArgumentValueNullable(t *testing.T) {
	a1 := config.Action{
		Title: "Release the hounds",
		Shell: "echo 'Releasing {{ count }} hounds'",
		Arguments: []config.ActionArgument{
			{
				Name: "count",
				Type: "int",
			},
		},
	}

	values := map[string]string{
		"count": "",
	}

	out, err := parseActionArguments(values, &a1, nil)

	assert.Equal(t, "echo 'Releasing  hounds'", out)
	assert.Nil(t, err)

	a1.Arguments[0].RejectNull = true

	_, err = parseActionArguments(values, &a1, nil)

	assert.NotNil(t, err)
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

	out, err := parseActionArguments(values, &a1, nil)

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

	out, err := parseActionArguments(values, &a1, nil)

	assert.Equal(t, "", out)
	assert.Equal(t, err.Error(), "required arg not provided: personName")
}

func TestExecArrayParsing(t *testing.T) {
	a1 := config.Action{
		Title:     "List files",
		Exec:      []string{"ls", "-alh"},
		Arguments: []config.ActionArgument{},
	}

	values := map[string]string{}

	out, err := parseActionExec(values, &a1, nil)

	assert.Nil(t, err)
	assert.Equal(t, []string{"ls", "-alh"}, out)
}

func TestExecArrayWithTemplateReplacement(t *testing.T) {
	a1 := config.Action{
		Title: "List specific path",
		Exec:  []string{"ls", "-alh", "{{path}}"},
		Arguments: []config.ActionArgument{
			{
				Name: "path",
				Type: "ascii_identifier",
			},
		},
	}

	values := map[string]string{
		"path": "tmp",
	}

	out, err := parseActionExec(values, &a1, nil)

	assert.Nil(t, err)
	assert.Equal(t, []string{"ls", "-alh", "tmp"}, out)
}

func TestCheckShellArgumentSafetyWithURL(t *testing.T) {
	a1 := config.Action{
		Title: "Download file",
		Shell: "curl {{url}}",
		Arguments: []config.ActionArgument{
			{
				Name: "url",
				Type: "url",
			},
		},
	}

	err := checkShellArgumentSafety(&a1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unsafe argument type 'url' cannot be used with Shell execution")
	assert.Contains(t, err.Error(), "https://docs.olivetin.app/action_execution/shellvsexec.html")
}

func TestCheckShellArgumentSafetyWithEmail(t *testing.T) {
	a1 := config.Action{
		Title: "Send email",
		Shell: "sendmail {{email}}",
		Arguments: []config.ActionArgument{
			{
				Name: "email",
				Type: "email",
			},
		},
	}

	err := checkShellArgumentSafety(&a1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unsafe argument type 'email' cannot be used with Shell execution")
}

func TestCheckShellArgumentSafetyWithExec(t *testing.T) {
	a1 := config.Action{
		Title: "Download file",
		Exec:  []string{"curl", "{{url}}"},
		Arguments: []config.ActionArgument{
			{
				Name: "url",
				Type: "url",
			},
		},
	}

	err := checkShellArgumentSafety(&a1)
	assert.Nil(t, err)
}

func TestCheckShellArgumentSafetyWithSafeTypes(t *testing.T) {
	a1 := config.Action{
		Title: "List files",
		Shell: "ls {{path}}",
		Arguments: []config.ActionArgument{
			{
				Name: "path",
				Type: "ascii_identifier",
			},
		},
	}

	err := checkShellArgumentSafety(&a1)
	assert.Nil(t, err)
}

func TestTypeSafetyCheckUrl(t *testing.T) {
	assert.Nil(t, TypeSafetyCheck("test1", "http://google.com", "url"), "Test URL: google.com")
	assert.Nil(t, TypeSafetyCheck("test2", "http://technowax.net:80?foo=bar", "url"), "Test URL: technowax.net with query arguments")
	assert.Nil(t, TypeSafetyCheck("test3", "http://localhost:80?foo=bar", "url"), "Test URL: localhost with query arguments")
	assert.NotNil(t, TypeSafetyCheck("test4", "http://lo  host:80", "url"), "Test a badly formed URL")
	assert.NotNil(t, TypeSafetyCheck("test5", "12345", "url"), "Test a badly formed URL")
	assert.NotNil(t, TypeSafetyCheck("test6", "_!23;", "url"), "Test a badly formed URL")
}

func TestTypeSafetyCheckRegex(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		pattern  string
		value    string
		hasError bool
	}{
		{
			name:     "Issue #578 - Domain",
			field:    "domain",
			pattern:  "regex:^(?:[a-zA-Z0-9-]{1,63}.)+[a-zA-Z]{2,63}$",
			value:    "immich.example.dev",
			hasError: false,
		},
		{
			name:     "Don't allow numbers in username",
			field:    "Username",
			pattern:  "regex:^[a-zA-Z]$",
			value:    "James1234",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := typeSafetyCheckRegex(tt.field, tt.value, tt.pattern)

			if tt.hasError {
				assert.NotNil(t, err, "Expected error for value %s with pattern %s, but got no error", tt.value, tt.pattern)
			} else {
				assert.Nil(t, err, "Expected no error for value %s with pattern %s, but got error: %v", tt.value, tt.pattern, err)
			}
		})
	}
}

func TestRedactShellCommand(t *testing.T) {
	cmd := "echo 'The password for Fred is toomanysecrets'"

	args := []config.ActionArgument{
		{
			Name: "personName",
			Type: "ascii",
		},
		{
			Name: "password",
			Type: "password",
		},
	}

	values := map[string]string{
		"personName": "Fred",
		"password":   "toomanysecrets",
	}

	res := redactShellCommand(cmd, args, values)

	assert.Equal(t, "echo 'The password for Fred is <redacted>'", res, "Redacted shell command should mask the password argument")

	// Test with empty password
	values["password"] = ""
	res = redactShellCommand(cmd, args, values)
	assert.Equal(t, cmd, res, "Empty password should not change the command")

	// Test with missing password argument
	delete(values, "password")
	res = redactShellCommand(cmd, args, values)
	assert.Equal(t, cmd, res, "Missing password argument should not change the command")
}

func TestTypeSafetyCheckEmail(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		hasError bool
	}{
		{"Valid simple email", "email", "user@example.com", false},
		{"Valid email with subdomain", "email", "user@mail.example.com", false},
		{"Valid email with plus", "email", "user+test@example.com", false},
		{"Valid email with dash", "email", "user-name@example.com", false},
		{"Valid email with numbers", "email", "user123@example123.com", false},
		{"Invalid email no @", "email", "userexample.com", true},
		{"Invalid email no domain", "email", "user@", true},
		{"Invalid email no user", "email", "@example.com", true},
		{"Invalid email spaces", "email", "user name@example.com", true},
		{"Invalid email double @", "email", "user@@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TypeSafetyCheck(tt.field, tt.value, "email")
			if tt.hasError {
				assert.NotNil(t, err, "Expected error for value '%s'", tt.value)
			} else {
				assert.Nil(t, err, "Expected no error for value '%s', but got: %v", tt.value, err)
			}
		})
	}
}

func TestTypeSafetyCheckDatetime(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		hasError bool
	}{
		{"Valid datetime", "datetime", "2023-12-25T15:30:45", false},
		{"Valid datetime morning", "datetime", "2023-01-01T00:00:00", false},
		{"Valid datetime evening", "datetime", "2023-12-31T23:59:59", false},
		{"Invalid format missing T", "datetime", "2023-12-25 15:30:45", true},
		{"Invalid format missing seconds", "datetime", "2023-12-25T15:30", true},
		{"Invalid date", "datetime", "2023-13-25T15:30:45", true},
		{"Invalid time", "datetime", "2023-12-25T25:30:45", true},
		{"Random string", "datetime", "not-a-date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TypeSafetyCheck(tt.field, tt.value, "datetime")
			if tt.hasError {
				assert.NotNil(t, err, "Expected error for value '%s'", tt.value)
			} else {
				assert.Nil(t, err, "Expected no error for value '%s', but got: %v", tt.value, err)
			}
		})
	}
}

func TestTypeSafetyCheckRawStringMultiline(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value string
	}{
		{"Simple string", "content", "hello world"},
		{"Multiline string", "content", "line1\nline2\nline3"},
		{"String with special chars", "content", "!@#$%^&*()"},
		{"Unicode string", "content", "héllo wörld 🌍"},
		{"Very long string", "content", strings.Repeat("a", 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TypeSafetyCheck(tt.field, tt.value, "raw_string_multiline")
			assert.Nil(t, err, "raw_string_multiline should accept any value")
		})
	}
}

func TestTypeSafetyCheckUnicodeIdentifier(t *testing.T) {
	tests := []struct {
		name         string
		field        string
		value        string
		expectsError bool
	}{
		{"Valid unicode identifier", "name", "hello_world", false},
		{"Valid with numbers", "name", "test123", false},
		{"Valid with dots", "name", "file.txt", false},
		{"Valid with underscores", "name", "my_file_name", false},
		{"Invalid with special chars", "name", "hello@world", true},
		{"Invalid with brackets", "name", "hello[world]", true},
		{"Invalid with spaces", "name", "hello world", true},
		{"Invalid with path separators", "name", "path/to/file", true},
		{"Invalid with backslashes", "name", "path\\to\\file", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TypeSafetyCheck(tt.field, tt.value, "unicode_identifier")
			validateTypeSafetyResult(t, tt.value, tt.expectsError, err)
		})
	}
}

func validateTypeSafetyResult(t *testing.T, value string, expectsError bool, err error) {
	if expectsError {
		assertErrorExpected(t, value, err)
	} else {
		assertNoErrorExpected(t, value, err)
	}
}

func assertErrorExpected(t *testing.T, value string, err error) {
	if err == nil {
		t.Errorf("Expected error for value '%s', but got none", value)
	} else {
		t.Logf("Received expected error for value '%s': %v", value, err)
	}
}

func assertNoErrorExpected(t *testing.T, value string, err error) {
	if err != nil {
		t.Errorf("Expected no error for value '%s', but got: %v", value, err)
	} else {
		t.Logf("No error for valid value '%s' as expected", value)
	}
}

func TestTypeSafetyCheckAsciiIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		hasError bool
	}{
		{"Valid identifier", "name", "hello_world", false},
		{"Valid with numbers", "name", "test123", false},
		{"Valid with dots", "name", "file.txt", false},
		{"Valid with dashes", "name", "my-file", false},
		{"Valid with underscores", "name", "my_file", false},
		{"Invalid with spaces", "name", "hello world", true},
		{"Invalid with special chars", "name", "hello@world", true},
		{"Invalid unicode", "name", "héllo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TypeSafetyCheck(tt.field, tt.value, "ascii_identifier")
			if tt.hasError {
				assert.NotNil(t, err, "Expected error for value '%s'", tt.value)
			} else {
				assert.Nil(t, err, "Expected no error for value '%s', but got: %v", tt.value, err)
			}
		})
	}
}

func TestTypeSafetyCheckAsciiSentence(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		hasError bool
	}{
		{"Valid sentence", "text", "Hello world", false},
		{"Valid with numbers", "text", "Test 123", false},
		{"Valid with commas", "text", "Hello, world", false},
		{"Valid with periods", "text", "Hello world.", false},
		{"Valid with multiple spaces", "text", "Hello  world", false},
		{"Invalid with special chars", "text", "Hello@world", true},
		{"Invalid with parentheses", "text", "Hello (world)", true},
		{"Invalid unicode", "text", "Héllo world", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TypeSafetyCheck(tt.field, tt.value, "ascii_sentence")
			if tt.hasError {
				assert.NotNil(t, err, "Expected error for value '%s'", tt.value)
			} else {
				assert.Nil(t, err, "Expected no error for value '%s', but got: %v", tt.value, err)
			}
		})
	}
}

func TestTypecheckActionArgumentEmptyName(t *testing.T) {
	arg := config.ActionArgument{
		Name: "",
		Type: "ascii",
	}
	action := config.Action{Title: "Test"}

	err := typecheckActionArgument(&arg, "test", &action)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "argument name cannot be empty")
}

func TestTypecheckActionArgumentConfirmation(t *testing.T) {
	arg := config.ActionArgument{
		Name: "confirm",
		Type: "confirmation",
	}
	action := config.Action{Title: "Test"}

	err := typecheckActionArgument(&arg, "any_value", &action)
	assert.Nil(t, err, "Confirmation type should always pass validation")
}

func TestParseCommandForReplacements(t *testing.T) {
	tests := []struct {
		name           string
		shellCommand   string
		values         map[string]string
		expectedOutput string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "Simple replacement",
			shellCommand:   "echo {{ name }}",
			values:         map[string]string{"name": "John"},
			expectedOutput: "echo John",
			expectError:    false,
		},
		{
			name:           "Multiple replacements",
			shellCommand:   "echo {{ first }} {{ last }}",
			values:         map[string]string{"first": "John", "last": "Doe"},
			expectedOutput: "echo John Doe",
			expectError:    false,
		},
		{
			name:           "Replacement with spaces in template",
			shellCommand:   "echo {{  name  }}",
			values:         map[string]string{"name": "John"},
			expectedOutput: "echo John",
			expectError:    false,
		},
		{
			name:           "Missing argument",
			shellCommand:   "echo {{ missing }}",
			values:         map[string]string{},
			expectedOutput: "",
			expectError:    true,
			errorContains:  "required arg not provided: missing",
		},
		{
			name:           "No replacements needed",
			shellCommand:   "echo hello",
			values:         map[string]string{},
			expectedOutput: "echo hello",
			expectError:    false,
		},
		{
			name:           "Multiple same argument",
			shellCommand:   "echo {{ name }} says hello {{ name }}",
			values:         map[string]string{"name": "Alice"},
			expectedOutput: "echo Alice says hello Alice",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := parseCommandForReplacements(tt.shellCommand, tt.values, nil)

			if tt.expectError {
				assert.NotNil(t, err, "Expected error but got none")
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.Nil(t, err, "Expected no error but got: %v", err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestArgumentChoicesValidation(t *testing.T) {
	tests := []struct {
		name        string
		action      config.Action
		values      map[string]string
		expectError bool
		description string
	}{
		{
			name: "Valid choice",
			action: config.Action{
				Title: "Test choices",
				Shell: "echo {{ option }}",
				Arguments: []config.ActionArgument{
					{
						Name: "option",
						Type: "ascii",
						Choices: []config.ActionArgumentChoice{
							{Value: "option1", Title: "Option 1"},
							{Value: "option2", Title: "Option 2"},
						},
					},
				},
			},
			values:      map[string]string{"option": "option1"},
			expectError: false,
			description: "Should accept valid choice",
		},
		{
			name: "Invalid choice",
			action: config.Action{
				Title: "Test choices",
				Shell: "echo {{ option }}",
				Arguments: []config.ActionArgument{
					{
						Name: "option",
						Type: "ascii",
						Choices: []config.ActionArgumentChoice{
							{Value: "option1", Title: "Option 1"},
							{Value: "option2", Title: "Option 2"},
						},
					},
				},
			},
			values:      map[string]string{"option": "invalid_option"},
			expectError: true,
			description: "Should reject invalid choice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseActionArguments(tt.values, &tt.action, nil)

			if tt.expectError {
				assert.NotNil(t, err, tt.description)
				assert.Contains(t, err.Error(), "predefined choices")
			} else {
				assert.Nil(t, err, tt.description)
			}
		})
	}
}

func TestTypeSafetyCheckVeryDangerousRawString(t *testing.T) {
	// This type should allow anything without validation
	tests := []string{
		"normal text",
		"_zomg_ c:/ haxxor ' bobby tables && rm -rf /",
		"$(rm -rf /)",
		"; DROP TABLE users; --",
		"../../../../etc/passwd",
		"",
		"unicode: 你好世界",
		"emojis: 🔥💀☠️",
	}

	for _, value := range tests {
		t.Run(fmt.Sprintf("Value: %s", value), func(t *testing.T) {
			err := TypeSafetyCheck("test", value, "very_dangerous_raw_string")
			assert.Nil(t, err, "very_dangerous_raw_string should accept any value including: %s", value)
		})
	}
}

func TestParseActionArgumentsWithEntityPrefix(t *testing.T) {
	action := config.Action{
		Title: "Test entity prefix",
		Shell: "echo 'Processing {{ name }} for entity'",
		Arguments: []config.ActionArgument{
			{Name: "name", Type: "ascii"},
		},
	}

	values := map[string]string{
		"name": "testuser",
	}

	ent := &entities.Entity{
		Title: "entity_123",
	}

	// Test with entity prefix
	output, err := parseActionArguments(values, &action, ent)
	assert.Nil(t, err)
	assert.Contains(t, output, "testuser")
}

func TestComplexRegexPatterns(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		value    string
		hasError bool
	}{
		{
			name:     "Phone number pattern",
			pattern:  "regex:^\\+?[1-9]\\d{1,14}$",
			value:    "+1234567890",
			hasError: false,
		},
		{
			name:     "Invalid phone number",
			pattern:  "regex:^\\+?[1-9]\\d{1,14}$",
			value:    "123abc",
			hasError: true,
		},
		{
			name:     "Semantic version pattern",
			pattern:  "regex:^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)$",
			value:    "1.2.3",
			hasError: false,
		},
		{
			name:     "Invalid semantic version",
			pattern:  "regex:^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)$",
			value:    "1.2",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := typeSafetyCheckRegex("test", tt.value, tt.pattern)
			if tt.hasError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

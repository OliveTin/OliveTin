package logfilter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

const maxFilterLength = 512

var (
	comparePattern  = regexp.MustCompile(`(?i)\b(Status|Action|User|ExitCode|Blocked|TimedOut|Running)\s*(==|!=)\s*("[^"]*"|\S+)`)
	containsPattern = regexp.MustCompile(`(?i)\b(Status|Action|User|Output)\s+contains\s+("[^"]*"|\S+)`)
)

// Compile parses and compiles a filter expression. Returns an error for invalid syntax.
func Compile(expression string) (*vm.Program, error) {
	trimmed := strings.TrimSpace(expression)
	if trimmed == "" {
		return nil, nil
	}
	if len(trimmed) > maxFilterLength {
		return nil, fmt.Errorf("filter expression exceeds maximum length of %d characters", maxFilterLength)
	}

	normalized, err := normalizeExpression(trimmed)
	if err != nil {
		return nil, err
	}

	return compileNormalized(normalized)
}

func compileNormalized(normalized string) (*vm.Program, error) {
	return expr.Compile(normalized,
		expr.Env(Record{}),
		expr.AsBool(),
		expr.Function("includes", includes),
		expr.Function("hasTag", hasTag),
	)
}

func includes(params ...any) (any, error) {
	haystack, ok := params[0].(string)
	if !ok {
		return false, nil
	}
	needle, ok := params[1].(string)
	if !ok {
		return false, nil
	}
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle)), nil
}

func hasTag(params ...any) (any, error) {
	tags, ok := params[0].([]string)
	if !ok {
		return false, nil
	}
	needle, ok := params[1].(string)
	if !ok {
		return false, nil
	}
	return tagListIncludes(tags, needle), nil
}

func tagListIncludes(tags []string, needle string) bool {
	needle = strings.ToLower(needle)
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), needle) {
			return true
		}
	}
	return false
}

// Matches evaluates a compiled filter against a log record.
func Matches(program *vm.Program, record Record) (bool, error) {
	if program == nil {
		return true, nil
	}

	result, err := expr.Run(program, record)
	if err != nil {
		return false, err
	}

	matched, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("filter expression must return a boolean")
	}

	return matched, nil
}

func normalizeExpression(expression string) (string, error) {
	if isNegatedSearchTerm(expression) {
		term := quoteLiteral(strings.TrimPrefix(expression, "!"))
		return negatedSearchExpression(term), nil
	}

	if isPositiveSearchTerm(expression) {
		return positiveSearchExpression(quoteLiteral(expression)), nil
	}

	normalized := replaceContainsOperators(expression)
	normalized = replaceComparisons(normalized)
	return replaceBooleanWords(normalized), nil
}

func isNegatedSearchTerm(expression string) bool {
	if !strings.HasPrefix(expression, "!") {
		return false
	}
	remainder := strings.TrimSpace(expression[1:])
	return remainder != "" && !containsExpressionOperators(remainder)
}

func isPositiveSearchTerm(expression string) bool {
	return expression != "" && !containsExpressionOperators(expression)
}

func containsExpressionOperators(expression string) bool {
	lower := strings.ToLower(expression)
	operators := []string{"==", "!=", "&&", "||", " contains ", "(", ")"}
	for _, operator := range operators {
		if strings.Contains(lower, operator) {
			return true
		}
	}
	return false
}

func negatedSearchExpression(term string) string {
	return "!(" + positiveSearchExpression(term) + ")"
}

func positiveSearchExpression(term string) string {
	return "includes(Action, " + term + ") || includes(User, " + term + ") || includes(Status, " + term + ") || includes(Output, " + term + ") || hasTag(Tags, " + term + ")"
}

func replaceContainsOperators(expression string) string {
	return containsPattern.ReplaceAllStringFunc(expression, func(match string) string {
		parts := containsPattern.FindStringSubmatch(match)
		field := parts[1]
		value := quoteIfNeeded(parts[2])
		return fmt.Sprintf("includes(%s, %s)", field, value)
	})
}

func replaceComparisons(expression string) string {
	return comparePattern.ReplaceAllStringFunc(expression, func(match string) string {
		parts := comparePattern.FindStringSubmatch(match)
		field := parts[1]
		operator := parts[2]
		value := quoteIfNeeded(parts[3])
		return fmt.Sprintf("%s %s %s", field, operator, value)
	})
}

func replaceBooleanWords(expression string) string {
	replacer := strings.NewReplacer(" and ", " && ", " AND ", " && ", " or ", " || ", " OR ", " || ")
	return replacer.Replace(expression)
}

func quoteIfNeeded(value string) string {
	if strings.HasPrefix(value, "\"") {
		return value
	}
	return quoteLiteral(value)
}

func quoteLiteral(value string) string {
	return "\"" + strings.ReplaceAll(value, "\"", "\\\"") + "\""
}

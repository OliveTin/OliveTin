package logfilter

import (
	"fmt"

	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
)

// Filters only need field comparisons, boolean logic and the includes/hasTag
// helpers. Anything else (ranges, builtins such as count or map, closures,
// arithmetic) can make a short expression do a very large amount of work for
// every log entry, so it is rejected before compiling.

const maxFilterNodes = 256

var allowedUnaryOperators = map[string]bool{
	"!":   true,
	"not": true,
	"-":   true,
}

var allowedBinaryOperators = map[string]bool{
	"==":         true,
	"!=":         true,
	"<":          true,
	">":          true,
	"<=":         true,
	">=":         true,
	"&&":         true,
	"||":         true,
	"and":        true,
	"or":         true,
	"contains":   true,
	"startsWith": true,
	"endsWith":   true,
	"in":         true,
}

var allowedFunctions = map[string]bool{
	"includes": true,
	"hasTag":   true,
}

var errUnsupportedSyntax = fmt.Errorf("unsupported filter expression; use field comparisons, and/or/not, contains, includes() and hasTag()")

type allowlistVisitor struct {
	err error
}

func (v *allowlistVisitor) Visit(node *ast.Node) {
	if v.err != nil {
		return
	}
	v.err = checkNode(*node)
}

func checkAllowedSyntax(expression string) error {
	tree, err := parser.Parse(expression)
	if err != nil {
		return err
	}

	visitor := &allowlistVisitor{}
	ast.Walk(&tree.Node, visitor)
	return visitor.err
}

func checkNode(node ast.Node) error {
	if isLiteralOrIdentifier(node) {
		return nil
	}

	switch typed := node.(type) {
	case *ast.UnaryNode:
		return checkOperator(typed.Operator, allowedUnaryOperators)
	case *ast.BinaryNode:
		return checkOperator(typed.Operator, allowedBinaryOperators)
	case *ast.CallNode:
		return checkCall(typed)
	}

	return errUnsupportedSyntax
}

func isLiteralOrIdentifier(node ast.Node) bool {
	switch node.(type) {
	case *ast.NilNode, *ast.IdentifierNode, *ast.IntegerNode, *ast.FloatNode, *ast.BoolNode, *ast.StringNode, *ast.ArrayNode:
		return true
	}
	return false
}

func checkOperator(operator string, allowed map[string]bool) error {
	if !allowed[operator] {
		return fmt.Errorf("unsupported operator %q in filter expression", operator)
	}
	return nil
}

func checkCall(call *ast.CallNode) error {
	callee, ok := call.Callee.(*ast.IdentifierNode)
	if !ok || !allowedFunctions[callee.Value] {
		return errUnsupportedSyntax
	}
	return nil
}

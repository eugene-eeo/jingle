package testutils

import (
	"jingle/ast"
	"jingle/scanner"

	// "jingle/parser"
	"testing"
)

type ASTIdent struct{ Name string }
type ASTNil struct{}
type ASTNumber struct{ Value float64 }
type ASTString struct{ Value string }
type ASTBoolean struct{ Value bool }
type ASTArray []interface{}
type ASTAssign struct {
	Left  interface{}
	Right interface{}
}

func TestNode(t *testing.T, node ast.Node, v interface{}) bool {
	switch v := v.(type) {
	case ASTIdent:
		return TestIdentifierLiteral(t, node, v)
	case ASTNil:
		return TestNullLiteral(t, node)
	case ASTNumber:
		return TestNumberLiteral(t, node, v)
	case ASTString:
		return TestStringLiteral(t, node, v)
	case ASTBoolean:
		return TestBooleanLiteral(t, node, v)
	case ASTAssign:
		return TestAssignmentExpression(t, node, v)
	case ASTArray:
		return TestArrayLiteral(t, node, v)
	}
	panic("unhandled type")
}

// ==========
// Statements
// ==========

// ===========
// Expressions
// ===========

func TestInfixExpression(t *testing.T, node ast.Node, left interface{}, op string, right interface{}) bool {
	if !TestNodeType(t, node, ast.INFIX_EXPRESSION) {
		return false
	}
	infixExpr := node.(*ast.InfixExpression)
	if infixExpr.Op != op {
		t.Errorf("invalid node.Op. expected=%q, got=%q", op, infixExpr.Op)
		return false
	}
	return TestNode(t, infixExpr.Left, left) && TestNode(t, infixExpr.Right, right)
}

func TestOrExpression(t *testing.T, node ast.Node, left interface{}, right interface{}) bool {
	if !TestNodeType(t, node, ast.OR_EXPRESSION) {
		return false
	}
	expr := node.(*ast.OrExpression)
	if !testTokenType(t, expr.Token, scanner.TokenOr) {
		return false
	}
	return TestNode(t, expr.Left, left) && TestNode(t, expr.Right, right)
}

func TestAndExpression(t *testing.T, node ast.Node, left interface{}, right interface{}) bool {
	if !TestNodeType(t, node, ast.AND_EXPRESSION) {
		return false
	}
	expr := node.(*ast.AndExpression)
	if !testTokenType(t, expr.Token, scanner.TokenAnd) {
		return false
	}
	return TestNode(t, expr.Left, left) && TestNode(t, expr.Right, right)
}

func TestAssignmentExpression(t *testing.T, node ast.Node, v ASTAssign) bool {
	if !TestNodeType(t, node, ast.ASSIGNMENT_EXPRESSION) {
		return false
	}
	expr := node.(*ast.AssignmentExpression)
	if !testTokenType(t, expr.Token, scanner.TokenSet) {
		return false
	}
	return TestNode(t, expr.Left, v.Left) && TestNode(t, expr.Right, v.Right)
}

// ========
// Literals
// ========

func TestIdentifierLiteral(t *testing.T, node ast.Node, v ASTIdent) bool {
	if !TestNodeType(t, node, ast.IDENTIFIER_LITERAL) {
		return false
	}
	ident := node.(*ast.IdentifierLiteral)
	if !testTokenType(t, ident.Token, scanner.TokenIdent) {
		return false
	}
	if ident.Name() != v.Name {
		t.Errorf("invalid identifier name. expected=%q, got=%q", v.Name, ident.Name())
		return false
	}
	return true
}

func TestNullLiteral(t *testing.T, node ast.Node) bool {
	if !TestNodeType(t, node, ast.NIL_LITERAL) {
		return false
	}
	null := node.(*ast.NilLiteral)
	if !testTokenType(t, null.Token, scanner.TokenNil) {
		return false
	}
	return true
}

func TestNumberLiteral(t *testing.T, node ast.Node, v ASTNumber) bool {
	if !TestNodeType(t, node, ast.NUMBER_LITERAL) {
		return false
	}
	number := node.(*ast.NumberLiteral)
	if !testTokenType(t, number.Token, scanner.TokenNumber) {
		return false
	}
	if number.Value != v.Value {
		t.Errorf("invalid node.Value. expected=%f, got=%f", v, number.Value)
		return false
	}
	return true
}

func TestStringLiteral(t *testing.T, node ast.Node, v ASTString) bool {
	if !TestNodeType(t, node, ast.STRING_LITERAL) {
		return false
	}
	str := node.(*ast.StringLiteral)
	if !testTokenType(t, str.Token, scanner.TokenString) {
		return false
	}
	if str.Value != v.Value {
		t.Errorf("invalid node.Value. expected=%q, got=%q", v, str.Value)
		return false
	}
	return true
}

func TestBooleanLiteral(t *testing.T, node ast.Node, v ASTBoolean) bool {
	if !TestNodeType(t, node, ast.BOOLEAN_LITERAL) {
		return false
	}
	expr := node.(*ast.BooleanLiteral)
	if !testTokenType(t, expr.Token, scanner.TokenBoolean) {
		return false
	}
	if expr.Value != v.Value {
		t.Errorf("invalid node.Value. expected=%t, got=%t", v, expr.Value)
		return false
	}
	return true
}

func TestArrayLiteral(t *testing.T, node ast.Node, v ASTArray) bool {
	if !TestNodeType(t, node, ast.ARRAY_LITERAL) {
		return false
	}
	expr := node.(*ast.ArrayLiteral)
	if !testTokenType(t, expr.Token, scanner.TokenLBracket) {
		return false
	}
	if len(v) != len(expr.Elems) {
		t.Errorf("invalid no. of elements. expected=%d, got=%d",
			len(v), len(expr.Elems))
		return false
	}
	for i, elem := range expr.Elems {
		if !TestNode(t, elem, v[i]) {
			t.Errorf("elems[%d] expected=%+v, got=%+v",
				i, v[i], elem)
			return false
		}
	}
	return true
}

// ===============
// utils for utils
// ===============

func TestNodeType(t *testing.T, node ast.Node, nodeType ast.NodeType) bool {
	if node.Type() != nodeType {
		t.Errorf("invalid node.Type(). expected=%s, got=%s",
			nodeType,
			node.Type())
		return false
	}
	return true
}

func testTokenType(t *testing.T, token scanner.Token, tokenType scanner.TokenType) bool {
	if token.Type != tokenType {
		t.Errorf("invalid node.Start().Type. expected=%s, got=%s",
			tokenType,
			token.Type)
		return false
	}
	return true
}

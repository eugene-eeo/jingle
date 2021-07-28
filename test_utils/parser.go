package testutils

import (
	"jingle/ast"
	"jingle/token"

	// "jingle/parser"
	"testing"
)

type ASTIdent struct{ Name string }
type ASTNull struct{}
type ASTNumber struct{ Value float64 }
type ASTString struct{ Value string }

func TestNode(t *testing.T, node ast.Node, v interface{}) bool {
	switch v := v.(type) {
	case ASTIdent:
		return TestIdentifierLiteral(t, node, v)
	case ASTNull:
		return TestNullLiteral(t, node)
	case ASTNumber:
		return TestNumberLiteral(t, node, v)
	case ASTString:
		return TestStringLiteral(t, node, v)
	}
	panic("unhandled type")
}

// ==========
// Statements
// ==========

func TestLetStatement(t *testing.T, node ast.Node, left, right interface{}) bool {
	if !TestNodeType(t, node, ast.LET_STATEMENT) {
		return false
	}
	letStmt := node.(*ast.LetStatement)
	if !testTokenType(t, letStmt.Token, token.LET) {
		return false
	}
	return TestNode(t, letStmt.Left, left) && TestNode(t, letStmt.Right, right)
}

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
	if !testTokenType(t, expr.Token, token.OR) {
		return false
	}
	return TestNode(t, expr.Left, left) && TestNode(t, expr.Right, right)
}

func TestAndExpression(t *testing.T, node ast.Node, left interface{}, right interface{}) bool {
	if !TestNodeType(t, node, ast.AND_EXPRESSION) {
		return false
	}
	expr := node.(*ast.AndExpression)
	if !testTokenType(t, expr.Token, token.AND) {
		return false
	}
	return TestNode(t, expr.Left, left) && TestNode(t, expr.Right, right)
}

// ========
// Literals
// ========

func TestIdentifierLiteral(t *testing.T, node ast.Node, v ASTIdent) bool {
	if !TestNodeType(t, node, ast.IDENTIFIER_LITERAL) {
		return false
	}
	ident := node.(*ast.IdentifierLiteral)
	if !testTokenType(t, ident.Token, token.IDENT) {
		return false
	}
	if ident.Name() != v.Name {
		t.Errorf("invalid identifier name. expected=%q, got=%q", v.Name, ident.Name())
		return false
	}
	return true
}

func TestNullLiteral(t *testing.T, node ast.Node) bool {
	if !TestNodeType(t, node, ast.NULL_LITERAL) {
		return false
	}
	null := node.(*ast.NullLiteral)
	if !testTokenType(t, null.Token, token.NULL) {
		return false
	}
	return true
}

func TestNumberLiteral(t *testing.T, node ast.Node, v ASTNumber) bool {
	if !TestNodeType(t, node, ast.NUMBER_LITERAL) {
		return false
	}
	number := node.(*ast.NumberLiteral)
	if !testTokenType(t, number.Token, token.NUMBER) {
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
	if !testTokenType(t, str.Token, token.STRING) {
		return false
	}
	if str.Value != v.Value {
		t.Errorf("invalid node.Value. expected=%q, got=%q", v, str.Value)
		return false
	}
	return true
}

// ===============
// utils for utils
// ===============

func TestNodeType(t *testing.T, node ast.Node, nodeType ast.NodeType) bool {
	if node.Type() != nodeType {
		t.Errorf("invalid node.Type(). expected=%s, got=%s",
			ast.NodeTypeAsString(nodeType),
			ast.NodeTypeAsString(node.Type()))
		return false
	}
	return true
}

func testTokenType(t *testing.T, token token.Token, tokenType token.TokenType) bool {
	if token.Type != tokenType {
		t.Errorf("invalid node.Start().Type. expected=%s, got=%s",
			tokenType,
			token.Type)
		return false
	}
	return true
}

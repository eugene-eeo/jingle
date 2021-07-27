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
		return TestIdent(t, node, v)
	case ASTNull:
		return TestNullLiteral(t, node)
	case ASTNumber:
		return TestNumberLiteral(t, node, v)
	case ASTString:
		return TestStringLiteral(t, node, v)
	}
	panic("unhandled type")
}

func TestIdent(t *testing.T, node ast.Node, v ASTIdent) bool {
	if node.Type() != ast.IDENTIFIER_LITERAL {
		t.Errorf("invalid node.Type(). expected=%s, got=%s",
			ast.NodeTypeAsString(node.Type()),
			ast.NodeTypeAsString(ast.IDENTIFIER_LITERAL))
	}
	ident := node.(*ast.IdentifierLiteral)
	if !testLiteralToken(t, ident, token.IDENT) {
		return false
	}
	if ident.Name() != v.Name {
		t.Errorf("invalid identifier name. expected=%q, got=%q", v.Name, ident.Name())
		return false
	}
	return true
}

func TestNullLiteral(t *testing.T, node ast.Node) bool {
	if node.Type() != ast.NULL_LITERAL {
		t.Errorf("invalid node.Type(). expected=%s, got=%s",
			ast.NodeTypeAsString(node.Type()),
			ast.NodeTypeAsString(ast.NULL_LITERAL))
	}
	null := node.(*ast.NullLiteral)
	if !testLiteralToken(t, null, token.NULL) {
		return false
	}
	return true
}

func TestNumberLiteral(t *testing.T, node ast.Node, v ASTNumber) bool {
	if node.Type() != ast.NUMBER_LITERAL {
		t.Errorf("invalid node.Type(). expected=%s, got=%s",
			ast.NodeTypeAsString(node.Type()),
			ast.NodeTypeAsString(ast.NUMBER_LITERAL))
	}
	number := node.(*ast.NumberLiteral)
	if !testLiteralToken(t, number, token.NUMBER) {
		return false
	}
	if number.Value != v.Value {
		t.Errorf("invalid node.Value. expected=%f, got=%f", v, number.Value)
		return false
	}
	return true
}

func TestStringLiteral(t *testing.T, node ast.Node, v ASTString) bool {
	if node.Type() != ast.STRING_LITERAL {
		t.Errorf("invalid node.Type(). expected=%s, got=%s",
			ast.NodeTypeAsString(node.Type()),
			ast.NodeTypeAsString(ast.STRING_LITERAL))
	}
	str := node.(*ast.StringLiteral)
	if !testLiteralToken(t, str, token.STRING) {
		return false
	}
	if str.Value != v.Value {
		t.Errorf("invalid node.Value. expected=%q, got=%q", v, str.Value)
		return false
	}
	return true
}

func testLiteralToken(t *testing.T, node ast.Node, tokenType token.TokenType) bool {
	if node.Start().Type != tokenType {
		t.Errorf("invalid node.Start().Type. expected=%s, got=%s", tokenType, node.Start().Type)
		return false
	}
	if node.End().Type != tokenType {
		t.Errorf("invalid node.End().Type. expected=%s, got=%s", tokenType, node.Start().Type)
		return false
	}
	return true
}

func TestLetStatement(t *testing.T, node ast.Node, left, right interface{}) bool {
	if node.Type() != ast.LET_STATEMENT {
		t.Errorf("invalid node.Type(). expected=%s, got=%s",
			ast.NodeTypeAsString(node.Type()),
			ast.NodeTypeAsString(ast.LET_STATEMENT))
	}
	letStmt := node.(*ast.LetStatement)
	return TestNode(t, letStmt.Left, left) && TestNode(t, letStmt.Right, right)
}

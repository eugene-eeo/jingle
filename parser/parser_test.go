package parser_test

import (
	"jingle/ast"
	"jingle/parser"
	"jingle/scanner"
	ut "jingle/test_utils"
	"testing"
)

// ========================
// Statements
// ========================

// ========================
// Literals
// ========================

func TestParseLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"foobar'", ut.ASTIdent{"foobar'"}},
		{"nil", ut.ASTNil{}},
		{"100", ut.ASTNumber{100}},
		{"5.5", ut.ASTNumber{5.5}},
		{`"hello"`, ut.ASTString{"hello"}},
		{`true`, ut.ASTBoolean{true}},
		{`false`, ut.ASTBoolean{false}},
		{`[1,true,nil]`, ut.ASTArray{ut.ASTNumber{1}, ut.ASTBoolean{true}, ut.ASTNil{}}},
	}
	for i, tt := range tests {
		node, ok := checkParseOneline(t, tt.input)
		if !ok {
			t.Fatalf("test[%d] failed", i)
		}
		if !ut.TestNode(t, node, tt.expected) {
			t.Fatalf("test[%d] failed", i)
		}
	}
}

// ========================
// Expressions
// ========================

func TestParseInfixExpression(t *testing.T) {
	tests := []struct {
		input string
		left  interface{}
		op    string
		right interface{}
	}{
		{"1 + 1", ut.ASTNumber{1}, "+", ut.ASTNumber{1}},
		{"\"abc\" * nil", ut.ASTString{"abc"}, "*", ut.ASTNil{}},
	}
	for i, tt := range tests {
		node, ok := checkParseOneline(t, tt.input)
		if !ok {
			t.Fatalf("test[%d] failed", i)
		}
		if !ut.TestInfixExpression(t, node, tt.left, tt.op, tt.right) {
			t.Fatalf("test[%d] failed", i)
		}
	}
}

func TestParseShortCiruiting(t *testing.T) {
	tests := []struct {
		input string
		op    string
		left  interface{}
		right interface{}
	}{
		{"1 or 1", "or", ut.ASTNumber{1}, ut.ASTNumber{1}},
		{"\"abc\" or nil", "or", ut.ASTString{"abc"}, ut.ASTNil{}},
		{"def and nil", "and", ut.ASTIdent{"def"}, ut.ASTNil{}},
	}
	for i, tt := range tests {
		node, ok := checkParseOneline(t, tt.input)
		if !ok {
			t.Fatalf("test[%d] failed", i)
		}
		if tt.op == "or" {
			if !ut.TestOrExpression(t, node, tt.left, tt.right) {
				t.Fatalf("test[%d] failed", i)
			}
		} else {
			if !ut.TestAndExpression(t, node, tt.left, tt.right) {
				t.Fatalf("test[%d] failed", i)
			}
		}
	}
}

func TestParseAssignExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		test     ut.ASTAssign
	}{
		{"u = 1", "(u = 1)", ut.ASTAssign{ut.ASTIdent{"u"}, ut.ASTNumber{1}}},
		{"a = b = c", "(a = (b = c))", ut.ASTAssign{ut.ASTIdent{"a"}, ut.ASTAssign{ut.ASTIdent{"b"}, ut.ASTIdent{"c"}}}},
		{"[a=b] = [c]", "([(a = b)] = [c])", ut.ASTAssign{
			ut.ASTArray{ut.ASTAssign{ut.ASTIdent{"a"}, ut.ASTIdent{"b"}}},
			ut.ASTArray{ut.ASTIdent{"c"}},
		}},
	}
	for i, tt := range tests {
		node, ok := checkParseOneline(t, tt.input)
		if !ok {
			t.Fatalf("test[%d] failed", i)
		}
		if node.String() != tt.expected {
			t.Fatalf("test[%d] expected=%q, got=%q", i, tt.expected, node.String())
		}
		if !ut.TestAssignmentExpression(t, node, tt.test) {
			t.Fatalf("test[%d] failed", i)
		}
	}
}

func TestPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 1", "(1 + 1)"},
		{"1 + 2 + 3", "((1 + 2) + 3)"},
		{"1 * 2 + 3", "((1 * 2) + 3)"},
		{"1 + 2 * 3", "(1 + (2 * 3))"},
		{"1 / 2 * 3", "((1 / 2) * 3)"},
		{"a = b * c", "(a = (b * c))"},
		{"a * (b + c)", "(a * (b + c))"},
		{"a = b = c", "(a = (b = c))"},
		{"a.b.c", "((a).b).c"},
		{"d = a.b.c", "(d = ((a).b).c)"},
	}
	for i, tt := range tests {
		node, ok := checkParseOneline(t, tt.input)
		if !ok {
			t.Fatalf("test[%d] failed", i)
		}
		if node.String() != tt.expected {
			t.Fatalf("test[%d] expected=%q, got=%q", i, tt.expected, node.String())
		}
	}
}

// ====================
// Utils
// ====================

func checkParseOneline(t *testing.T, input string) (ast.Node, bool) {
	s := scanner.New("", input)
	s.ScanAll()
	if s.Errors() != nil {
		t.Errorf("cannot scan:\n\t%q", input)
		t.Errorf("scanner errors:\n")
		for _, e := range s.Errors() {
			t.Errorf("\t%s\n", e)
		}
		return nil, false
	}
	p := parser.New("", s.Tokens())
	program, err := p.Parse()
	if err != nil {
		t.Errorf("cannot parse:\n\t%q", input)
		t.Errorf("failed with error:\n\t%s", err)
		return nil, false
	}
	if len(program.Statements) != 1 {
		t.Errorf("expected len(program.Statements)=1, got=%d", len(program.Statements))
		return nil, false
	}
	return program.Statements[0], true
}

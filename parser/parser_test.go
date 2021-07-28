package parser_test

import (
	"jingle/lexer"
	"jingle/parser"
	ut "jingle/test_utils"
	"testing"
)

func TestParseLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"foobar'", ut.ASTIdent{"foobar'"}},
		{"null", ut.ASTNull{}},
		{"100", ut.ASTNumber{100}},
		{"5.5", ut.ASTNumber{5.5}},
		{`"hello"`, ut.ASTString{"hello"}},
	}
	for i, tt := range tests {
		p := parser.New(lexer.New(tt.input))
		program, err := p.Parse()
		if err != nil {
			t.Errorf("test[%d] failed", i)
			t.Fatalf("cannot parse: %q, got err=%s", tt.input, err)
		}
		if len(program.Nodes) != 1 {
			t.Fatalf("test[%d] expected len(program.Nodes)=1, got=%d", len(program.Nodes), i)
		}
		if !ut.TestNode(t, program.Nodes[0], tt.expected) {
			t.Fatalf("test[%d] failed", i)
		}
	}
}

func TestParseLetStatements(t *testing.T) {
	tests := []struct {
		input  string
		output string
		left   interface{}
		right  interface{}
	}{
		{"let a = b", "let a = b", ut.ASTIdent{"a"}, ut.ASTIdent{"b"}},
		{"let foo = bar", "let foo = bar", ut.ASTIdent{"foo"}, ut.ASTIdent{"bar"}},
		{"let foo = null", "let foo = null", ut.ASTIdent{"foo"}, ut.ASTNull{}},
	}
	for i, tt := range tests {
		p := parser.New(lexer.New(tt.input))
		program, err := p.Parse()
		if err != nil {
			t.Errorf("test[%d] failed", i)
			t.Fatalf("cannot parse: %q, got err=%s", tt.input, err)
		}
		if len(program.Nodes) != 1 {
			t.Fatalf("test[%d] expected len(program.Nodes)=1, got=%d", len(program.Nodes), i)
		}
		if !ut.TestLetStatement(t, program.Nodes[0], tt.left, tt.right) {
			t.Logf("%+v", program.Nodes[0])
			t.Fatalf("test[%d] failed", i)
		}
		if program.Nodes[0].String() != tt.output {
			t.Fatalf("test[%d] expected=%q, got=%q", i, tt.output, program.Nodes[0].String())
		}
	}
}

func TestParseInfixExpression(t *testing.T) {
	tests := []struct {
		input string
		left  interface{}
		op    string
		right interface{}
	}{
		{"1 + 1", ut.ASTNumber{1}, "+", ut.ASTNumber{1}},
		{"\"abc\" * null", ut.ASTString{"abc"}, "*", ut.ASTNull{}},
	}
	for i, tt := range tests {
		p := parser.New(lexer.New(tt.input))
		program, err := p.Parse()
		if err != nil {
			t.Errorf("test[%d] failed", i)
			t.Fatalf("cannot parse: %q, got err=%s", tt.input, err)
		}
		if len(program.Nodes) != 1 {
			t.Logf("%+v", program.Nodes)
			t.Fatalf("test[%d] expected len(program.Nodes)=1, got=%d", len(program.Nodes), i)
		}
		if !ut.TestInfixExpression(t, program.Nodes[0], tt.left, tt.op, tt.right) {
			t.Logf("%+v", program.Nodes[0])
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
	}
	for i, tt := range tests {
		p := parser.New(lexer.New(tt.input))
		program, err := p.Parse()
		if err != nil {
			t.Errorf("test[%d] failed", i)
			t.Fatalf("cannot parse: %q, got err=%s", tt.input, err)
		}
		if len(program.Nodes) != 1 {
			t.Logf("%+v", program.Nodes)
			t.Fatalf("test[%d] expected len(program.Nodes)=1, got=%d", len(program.Nodes), i)
		}
		expr := program.Nodes[0]
		if expr.String() != tt.expected {
			t.Fatalf("test[%d] expected=%q, got=%q", i, tt.expected, expr.String())
		}
	}
}

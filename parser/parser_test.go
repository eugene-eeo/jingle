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
			t.Fatalf("cannot parse: %q", tt.input)
		}
		if len(program.Nodes) != 1 {
			t.Fatalf("expected len(program.Nodes)=1, got=%d", len(program.Nodes))
		}
		if !ut.TestNode(t, program.Nodes[0], tt.expected) {
			t.Fatalf("failed")
		}
	}
}

func TestParseLetStatements(t *testing.T) {
	tests := []struct {
		input string
		left  interface{}
		right interface{}
	}{
		{"let a = b", ut.ASTIdent{"a"}, ut.ASTIdent{"b"}},
		{"let foo = bar", ut.ASTIdent{"foo"}, ut.ASTIdent{"bar"}},
		{"let foo = null", ut.ASTIdent{"foo"}, ut.ASTNull{}},
	}
	for i, tt := range tests {
		p := parser.New(lexer.New(tt.input))
		program, err := p.Parse()
		if err != nil {
			t.Errorf("test[%d] failed", i)
			t.Fatalf("cannot parse: %q", tt.input)
		}
		if len(program.Nodes) != 1 {
			t.Fatalf("expected len(program.Nodes)=1, got=%d", len(program.Nodes))
		}
		if !ut.TestLetStatement(t, program.Nodes[0], tt.left, tt.right) {
			t.Fatalf("failed")
		}
	}
}

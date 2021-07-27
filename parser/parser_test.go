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
			t.Fatalf("cannot parse: %q, got err=%s", tt.input, err)
		}
		if len(program.Nodes) != 1 {
			t.Fatalf("test[%d] expected len(program.Nodes)=1, got=%d", len(program.Nodes), i)
		}
		if !ut.TestLetStatement(t, program.Nodes[0], tt.left, tt.right) {
			t.Logf("%+v", program.Nodes[0])
			t.Fatalf("test[%d] failed", i)
		}
	}
}

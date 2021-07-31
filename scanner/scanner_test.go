package scanner_test

import (
	"jingle/scanner"
	"testing"
)

func TestScanner(t *testing.T) {
	s := scanner.New("", `a let foobar = 1.5
	ged.b = 12.; gab[f]=3


!===*/
fn() end
"abcdef""ghi""jkl\n\t\r\0"
nil == false`)
	for s.More() {
		s.Scan()
	}
	if s.Errors() != nil {
		t.Logf("unexpected parser errors:")
		for _, err := range s.Errors() {
			t.Logf("%s", err)
		}
		t.FailNow()
	}
	t.Logf("%v", s.Tokens())
	expected := []scanner.Token{
		{scanner.TokenIdent, "a", 1, 1},
		{scanner.TokenIdent, "let", 1, 3},
		{scanner.TokenIdent, "foobar", 1, 7},
		{scanner.TokenSet, "=", 1, 14},
		{scanner.TokenNumber, "1.5", 1, 16},
		{scanner.TokenSeparator, "\n\t", 1, 19},
		{scanner.TokenIdent, "ged", 2, 2},
		{scanner.TokenDot, ".", 2, 5},
		{scanner.TokenIdent, "b", 2, 6},
		{scanner.TokenSet, "=", 2, 8},
		{scanner.TokenNumber, "12.", 2, 10},
		{scanner.TokenSemicolon, ";", 2, 13},
		{scanner.TokenIdent, "gab", 2, 15},
		{scanner.TokenLBracket, "[", 2, 18},
		{scanner.TokenIdent, "f", 2, 19},
		{scanner.TokenRBracket, "]", 2, 20},
		{scanner.TokenSet, "=", 2, 21},
		{scanner.TokenNumber, "3", 2, 22},
		{scanner.TokenSeparator, "\n\n\n", 2, 23},
		{scanner.TokenNeq, "!=", 5, 1},
		{scanner.TokenEq, "==", 5, 3},
		{scanner.TokenMul, "*", 5, 5},
		{scanner.TokenDiv, "/", 5, 6},
		{scanner.TokenSeparator, "\n", 5, 7},
		{scanner.TokenFn, "fn", 6, 1},
		{scanner.TokenLParen, "(", 6, 3},
		{scanner.TokenRParen, ")", 6, 4},
		{scanner.TokenEnd, "end", 6, 6},
		{scanner.TokenSeparator, "\n", 6, 9},
		{scanner.TokenString, "abcdef", 7, 1},
		{scanner.TokenString, "ghi", 7, 9},
		{scanner.TokenString, "jkl\n\t\r\u0000", 7, 14},
		{scanner.TokenSeparator, "\n", 7, 27},
		{scanner.TokenNil, "nil", 8, 1},
		{scanner.TokenEq, "==", 8, 5},
		{scanner.TokenBoolean, "false", 8, 8},
		{scanner.TokenEOF, "", 8, 13},
	}
	tokens := s.Tokens()
	for i, tok := range expected {
		if tok != tokens[i] {
			t.Logf("expected=%s, got=%s", tok, tokens[i])
			t.Fatalf("failed at index=%d", i)
		}
	}
}

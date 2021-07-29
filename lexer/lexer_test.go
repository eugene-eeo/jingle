package lexer_test

import (
	"jingle/lexer"
	"jingle/token"
	"regexp"
	"testing"
)

func TestLexerNextToken(t *testing.T) {
	l := lexer.New(`
let foo
let a = foobar';

let f' = fn(x)
  return x + y * z == null
end

1 || 2
3 && 1000
"hello" || (!-3)
3.14159265
foo.bar != git
`)
	testTokens := []Token{
		{token.SEP, "\n", 1, 1},
		{token.LET, "let", 2, 1},
		{token.IDENT, "foo", 2, 5},
		{token.SEP, "\n", 2, 8},
		{token.LET, "let", 3, 1},
		{token.IDENT, "a", 3, 5},
		{token.ASSIGN, "=", 3, 7},
		{token.IDENT, "foobar'", 3, 9},
		{token.SEMICOLON, ";", 3, 16},
		{token.SEP, "\n\n", 3, 17},
		{token.LET, "let", 5, 1},
		{token.IDENT, "f'", 5, 5},
		{token.ASSIGN, "=", 5, 8},
		{token.FUNCTION, "fn", 5, 10},
		{token.LPAREN, "(", 5, 12},
		{token.IDENT, "x", 5, 13},
		{token.RPAREN, ")", 5, 14},
		{token.SEP, "\n", 5, 15},
		{token.RETURN, "return", 6, 3},
		{token.IDENT, "x", 6, 10},
		{token.PLUS, "+", 6, 12},
		{token.IDENT, "y", 6, 14},
		{token.ASTERISK, "*", 6, 16},
		{token.IDENT, "z", 6, 18},
		{token.EQ, "==", 6, 20},
		{token.NULL, "null", 6, 23},
		{token.SEP, "\n", 6, 27},
		{token.END, "end", 7, 1},
		{token.SEP, "\n\n", 7, 4},
		{token.NUMBER, "1", 9, 1},
		{token.OR, "||", 9, 3},
		{token.NUMBER, "2", 9, 6},
		{token.SEP, "\n", 9, 7},
		{token.NUMBER, "3", 10, 1},
		{token.AND, "&&", 10, 3},
		{token.NUMBER, "1000", 10, 6},
		{token.SEP, "\n", 10, 10},
		{token.STRING, "hello", 11, 1},
		{token.OR, "||", 11, 9},
		{token.LPAREN, "(", 11, 12},
		{token.BANG, "!", 11, 13},
		{token.MINUS, "-", 11, 14},
		{token.NUMBER, "3", 11, 15},
		{token.RPAREN, ")", 11, 16},
		{token.SEP, "\n", 11, 17},
		{token.NUMBER, "3.14159265", 12, 1},
		{token.SEP, "\n", 12, 11},
		{token.IDENT, "foo", 13, 1},
		{token.DOT, ".", 13, 4},
		{token.IDENT, "bar", 13, 5},
		{token.NOT_EQ, "!=", 13, 9},
		{token.IDENT, "git", 13, 12},
		{token.SEP, "\n", 13, 15},
		{token.EOF, "\u0000", 14, 1},
		{token.EOF, "\u0000", 14, 1},
		{token.EOF, "\u0000", 14, 1},
		{token.EOF, "\u0000", 14, 1},
	}
	for i, test := range testTokens {
		tok := l.NextToken()
		if l.Error() != nil {
			t.Fatalf("[%d] NextToken(): %s", i, l.Error())
			return
		}
		if !testToken(t, tok, test) {
			t.Errorf("[%d] test=%+v", i, test)
			t.Fatalf("[%d] testToken() failed", i)
			return
		}
	}
}

func TestLexIdentifier(t *testing.T) {
	tests := []struct {
		input string
		test  interface{}
	}{
		{"a", Token{token.IDENT, "a", 1, 1}},
		{"a1", Token{token.IDENT, "a1", 1, 1}},
		{"_1", Token{token.IDENT, "_1", 1, 1}},
		{"f'", Token{token.IDENT, "f'", 1, 1}},
		{"f'a", ParsingError{regexp.MustCompile(`:1:1: invalid IDENT: (.+)`)}},
	}
	for i, test := range tests {
		if !testOneToken(t, test.input, test.test) {
			t.Errorf("tests[%d]: failed", i)
		}
	}
}

func TestLexString(t *testing.T) {
	tests := []struct {
		input string
		test  interface{}
	}{
		{`"hello"`, Token{token.STRING, "hello", 1, 1}},
		{`  "hello\""`, Token{token.STRING, `hello"`, 1, 3}},
		{`  "hello\t\r\n\0"`, Token{token.STRING, "hello\t\r\n\u0000", 1, 3}},
		{` "hello\\"`, Token{token.STRING, `hello\`, 1, 2}},
		{"\"hello\n\"", ParsingError{regexp.MustCompile(`:1:7: invalid char in string literal: .+`)}},
	}
	for i, test := range tests {
		if !testOneToken(t, test.input, test.test) {
			t.Fatalf("tests[%d]: failed", i)
		}
	}
}

func TestLexerError(t *testing.T) {
	tests := []struct {
		input string
		regex string
	}{
		{"$", `^:1:1:`},
		{" a$", `^:1:2:`},
		{`    1$`, `^:1:5:`},
		{"\n1.s34\ndef", `^:2:1:`},
		{"abc\nq = \"abc\n\"", `^:2:9:`},
	}
	for i, test := range tests {
		if !testParserError(t, test.input, ParsingError{regexp.MustCompile(test.regex)}) {
			t.Errorf("tests[%d]: failed", i)
		}
	}
}

// =================
// Testing Utilities
// =================

type Token struct {
	expectedType    token.TokenType
	expectedLiteral interface{}
	expectedLineNo  interface{}
	expectedColumn  interface{}
}

type ParsingError struct {
	expectedMsg interface{}
}

func testToken(t *testing.T, tok token.Token, test Token) bool {
	if tok.Type != test.expectedType {
		t.Errorf("invalid tok.Type. expected=%s, got=%s",
			test.expectedType, tok.Type)
		return false
	}
	if test.expectedLiteral != nil {
		if tok.Literal != test.expectedLiteral.(string) {
			t.Logf("tok=%#v, test=%#v", tok, test)
			t.Errorf("invalid tok.Literal. expected=%q, got=%q",
				test.expectedLiteral, tok.Literal)
			return false
		}
	}
	if test.expectedLineNo != nil {
		if tok.LineNo != test.expectedLineNo.(int) {
			t.Logf("tok=%#v, test=%#v", tok, test)
			t.Errorf("invalid tok.LineNo. expected=%d, got=%d",
				test.expectedLineNo.(int), tok.LineNo)
			return false
		}
	}
	if test.expectedColumn != nil {
		if tok.Column != test.expectedColumn.(int) {
			t.Logf("tok=%#v, test=%#v", tok, test)
			t.Errorf("invalid tok.Column. expected=%d, got=%d",
				test.expectedColumn.(int), tok.Column)
			return false
		}
	}
	return true
}

func testParsingError(
	t *testing.T,
	err error,
	test ParsingError,
) bool {
	lexerErr, ok := err.(lexer.LexerError)
	if !ok {
		t.Errorf("Expected a lexer.LexerError. got=%T", err)
		return false
	}
	if lexerErr.Unwrap() != nil {
		t.Errorf("Expected lexerErr.Unwrap()=nil. got=%#v",
			lexerErr.Unwrap())
		return false
	}
	errMsg := lexerErr.Error()
	switch expectedMsg := test.expectedMsg.(type) {
	case *regexp.Regexp:
		if !expectedMsg.MatchString(errMsg) {
			t.Errorf("Unmatched lexerErr.Error(). got=%q",
				errMsg)
			return false
		}
	case string:
		if errMsg != expectedMsg {
			t.Errorf("Expected lexerErr.Error()=%q. got=%q",
				expectedMsg, errMsg)
			return false
		}
	default:
		t.Fatalf("unhandled type: %T", expectedMsg)
	}
	return true
}

// testParserError asserts that the given input eventually
// produces a parser error.
func testParserError(t *testing.T, input string, test ParsingError) bool {
	l := lexer.New(input)
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			t.Errorf("expected parser to return error")
			return false
		}
		err := l.Error()
		if err != nil {
			if !testParsingError(t, err, test) {
				return false
			}
			tok = l.NextToken()
			if tok.Type != token.ILLEGAL {
				t.Errorf("expected NextToken() after error to return ILLEGAL. got=%+v",
					tok)
				return false
			}
			if l.Error() != err {
				t.Errorf("invalid lexer.Error(). expected=%e, got=%e",
					err,
					l.Error())
				return false
			}
			return true
		}
	}
}

// testOneToken asserts that the given input produces exactly
// one token.
func testOneToken(
	t *testing.T,
	input string,
	test interface{},
) bool {
	l := lexer.New(input)
	tok := l.NextToken()
	err := l.Error()
	switch test := test.(type) {
	case ParsingError:
		if err == nil {
			t.Errorf("expected l.NextToken() to produce error, got=%#v",
				err)
			return false
		}
		if tok.Type != token.ILLEGAL {
			t.Errorf("expected token.Type == ILLEGAL "+
				"when l.NextToken() gives error. got=%s",
				tok.Type)
			return false
		}
		if !testParsingError(t, err, test) {
			return false
		}
		tok = l.NextToken()
		if tok.Type != token.ILLEGAL {
			t.Errorf("expected ILLEGAL token, got=%#v", tok)
			return false
		}
		if l.Error() != err {
			t.Errorf("expected l.NextToken() to return=%+v, got=%+v", err, l.Error())
			return false
		}
	case Token:
		if err != nil {
			t.Errorf("expected l.NextToken() to have no errors, got=%s", err)
			return false
		}
		if !testToken(t, tok, test) {
			return false
		}
		tok = l.NextToken()
		err = l.Error()
		if err != nil {
			t.Errorf("l.NextToken(): %s", err)
			return false
		}
		if tok.Type != token.EOF {
			t.Errorf("expected EOF, got=%#v", tok)
			return false
		}
	}
	return true
}

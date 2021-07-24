package lexer2_test

import (
	lexer "jingle/lexer2"
	"jingle/token"
	"regexp"
	"testing"
)

func TestLexer2NextToken(t *testing.T) {
	l := lexer.New(`let five = 5;
let ten = 10;

let add = fn(x, y) {
  x + y;
};

let result = add(five, ten);
!-/*5;
5 < 10 > 5;

if (5 < 10) {
  return true;
} else {
  return false;
}

10 == 10;
10 != 9;
null == null;
a || b;
a && b;
a >= b;
a <= b;
"hello world";
"\"hello world\t\r\n\\\0";
[1, 2];
{"foo": "bar"};
10 is 10;
`)
	testTokens := []Token{
		{token.LET, "let", 1, 1},
		{token.IDENT, "five", 1, 5},
		{token.ASSIGN, "=", 1, 10},
		{token.INT, "5", 1, 12},
		{token.SEMICOLON, ";", 1, 13},
		{token.LET, "let", 2, 1},
		{token.IDENT, "ten", 2, 5},
		{token.ASSIGN, "=", 2, 9},
		{token.INT, "10", 2, 11},
		{token.SEMICOLON, ";", 2, 13},
		{token.LET, "let", 4, 1},
		{token.IDENT, "add", 4, 5},
		{token.ASSIGN, "=", 4, 9},
		{token.FUNCTION, "fn", 4, 11},
		{token.LPAREN, "(", 4, 13},
		{token.IDENT, "x", 4, 14},
		{token.COMMA, ",", 4, 15},
		{token.IDENT, "y", 4, 17},
		{token.RPAREN, ")", 4, 18},
		{token.LBRACE, "{", 4, 20},
		{token.IDENT, "x", 5, 3},
		{token.PLUS, "+", 5, 5},
		{token.IDENT, "y", 5, 7},
		{token.SEMICOLON, ";", 5, 8},
		{token.RBRACE, "}", 6, 1},
		{token.SEMICOLON, ";", 6, 2},
		{token.LET, "let", 8, 1},
		{token.IDENT, "result", 8, 5},
		{token.ASSIGN, "=", 8, 12},
		{token.IDENT, "add", 8, 14},
		{token.LPAREN, "(", 8, 17},
		{token.IDENT, "five", 8, 18},
		{token.COMMA, ",", 8, 22},
		{token.IDENT, "ten", 8, 24},
		{token.RPAREN, ")", 8, 27},
		{token.SEMICOLON, ";", 8, 28},
		{token.BANG, "!", 9, 1},
		{token.MINUS, "-", 9, 2},
		{token.SLASH, "/", 9, 3},
		{token.ASTERISK, "*", 9, 4},
		{token.INT, "5", 9, 5},
		{token.SEMICOLON, ";", 9, 6},
		{token.INT, "5", 10, 1},
		{token.LT, "<", 10, 3},
		{token.INT, "10", 10, 5},
		{token.GT, ">", 10, 8},
		{token.INT, "5", 10, 10},
		{token.SEMICOLON, ";", 10, 11},
		{token.IF, "if", 12, 1},
		{token.LPAREN, "(", 12, 4},
		{token.INT, "5", 12, 5},
		{token.LT, "<", 12, 7},
		{token.INT, "10", 12, 9},
		{token.RPAREN, ")", 12, 11},
		{token.LBRACE, "{", 12, 13},
		{token.RETURN, "return", 13, 3},
		{token.TRUE, "true", 13, 10},
		{token.SEMICOLON, ";", 13, 14},
		{token.RBRACE, "}", 14, 1},
		{token.ELSE, "else", 14, 3},
		{token.LBRACE, "{", 14, 8},
		{token.RETURN, "return", 15, 3},
		{token.FALSE, "false", 15, 10},
		{token.SEMICOLON, ";", 15, 15},
		{token.RBRACE, "}", 16, 1},
		{token.INT, "10", 18, 1},
		{token.EQ, "==", 18, 4},
		{token.INT, "10", 18, 7},
		{token.SEMICOLON, ";", 18, 9},
		{token.INT, "10", 19, 1},
		{token.NOT_EQ, "!=", 19, 4},
		{token.INT, "9", 19, 7},
		{token.SEMICOLON, ";", 19, 8},
		{token.NULL, "null", 20, 1},
		{token.EQ, "==", 20, 6},
		{token.NULL, "null", 20, 9},
		{token.SEMICOLON, ";", 20, 13},
		{token.IDENT, "a", 21, 1},
		{token.OR, "||", 21, 3},
		{token.IDENT, "b", 21, 6},
		{token.SEMICOLON, ";", 21, 7},
		{token.IDENT, "a", 22, 1},
		{token.AND, "&&", 22, 3},
		{token.IDENT, "b", 22, 6},
		{token.SEMICOLON, ";", 22, 7},
		{token.IDENT, "a", 23, 1},
		{token.GEQ, ">=", 23, 3},
		{token.IDENT, "b", 23, 6},
		{token.SEMICOLON, ";", 23, 7},
		{token.IDENT, "a", 24, 1},
		{token.LEQ, "<=", 24, 3},
		{token.IDENT, "b", 24, 6},
		{token.SEMICOLON, ";", 24, 7},
		{token.STRING, "hello world", 25, 1},
		{token.SEMICOLON, ";", 25, 14},
		{token.STRING, "\"hello world\t\r\n\\\u0000", 26, 1},
		{token.SEMICOLON, ";", 26, 26},
		{token.LBRACKET, "[", 27, 1},
		{token.INT, "1", 27, 2},
		{token.COMMA, ",", 27, 3},
		{token.INT, "2", 27, 5},
		{token.RBRACKET, "]", 27, 6},
		{token.SEMICOLON, ";", 27, 7},
		{token.LBRACE, "{", 28, 1},
		{token.STRING, "foo", 28, 2},
		{token.COLON, ":", 28, 7},
		{token.STRING, "bar", 28, 9},
		{token.RBRACE, "}", 28, 14},
		{token.SEMICOLON, ";", 28, 15},
		{token.INT, "10", 29, 1},
		{token.IS, "is", 29, 4},
		{token.INT, "10", 29, 7},
		{token.SEMICOLON, ";", 29, 9},
		{token.EOF, "\u0000", 30, 1},
	}
	for i, test := range testTokens {
		tok, err := l.NextToken()
		if err != nil {
			t.Fatalf("[%d] NextToken(): %s", i, err)
			return
		}
		if !testToken(t, tok, test) {
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
		{"\"hello\n\"", ParsingError{regexp.MustCompile(`:2:0: invalid char in string literal: .+`)}},
	}
	for i, test := range tests {
		if !testOneToken(t, test.input, test.test) {
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
		}
	case string:
		if errMsg != expectedMsg {
			t.Errorf("Expected lexerErr.Error()=%q. got=%q",
				expectedMsg, errMsg)
		}
	default:
		t.Fatalf("unhandled type: %T", expectedMsg)
	}
	return true
}

// testOneToken asserts that the given input produces exactly
// one token.
func testOneToken(
	t *testing.T,
	input string,
	test interface{},
) bool {
	l := lexer.New(input)
	tok, err := l.NextToken()
	switch test := test.(type) {
	case ParsingError:
		if err == nil {
			t.Errorf("expected l.NextToken() to produce error, got=%#v", err)
			return false
		}
		if !testParsingError(t, err, test) {
			return false
		}
	case Token:
		if err != nil {
			t.Errorf("l.NextToken(): %s", err)
			return false
		}
		if !testToken(t, tok, test) {
			return false
		}
	}
	tok, err = l.NextToken()
	if err != nil {
		t.Errorf("l.NextToken(): %s", err)
		return false
	}
	if tok.Type != token.EOF {
		t.Errorf("expected EOF, got=%#v", tok)
		return false
	}
	return true
}

package lexer_test

import (
	"jingle/lexer"
	"jingle/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `let five = 5;
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
`

	tests := []Token{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.NULL, "null"},
		{token.EQ, "=="},
		{token.NULL, "null"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "a"},
		{token.OR, "||"},
		{token.IDENT, "b"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "a"},
		{token.AND, "&&"},
		{token.IDENT, "b"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "a"},
		{token.GEQ, ">="},
		{token.IDENT, "b"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "a"},
		{token.LEQ, "<="},
		{token.IDENT, "b"},
		{token.SEMICOLON, ";"},
		{token.STRING, "hello world"},
		{token.SEMICOLON, ";"},
		{token.STRING, "\"hello world\t\r\n\\\u0000"},
		{token.SEMICOLON, ";"},
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.IS, "is"},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := lexer.New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if !testToken(t, tok, tt) {
			t.Fatalf("tests[%d]: failed", i)
			return
		}
	}
}

func TestLexIdentifier(t *testing.T) {
	tests := []struct {
		input     string
		testToken Token
	}{
		{"a", Token{token.IDENT, "a"}},
		{"a1", Token{token.IDENT, "a1"}},
		{"_1", Token{token.IDENT, "_1"}},
		{"f'", Token{token.IDENT, "f'"}},
		{"f'a", Token{token.ILLEGAL, "invalid IDENT: \"f'a\""}},
	}
	for i, test := range tests {
		if !testOneToken(t, test.input, test.testToken) {
			t.Errorf("tests[%d]: failed", i)
		}
	}
}

func TestLexNumber(t *testing.T) {
	tests := []struct {
		input     string
		testToken Token
	}{
		{"1", Token{token.INT, "1"}},
		{"1a", Token{token.ILLEGAL, nil}},
		{"1b", Token{token.ILLEGAL, nil}},
		{"100", Token{token.INT, "100"}},
		{"1239421", Token{token.INT, "1239421"}},
	}
	for i, test := range tests {
		if !testOneToken(t, test.input, test.testToken) {
			t.Errorf("tests[%d]: failed", i)
		}
	}
}

func TestLexString(t *testing.T) {
	tests := []struct {
		input     string
		testToken Token
	}{
		{`"hello"`, Token{token.STRING, "hello"}},
		{`"hello\""`, Token{token.STRING, `hello"`}},
		{`"hello\t\r\n\0"`, Token{token.STRING, "hello\t\r\n\u0000"}},
		{`"hello\\"`, Token{token.STRING, `hello\`}},
		{"\"hello\n\"", Token{token.ILLEGAL, nil}},
	}
	for i, test := range tests {
		if !testOneToken(t, test.input, test.testToken) {
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
}

func testToken(t *testing.T, tok token.Token, test Token) bool {
	if tok.Type != test.expectedType {
		t.Errorf("invalid tok.Type. expected=%s, got=%s",
			test.expectedType, tok.Type)
		return false
	}
	if test.expectedLiteral != nil {
		if tok.Literal != test.expectedLiteral.(string) {
			t.Errorf("invalid tok.Literal. expected=%q, got=%q",
				test.expectedLiteral, tok.Literal)
			return false
		}
	}
	return true
}

// testOneToken asserts that the given input produces exactly
// one token.
func testOneToken(
	t *testing.T,
	input string,
	test Token,
) bool {
	l := lexer.New(input)
	tok := l.NextToken()
	if !testToken(t, tok, test) {
		return false
	}
	if tok := l.NextToken(); tok.Type != token.EOF {
		t.Errorf("expected EOF, got=%T(%#v)", tok, tok)
		return false
	}
	return true
}

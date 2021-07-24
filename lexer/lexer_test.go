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

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
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
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d]: tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
			// t.Fatalf("tests[%d]: %T(%+v)", i, tok, tok)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d]: literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexIdentifier(t *testing.T) {
	type testToken struct {
		_type token.TokenType
		lit   string
	}
	tests := []struct {
		input  string
		tokens []testToken
	}{
		{"a", []testToken{{token.IDENT, "a"}}},
		{"a1", []testToken{{token.IDENT, "a1"}}},
		{"_1", []testToken{{token.IDENT, "_1"}}},
		{"f'", []testToken{{token.IDENT, "f'"}}},
		{"'f", []testToken{{token.ILLEGAL, "'"}, {token.IDENT, "f"}}},
	}
	for i, test := range tests {
		l := lexer.New(test.input)
		for j, tt := range test.tokens {
			tok := l.NextToken()
			if tok.Type != tt._type {
				t.Fatalf("tests[%d][%d]: tokentype wrong. expected=%q, got=%q",
					i, j, tt._type, tok.Type)
			}
			if tok.Literal != tt.lit {
				t.Fatalf("tests[%d]: literal wrong. expected=%q, got=%q",
					i, tt.lit, tok.Literal)
			}
		}
		tok := l.NextToken()
		if tok.Type != token.EOF {
			t.Fatalf("tests[%d]: tokentype wrong. expected=%q, got=%q",
				i, token.EOF, tok.Type)
		}
	}
}

func TestLexString(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`"hello"`, "hello"},
		{`"hello\""`, `hello"`},
		{`"hello\t\r\n\0"`, "hello\t\r\n\u0000"},
		{`"hello\\"`, `hello\`},
		{"\"hello\n\"", nil},
	}

	for i, tt := range tests {
		l := lexer.New(tt.input)
		tok := l.NextToken()
		if tt.expected == nil {
			if tok.Type != token.ILLEGAL {
				t.Errorf("tests[%d]: tok.Type != token.ILLEGAL. got=%q", i, tok.Type)
				continue
			}
		} else {
			if tok.Type != token.STRING {
				t.Errorf("tests[%d]: tok.Type != token.STRING. got=%q", i, tok.Type)
				continue
			}
			if tok.Literal != tt.expected.(string) {
				t.Errorf("tests[%d]: tok.Literal != %q. got=%q", i, tt.expected, tok.Literal)
			}
			eof := l.NextToken()
			if eof.Type != token.EOF {
				t.Errorf("tests[%d]: eof.Type != token.EOF. got=%q", i, tok.Type)
				continue
			}
		}
	}
}

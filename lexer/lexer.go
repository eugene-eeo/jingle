package lexer

import (
	"bytes"
	"fmt"
	"jingle/token"
	"regexp"
)

type LexerError struct {
	error string
}

func (le LexerError) Error() string { return le.error }

var (
	identRegex   = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*[']*$`)
	integerRegex = regexp.MustCompile(`^[0-9]+$`)
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination. byte == input[position]
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		// EOF ==> set ch to NUL.
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) skipWhitespace() {
	for isWhiteSpace(l.ch) {
		l.readChar()
	}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '"':
		return l.scanString()
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case ':':
		tok = newToken(token.COLON, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.OR, Literal: literal}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.AND, Literal: literal}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.LEQ, Literal: literal}
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.GEQ, Literal: literal}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok = l.scanAtom(token.IDENT, identRegex)
			if tok.Type != token.ILLEGAL {
				tok.Type = token.LookupIdent(tok.Literal)
			}
			return tok
		} else if isDigit(l.ch) {
			tok = l.scanAtom(token.INT, integerRegex)
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) scanString() token.Token {
	escape := false
	var buf bytes.Buffer
outer:
	for {
		l.readChar() // skip over the " initially
		if escape {
			switch l.ch {
			case '\\':
				buf.WriteByte('\\')
			case 'n':
				buf.WriteByte('\n')
			case 't':
				buf.WriteByte('\t')
			case 'r':
				buf.WriteByte('\r')
			case '0':
				buf.WriteByte(0)
			case '"':
				buf.WriteByte('"')
			default:
				break outer
			}
			// remember to turn it off!
			escape = false
		} else {
			switch l.ch {
			case '\\':
				escape = true
			case '"':
				l.readChar() // consume '"'
				tok := token.Token{Type: token.STRING}
				tok.Literal = buf.String()
				return tok
			case '\n':
				break outer
			case '\r':
				break outer
			case 0:
				break outer
			default:
				buf.WriteByte(l.ch)
			}
		}
	}
	tok := token.Token{Type: token.ILLEGAL}
	tok.Literal = fmt.Sprintf("invalid character: %q", l.ch)
	for l.ch != 0 && l.ch != '"' {
		// try to consume until the next "
		l.readChar()
	}
	l.readChar() // consume the "
	return tok
}

func (l *Lexer) scanAtom(
	tokenType token.TokenType,
	re *regexp.Regexp,
) token.Token {
	position := l.position
	for !isOperator(l.ch) && !isWhiteSpace(l.ch) && l.ch != 0 {
		l.readChar()
	}
	tok := token.Token{}
	str := l.input[position:l.position]
	if indexes := re.FindIndex([]byte(str)); indexes == nil ||
		indexes[0] != 0 ||
		indexes[1] != len(str) {
		tok.Type = token.ILLEGAL
		tok.Literal = fmt.Sprintf("invalid %s: %q", tokenType, str)
		return tok
	}
	tok.Type = tokenType
	tok.Literal = str
	return tok
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isOperator(ch byte) bool {
	return ch == '=' ||
		ch == '+' ||
		ch == '-' ||
		ch == '*' ||
		ch == '/' ||
		ch == ':' ||
		ch == '|' ||
		ch == '&' ||
		ch == '<' ||
		ch == '>' ||
		ch == '!' ||
		ch == ',' ||
		ch == ';' ||
		ch == '(' ||
		ch == ')' ||
		ch == '{' ||
		ch == '}' ||
		ch == '[' ||
		ch == ']' ||
		ch == '"'
}

func isWhiteSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

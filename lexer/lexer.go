package lexer

import (
	"bytes"
	"fmt"
	"io"
	"jingle/token"
	"regexp"
)

var ErrorToken = token.Token{Type: token.ILLEGAL}

// atom regexes
var (
	identRegex  = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*[']*$`)
	numberRegex = regexp.MustCompile(`^[0-9]+(\.[0-9]*)?$`)
)

// atom allowedchars
var (
	identAllowed  = func(ch rune) bool { return !isPunctuation(ch) }
	numberAllowed = func(ch rune) bool { return ch == '.' || !isPunctuation(ch) }
)

type Lexer struct {
	Filename string
	input    *peeker
	ch       rune  // current rune under examination
	err      error // have we met an error?
}

func NewFromReader(filename string, r io.RuneReader) *Lexer {
	l := &Lexer{}
	l.input = newPeeker(r, 4)
	l.Filename = ""
	l.advance()
	return l
}

func New(s string) *Lexer {
	return NewFromReader("", bytes.NewReader([]byte(s)))
}

func (l *Lexer) Error() error {
	if l.err == io.EOF {
		return nil
	}
	return l.err
}

func (l *Lexer) advance() {
	if l.err != nil {
		return
	}
	r, err := l.input.Next()
	l.ch = r
	if err != nil {
		if err != io.EOF {
			err = l.wrapError(err)
		}
		l.err = err
		l.ch = 0
		return
	}
}

// -----------------------------
// The meat of the code is here!
// -----------------------------

func (l *Lexer) NextToken() token.Token {
	if l.err != nil {
		if l.err == io.EOF {
			return l.emitRuneToken(token.EOF, 0)
		}
		return ErrorToken
	}

	// Skip over whitespace
	for isWhiteSpace(l.ch) {
		l.advance()
		if l.err != nil {
			return ErrorToken
		}
	}

	tok := ErrorToken
	switch l.ch {
	case '.':
		tok = l.emitRuneToken(token.DOT, l.ch)
	case ',':
		tok = l.emitRuneToken(token.COMMA, l.ch)
	case ';':
		tok = l.emitRuneToken(token.SEMICOLON, l.ch)
	case '+':
		tok = l.emitRuneToken(token.PLUS, l.ch)
	case '-':
		tok = l.emitRuneToken(token.MINUS, l.ch)
	case '*':
		tok = l.emitRuneToken(token.ASTERISK, l.ch)
	case '/':
		tok = l.emitRuneToken(token.SLASH, l.ch)
	case ':':
		tok = l.emitRuneToken(token.COLON, l.ch)
	case '|':
		if l.peek(0) == '|' {
			tok = l.emitStringToken(token.OR, "||")
			l.advance() // peek != NUL ==> advance has no error.
		}
	case '&':
		if l.peek(0) == '&' {
			tok = l.emitStringToken(token.AND, "&&")
			l.advance()
		}
	case '<':
		if l.peek(0) == '=' {
			tok = l.emitStringToken(token.LEQ, "<=")
			l.advance()
		} else {
			tok = l.emitRuneToken(token.LT, l.ch)
		}
	case '>':
		if l.peek(0) == '=' {
			tok = l.emitStringToken(token.GEQ, ">=")
			l.advance()
		} else {
			tok = l.emitRuneToken(token.GT, l.ch)
		}
	case '=':
		if l.peek(0) == '=' {
			tok = l.emitStringToken(token.EQ, "==")
			l.advance()
		} else {
			tok = l.emitRuneToken(token.ASSIGN, l.ch)
		}
	case '!':
		if l.peek(0) == '=' {
			tok = l.emitStringToken(token.NOT_EQ, "!=")
			l.advance()
		} else {
			tok = l.emitRuneToken(token.BANG, l.ch)
		}
	case '(':
		tok = l.emitRuneToken(token.LPAREN, l.ch)
	case ')':
		tok = l.emitRuneToken(token.RPAREN, l.ch)
	case '{':
		tok = l.emitRuneToken(token.LBRACE, l.ch)
	case '}':
		tok = l.emitRuneToken(token.RBRACE, l.ch)
	case '[':
		tok = l.emitRuneToken(token.LBRACKET, l.ch)
	case ']':
		tok = l.emitRuneToken(token.RBRACKET, l.ch)
	case '"':
		return l.scanString()
	default:
		if isSeparator(l.ch) {
			return l.scanSeparators()
		} else if isDigit(l.ch) {
			return l.scanAtom(token.NUMBER, numberRegex, numberAllowed)
		} else if isLetter(l.ch) {
			tok := l.scanAtom(token.IDENT, identRegex, identAllowed)
			if tok.Type != token.ILLEGAL {
				tok.Type = token.LookupIdent(tok.Literal)
			}
			return tok
		} else {
			l.err = l.makeError("unhandled character: %U %q", l.ch, l.ch)
			return ErrorToken
		}
	}
	l.advance()
	return tok
}

// ========
// Utils...
// ========

func (l *Lexer) scanSeparators() token.Token {
	var buf bytes.Buffer
	tok := token.Token{Type: token.SEP, LineNo: l.input.lineNo, Column: l.input.column}
	for isSeparator(l.ch) {
		// consume one, and then continue eating separators
		buf.WriteRune(l.ch)
		l.advance()
	}
	tok.Literal = buf.String()
	return tok
}

func (l *Lexer) scanString() token.Token {
	escape := false
	var buf bytes.Buffer
	tok := token.Token{Type: token.STRING, LineNo: l.input.lineNo, Column: l.input.column}
	// we are on top of the '"' character -- skip over it.
outer:
	for {
		l.advance()
		if l.err != nil {
			return ErrorToken
		}
		if escape {
			switch l.ch {
			case '\\':
				buf.WriteRune('\\')
			case 'n':
				buf.WriteRune('\n')
			case 't':
				buf.WriteRune('\t')
			case 'r':
				buf.WriteRune('\r')
			case '0':
				buf.WriteRune(0)
			case '"':
				buf.WriteRune(l.ch)
			default:
				break outer
			}
			escape = false
		} else {
			switch l.ch {
			case '\\':
				escape = true
			case '"':
				l.advance() // consume '"'
				tok.Literal = buf.String()
				return tok
			case '\n':
				fallthrough
			case '\r':
				break outer
			default:
				buf.WriteRune(l.ch)
			}
		}
	}
	l.err = l.makeError("invalid char in string literal: %q", string(l.ch))
	return ErrorToken
}

func (l *Lexer) scanAtom(t token.TokenType, re *regexp.Regexp, allowed func(rune) bool) token.Token {
	// save these, because we will ruin them in a second.
	startingLineNo := l.input.lineNo
	startingColumn := l.input.column
	// begin scanning!
	var buf bytes.Buffer
	for !(l.ch == 0 || isWhiteSpace(l.ch) || isSeparator(l.ch)) && allowed(l.ch) {
		buf.WriteRune(l.ch)
		l.advance()
		// if l.err != nil {
		// 	return ErrorToken
		// }
	}
	// try to match the atom
	str := buf.String()
	if !re.MatchString(str) {
		err := newErrorFromLexer(l)
		err.message = fmt.Sprintf("invalid %s: %q", t, str)
		err.LineNo = startingLineNo
		err.Column = startingColumn
		l.err = err
		return ErrorToken
	}
	return token.Token{
		Type:    t,
		Literal: str,
		LineNo:  startingLineNo,
		Column:  startingColumn,
	}
}

func (l *Lexer) peek(n uint8) rune {
	// s := ""
	// for i := 1; i <= n; i++ {
	// 	r, err := l.input.Peek(i)
	// }
	// return s
	p, err := l.input.Peek(n)
	if err != nil {
		return 0
	}
	return p
}

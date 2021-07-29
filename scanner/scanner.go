// package scanner implements a scanner for jingle.
package scanner

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Token represents a token returned from the scanner.
type Token struct {
	Type   TokenType
	Value  string
	LineNo int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("%d(%d:%d:%q)", t.Type, t.LineNo, t.Column, t.Value)
}

type TokenType int

type Error struct {
	Filename string
	Message  string
	Value    string
	LineNo   int
	Column   int
}

func (e Error) String() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Filename, e.LineNo, e.Column, e.Message)
}

const (
	TokenEOF = TokenType(iota) // signals that EOF is reached.
	// Keywords
	TokenLet   // 'let'
	TokenOr    // 'or'
	TokenAnd   // 'and'
	TokenFn    // 'fn'
	TokenEnd   // 'end'
	TokenFor   // 'for'
	TokenWhile // 'while'
	// Literals
	TokenNil     // nil
	TokenBoolean // true or false
	TokenString  // a string literal
	TokenNumber  // floating point number (yuck)
	TokenIdent   // identifier
	// Delimiters
	TokenSemicolon // ';'
	TokenSeparator // any newlines (we are whitespace sensitive)
	TokenLParen    // '('
	TokenRParen    // ')'
	TokenLBrace    // '{'
	TokenRBrace    // '}'
	TokenLBracket  // '['
	TokenRBracket  // ']'
	// Operators
	TokenBang  // '!'
	TokenDot   // '.'
	TokenPlus  // '+'
	TokenMinus // '-'
	TokenMul   // '*'
	TokenDiv   // '/'
	TokenSet   // '='
	TokenEq    // '=='
	TokenNeq   // '!='
	TokenLt    // '<'
	TokenGt    // '>'
	TokenLeq   // '<='
	TokenGeq   // '>='
)

var keywords = map[string]TokenType{
	"let": TokenLet,
	"or":  TokenOr,
	"and": TokenAnd,
	"fn":  TokenFn,
	"end": TokenEnd,
	"while": TokenWhile,
}

type Scanner struct {
	filename  string // filename
	input     string // input
	ch        rune   // current rune under inspection
	pos       int    // index to the next rune to read.
	line      int    // our current positions in the input
	col       int
	start     int // starting position where we began reading the token
	startLine int
	startCol  int
	err       bool
	tokens    []Token // list of tokens
	errors    []Error // list of errors encountered
}

func New(filename string, input string) *Scanner {
	s := &Scanner{
		filename:  filename,
		input:     input,
		line:      1,
		col:       1,
		startLine: 1,
		startCol:  1,
		tokens:    []Token{},
		errors:    []Error{},
	}
	return s
}

// More() returns true if there is more input to be read
// from the input stream, and we haven't seen too many
// errors yet.
func (s *Scanner) More() bool {
	// we have to use <= here, so that the token stream
	// produces one EOF
	return s.pos <= len(s.input) && !s.err && len(s.errors) <= 10
}

func (s *Scanner) Tokens() []Token { return s.tokens }
func (s *Scanner) Errors() []Error {
	if len(s.errors) == 0 {
		return nil
	}
	return s.errors
}

// addToken adds a token under the current input.
func (s *Scanner) addToken(typ TokenType) {
	s.tokens = append(s.tokens, Token{
		Type:   typ,
		Value:  s.input[s.start:s.pos],
		LineNo: s.startLine,
		Column: s.startCol,
	})
	s.start = s.pos
	s.startLine = s.line
	s.startCol = s.col
}

// addError adds an error under the current input.
func (s *Scanner) addError(f string, args ...interface{}) {
	s.errors = append(s.errors, Error{
		Filename: s.filename,
		Message:  fmt.Sprintf(f, args...),
		Value:    s.input[s.start:s.pos],
		LineNo:   s.startLine,
		Column:   s.startCol,
	})
	s.start = s.pos
	s.startLine = s.line
	s.startCol = s.col
}

func (s *Scanner) advance() rune {
	if s.pos >= len(s.input) {
		s.pos++ // increment here -- want to avoid More() from looping forever
		s.ch = 0
		return s.ch
	}
	r, w := utf8.DecodeRuneInString(s.input[s.pos:])
	if r == utf8.RuneError {
		s.err = true
	}
	s.pos += w
	s.ch = r
	if r == '\n' {
		s.line++
		s.col = 0
	}
	s.col++
	return s.ch
}

func (s *Scanner) peek() rune {
	if s.pos == len(s.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(s.input[s.pos:])
	return r
}

// match advances the scanner if the lookahead runes
// match the given runes.
func (s *Scanner) match(prefix ...rune) bool {
	d := 0
	for _, p := range prefix {
		r, w := utf8.DecodeRuneInString(s.input[s.pos+d:])
		if r != p {
			return false
		}
		d += w
	}
	for i := 0; i < len(prefix); i++ {
		s.advance()
	}
	return true
}

// Scan() advances the scanner by at least one rune.
func (s *Scanner) Scan() {
	s.advance()
	switch s.ch {
	case 0:
		s.tokens = append(s.tokens, Token{TokenEOF, "", s.line, s.col})
	case ' ', '\t':
		s.munchWhitespace()
	case '\r', '\n':
		s.matchRun("\n\r \t")
		s.addToken(TokenSeparator)
	case '/':
		if s.match('/') {
			// a comment -- match up to newline or EOF
			for s.ch != '\n' && s.ch != 0 {
				s.advance()
			}
			s.ignore()
		} else {
			s.addToken(TokenDiv)
		}
	case '=':
		if s.match('=') {
			s.addToken(TokenEq)
		} else {
			s.addToken(TokenSet)
		}
	case '!':
		if s.match('=') {
			s.addToken(TokenNeq)
		} else {
			s.addToken(TokenBang)
		}
	case '*':
		s.addToken(TokenMul)
	case '+':
		s.addToken(TokenPlus)
	case '-':
		s.addToken(TokenMinus)
	case '<':
		if s.match('=') {
			s.addToken(TokenLeq)
		} else {
			s.addToken(TokenLt)
		}
	case '>':
		if s.match('=') {
			s.addToken(TokenGeq)
		} else {
			s.addToken(TokenGt)
		}
	case '.':
		s.addToken(TokenDot)
	case '(':
		s.addToken(TokenLParen)
	case ')':
		s.addToken(TokenRParen)
	case '{':
		s.addToken(TokenLBrace)
	case '}':
		s.addToken(TokenRBrace)
	case '[':
		s.addToken(TokenLBracket)
	case ']':
		s.addToken(TokenRBracket)
	case ';':
		s.addToken(TokenSemicolon)
	default:
		if isDigit(s.ch) {
			s.scanNumber()
		} else if isLetter(s.ch) {
			s.scanIdent()
		} else {
			s.addError("unrecognised character %U: %q", s.ch, s.ch)
		}
	}
}

// ignore ignores whatever was just consumed
func (s *Scanner) ignore() {
	s.start = s.pos
	s.startLine = s.line
	s.startCol = s.col
}

func (s *Scanner) munchWhitespace() {
	// whitespace tokens are '\t' and ' ', all ignored!
	s.matchRun("\t ")
	s.ignore()
}

func (s *Scanner) scanIdent() {
	// Idents match [a-zA-Z_][a-zA-Z_0-9]*'*
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}
	s.matchRun("'")
	word := s.input[s.start:s.pos]
	if typ, ok := keywords[word]; ok {
		s.addToken(typ)
	} else {
		s.addToken(TokenIdent)
	}
}

func (s *Scanner) scanNumber() {
	// we're currently on top of a digit.
	digits := "0123456789"
	s.matchRun(digits)
	if s.matchSet(".") {
		// this means that stuff like "2." are accepted.
		s.matchRun(digits)
	}
	s.addToken(TokenNumber)
}

func isAlphaNumeric(ch rune) bool {
	return isDigit(ch) || isLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch rune) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
}

func (s *Scanner) matchSet(set string) bool {
	if strings.ContainsRune(set, s.peek()) {
		s.advance()
		return true
	}
	return false
}

func (s *Scanner) matchRun(set string) {
	for strings.ContainsRune(set, s.peek()) {
		s.advance()
	}
}

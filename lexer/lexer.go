package lexer

import (
	"bytes"
	"fmt"
	"io"
	"jingle/token"
	"regexp"
)

// used when we have nothing to return
var EmptyToken = token.Token{Type: token.ILLEGAL}

// atom regexes
var (
	identRegex  = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*[']*$`)
	numberRegex = regexp.MustCompile(`^[0-9]+(\.[0-9]*)?$`)
)
// atom allowedchars
var (
	identAllowed = func(ch rune) bool { return isPunctuation(ch) }
	numberAllowed = func(ch rune) bool { return ch != '.' && isPunctuation(ch) }
)

type Lexer struct {
	Filename string
	input    *peekRuneReader
	// where are we in the file?
	position     int  // l.ch == l.input[position]
	readPosition int  // 1 after position
	ch           rune // current rune under examination
	// update this as we read the characters.
	column int
	lineNo int
	eof    *token.Token
}

func TryNew(filename string, r io.RuneReader) (*Lexer, error) {
	l := &Lexer{}
	l.input = newPeekRuneReader(r)
	l.Filename = ""
	if l.lineNo != 0 {
		panic("init() called twice")
	}
	l.lineNo = 1
	if err := l.advance(); err != nil {
		return nil, err
	}
	return l, nil
}

func New(s string) *Lexer {
	r := bytes.NewReader([]byte(s))
	l, err := TryNew("", r)
	if err != nil {
		panic(err)
	}
	return l
}

func (l *Lexer) advance() error {
	// check if we've seen an EOF token, so we don't keep reading.
	if l.eof != nil {
		return nil
	}
	r, _, err := l.input.ReadRune()
	if err != nil && err != io.EOF {
		return l.wrapError(err)
	}
	if err == io.EOF {
		r = 0
	}
	if l.ch == '\n' {
		// if the previous ch was a newline, then advance lineNo
		l.lineNo++
		l.column = 0
	}
	l.ch = r
	l.column++
	if l.ch == 0 {
		// to get the correct EOF token numbers -_-
		eofToken := l.emitRuneToken(token.EOF, 0)
		l.eof = &eofToken
	}
	return nil
}

// -----------------------------
// The meat of the code is here!
// -----------------------------

func (l *Lexer) NextToken() (token.Token, error) {
	tok := EmptyToken
	if l.ch == 0 {
		return *l.eof, nil
	}

	// Skip over whitespace
	for isWhiteSpace(l.ch) {
		if err := l.advance(); err != nil {
			return EmptyToken, err
		}
	}

	// Greedily match as many separators as possible
	if isSeparator(l.ch) {
		return l.scanSeparators()
	}

	// try to scan an operator
	// first two char operators...
	r1, r2 := l.ch, l.peek()
	rp := runePair{r1, r2}
	if tokenType, ok := twoCharOps[rp]; ok {
		tok = l.emitStringToken(tokenType, string(r1)+string(r2))
		if err := l.advance(); err != nil {
			return tok, err
		}
		goto end
	}
	// now one char operators...
	if tokenType, ok := oneCharOps[r1]; ok {
		tok = l.emitRuneToken(tokenType, r1)
		goto end
	}

	// try to scan a 'simple' literal (int, string, bool, null)
	switch {
	case l.ch == '"':
		return l.scanString()
	case isLetter(l.ch):
		tok, err := l.scanAtom(token.IDENT, identRegex, identAllowed)
		if err == nil {
			tok.Type = token.LookupIdent(tok.Literal)
		}
		return tok, err
	case isDigit(l.ch):
		return l.scanAtom(token.NUMBER, numberRegex, numberAllowed)
	}

end:
	// Still empty??
	if tok == EmptyToken {
		// need to put this here to prevent the advance()
		// from ruining our debug messages
		return EmptyToken,
			l.makeError("unhandled character: %U %q", l.ch, l.ch)
	}
	if err := l.advance(); err != nil {
		return EmptyToken, err
	}
	return tok, nil
}

// ========
// Utils...
// ========

func (l *Lexer) scanSeparators() (token.Token, error) {
	var buf bytes.Buffer
	tok := token.Token{Type: token.SEP, LineNo: l.lineNo, Column: l.column}
	for l.ch != 0 && isSeparator(l.ch) {
		// consume one, and then continue eating separators
		buf.WriteRune(l.ch)
		err := l.advance()
		if err != nil {
			return EmptyToken, err
		}
	}
	tok.Literal = buf.String()
	return tok, nil
}

func (l *Lexer) scanString() (token.Token, error) {
	escape := false
	var buf bytes.Buffer
	startingColumn := l.column
	startingLineNo := l.lineNo
outer:
	for {
		// skip over the " initially
		if err := l.advance(); err != nil {
			return EmptyToken, err
		}
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
				err := l.advance() // consume '"'
				if err != nil {
					return EmptyToken, err
				}
				return token.Token{
					Type:    token.STRING,
					Literal: buf.String(),
					LineNo:  startingLineNo,
					Column:  startingColumn,
				}, nil
			case '\n':
				break outer
			case '\r':
				break outer
			case 0:
				break outer
			default:
				buf.WriteRune(l.ch)
			}
		}
	}
	err := l.makeError("invalid char in string literal: %q", string(l.ch))
	for l.ch != 0 && l.ch != '"' {
		// try to consume until the next "
		if err := l.advance(); err != nil {
			return EmptyToken, err
		}
	}
	// consume the "
	if err := l.advance(); err != nil {
		return EmptyToken, err
	}
	return EmptyToken, err
}

func (l *Lexer) scanAtom(
	t token.TokenType,
	re *regexp.Regexp,
	notAllowed func(rune) bool,
) (token.Token, error) {
	// save these, because we will ruin them in a second.
	startingLineNo := l.lineNo
	startingColumn := l.column
	// begin scanning!
	var buf bytes.Buffer
	for !(l.ch == 0 || isWhiteSpace(l.ch) || isSeparator(l.ch)) && !notAllowed(l.ch) {
		buf.WriteRune(l.ch)
		err := l.advance()
		if err != nil {
			return EmptyToken, err
		}
	}
	// try to match the atom
	str := buf.String()
	if !re.MatchString(str) {
		err := newErrorFromLexer(l)
		err.message = fmt.Sprintf("invalid %s: %q", t, str)
		err.LineNo = startingLineNo
		err.Column = startingColumn
		return EmptyToken, err
	}
	return token.Token{
		Type:    t,
		Literal: str,
		LineNo:  startingLineNo,
		Column:  startingColumn,
	}, nil
}

func (l *Lexer) peek() rune {
	p, err := l.input.Peek()
	if err == io.EOF {
		return 0
	}
	return p
}

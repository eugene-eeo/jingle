package lexer

import "fmt"

// Represents an error encountered in the Lexer.
// This can be returned as a result of lexer.input.Read(),
// or an invalid token being encountered.
type LexerError struct {
	Filename string
	LineNo   int
	Column   int
	message  string
	under    error
}

func newErrorFromLexer(l *Lexer) LexerError {
	return LexerError{
		Filename: l.Filename,
		LineNo:   l.input.lineNo,
		Column:   l.input.column,
	}
}

func (e LexerError) Unwrap() error { return e.under }
func (e LexerError) Error() string {
	if e.under != nil {
		return fmt.Sprintf("%s:%d:%d: %e",
			e.Filename, e.LineNo, e.Column,
			e.Unwrap())
	} else {
		return fmt.Sprintf("%s:%d:%d: %s",
			e.Filename, e.LineNo, e.Column,
			e.message)
	}
}

func (l *Lexer) makeError(
	format string,
	args ...interface{},
) error {
	err := newErrorFromLexer(l)
	err.message = fmt.Sprintf(format, args...)
	return err
}

func (l *Lexer) wrapError(err error) error {
	lexErr := newErrorFromLexer(l)
	lexErr.under = err
	return lexErr
}

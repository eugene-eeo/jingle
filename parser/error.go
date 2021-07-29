package parser

import (
	"fmt"
	"jingle/scanner"
)

type ParserError struct {
	Token    *scanner.Token
	Filename string
	Msg      string
	Err      error
}

func (pe ParserError) Unwrap() error { return pe.Err }
func (pe ParserError) Error() string {
	if pe.Err != nil {
		return pe.Err.Error()
	} else if pe.Token == nil {
		return fmt.Sprintf("%s:%s", pe.Filename, pe.Msg)
	} else {
		return fmt.Sprintf("%s:%d:%d:%s",
			pe.Filename,
			pe.Token.LineNo,
			pe.Token.Column,
			pe.Msg,
		)
	}
}

func (p *Parser) errorErr(e error) {
	panic(ParserError{
		Filename: p.filename,
		Err:      e,
	})
}

func (p *Parser) errorStr(s string, args ...interface{}) {
	panic(ParserError{
		Filename: p.filename,
		Msg:      fmt.Sprintf(s, args...),
	})
}

func (p *Parser) errorToken(s string, args ...interface{}) {
	tok := p.last(1)
	panic(ParserError{
		Filename: p.filename,
		Token:    &tok,
		Msg:      fmt.Sprintf(s, args...),
	})
}

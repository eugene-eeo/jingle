package parser

import (
	"fmt"
	"jingle/scanner"
)

type ParserError struct {
	Token    scanner.Token
	Filename string
	Msg      string
}

func (pe ParserError) Error() string { return pe.String() }
func (pe ParserError) String() string {
	return fmt.Sprintf("%s:%d:%d:%s",
		pe.Filename,
		pe.Token.LineNo,
		pe.Token.Column,
		pe.Msg,
	)
}

func (p *Parser) errorToken(s string, args ...interface{}) {
	panic(ParserError{
		Filename: p.filename,
		Token:    p.previous(),
		Msg:      fmt.Sprintf(s, args...),
	})
}

package scanner

import "fmt"

// Token represents a token returned from the scanner.
type Token struct {
	Type   TokenType
	Value  string
	LineNo int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%d:%d:%q)", t.Type, t.LineNo, t.Column, t.Value)
}

//go:generate stringer -type=TokenType
type TokenType int

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
	TokenIn    // 'in'
	TokenDo    // 'do'
	// Literals
	TokenNil     // nil
	TokenBoolean // true or false
	TokenString  // a string literal
	TokenNumber  // floating point number (yuck)
	TokenIdent   // identifier
	// Delimiters
	TokenComma     // ','
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
	"let":   TokenLet,
	"or":    TokenOr,
	"and":   TokenAnd,
	"fn":    TokenFn,
	"end":   TokenEnd,
	"for":   TokenFor,
	"while": TokenWhile,
	"in":    TokenIn,
	"do":    TokenDo,
	"nil":   TokenNil,
	"true":  TokenBoolean,
	"false": TokenBoolean,
}

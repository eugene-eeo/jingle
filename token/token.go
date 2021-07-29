package token

type TokenType string

type Token struct {
	Type TokenType
	// TODO: make Literal a []rune to save memory
	Literal string // The contents of the token (what was consumed)
	LineNo  int
	Column  int
}

const (
	ILLEGAL   = "ILLEGAL"
	EOF       = "EOF"
	SEP       = "SEP" // _separators_, i.e. '[\r\n]+'
	SEMICOLON = ";"

	// Identifiers + literals
	IDENT  = "IDENT"
	NUMBER = "NUMBER"
	STRING = "STRING"

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	COLON    = ":"
	DOT      = "."

	OR  = "||"
	AND = "&&"

	LT = "<"
	GT = ">"

	LEQ    = "<="
	GEQ    = ">="
	EQ     = "=="
	NOT_EQ = "!="

	// Delimiters
	COMMA    = ","
	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	NULL     = "NULL"
	END      = "END"
	WHILE    = "WHILE"
	DO       = "DO"
	// FOR      = "FOR"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"true":   TRUE,
	"false":  FALSE,
	"null":   NULL,
	"end":    END,
	"while":  WHILE,
	"do":     DO,
	// "for":    FOR,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

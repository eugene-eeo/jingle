package lexer

import "jingle/token"

// ==============
// Operator Table
// ==============

type operator struct {
	ch        string
	tokenType token.TokenType
}

var operators = map[string]operator{}

func init() {
	opList := []operator{
		{"\u0000", token.EOF},
		{".", token.DOT},
		{";", token.SEMICOLON},
		{"=", token.ASSIGN},
		{"+", token.PLUS},
		{"-", token.MINUS},
		{"!", token.BANG},
		{"*", token.ASTERISK},
		{"/", token.SLASH},
		{":", token.COLON},
		{"||", token.OR},
		{"&&", token.AND},
		{"<", token.LT},
		{">", token.GT},
		{"<=", token.LEQ},
		{">=", token.GEQ},
		{"==", token.EQ},
		{"!=", token.NOT_EQ},
		{",", token.COMMA},
		{"(", token.LPAREN},
		{")", token.RPAREN},
		{"{", token.LBRACE},
		{"}", token.RBRACE},
		{"[", token.LBRACKET},
		{"]", token.RBRACKET},
	}
	for _, op := range opList {
		operators[op.ch] = op
	}
}

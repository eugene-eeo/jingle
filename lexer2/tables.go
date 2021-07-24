package lexer2

import "jingle/token"

// ==============
// Operator Table
// ==============

type operator struct {
	ch        string
	tokenType token.TokenType
}

type runePair [2]rune

var oneCharOps = map[rune]token.TokenType{}
var twoCharOps = map[runePair]token.TokenType{}

func init() {
	allOperators := []operator{
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
		{";", token.SEMICOLON},
		{"(", token.LPAREN},
		{")", token.RPAREN},
		{"{", token.LBRACE},
		{"}", token.RBRACE},
		{"[", token.LBRACKET},
		{"]", token.RBRACKET},
		{"\u0000", token.EOF},
	}
	for _, op := range allOperators {
		if len(op.ch) == 1 {
			oneCharOps[rune(op.ch[0])] = op.tokenType
			continue
		}
		if len(op.ch) == 2 {
			r := runePair{
				rune(op.ch[0]),
				rune(op.ch[1]),
			}
			twoCharOps[r] = op.tokenType
			continue
		}
	}
}

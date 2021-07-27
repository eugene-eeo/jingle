package parser

import (
	// "fmt"
	"jingle/ast"
	"jingle/token"
	"strconv"
)

type (
	// parse for a prefix expression: <OP><EXPR>
	prefixParseFn func() ast.Node
	// parse for an infix expression: <EXPR><OP><EXPR>
	infixParseFn func(ast.Node) ast.Node
)

const (
	LOWEST     = iota
	ASSIGNMENT // assignment
	ADD        // addition, subtraction
	PRODUCT    // multiplication
)

func (p *Parser) initExpressions() {
	p.prefixHandlers = map[token.TokenType]prefixParseFn{
		token.IDENT:  p.parseIdentifierLiteral,
		token.NULL:   p.parseNullLiteral,
		token.NUMBER: p.parseNumberLiteral,
		token.STRING: p.parseStringLiteral,
	}
	p.infixHandlers = map[token.TokenType]infixParseFn{}
	p.precedence = map[token.TokenType]int{}
}

// ===================
// Simple Pratt parser
// ===================
func (p *Parser) parseExpression() ast.Node { return p.parsePrecedence(LOWEST) }
func (p *Parser) parsePrecedence(precedence int) ast.Node {
	// must have a matching prefix parser, otherwise we cannot
	// parse anything!
	tok := p.consume()
	// fmt.Printf("tok=%+v\n", tok)
	prefixParser, ok := p.prefixHandlers[tok.Type]
	if !ok {
		p.errorToken("unrecognised expression (%s)", tok.Type)
	}
	left := prefixParser()
	for precedence < p.getPrecedence(p.current().Type) {
		// note -- we will never come here if p.getPrecedence()
		// returned LOWEST, since all other precedences > LOWEST.
		tok := p.consume()
		infixParser := p.infixHandlers[tok.Type]
		left = infixParser(left)
	}
	return left
}

func (p *Parser) getPrecedence(tok token.TokenType) int {
	prec, ok := p.precedence[tok]
	if !ok {
		return LOWEST
	}
	return prec
}

// ===================
// Expression Literals
// ===================

func (p *Parser) parseIdentifierLiteral() ast.Node {
	return &ast.IdentifierLiteral{Token: p.last(1)}
}

func (p *Parser) parseNullLiteral() ast.Node {
	return &ast.NullLiteral{Token: p.last(1)}
}

func (p *Parser) parseNumberLiteral() ast.Node {
	tok := p.last(1)
	val, err := strconv.ParseFloat(tok.Literal, 64)
	if err != nil {
		p.errorToken("invalid number: %e", err)
	}
	return &ast.NumberLiteral{Token: tok, Value: val}
}

func (p *Parser) parseStringLiteral() ast.Node {
	tok := p.last(1)
	return &ast.StringLiteral{Token: tok, Value: tok.Literal}
}

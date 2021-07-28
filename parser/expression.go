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
	PREC_LOWEST     = iota
	PREC_ASSIGNMENT // assignment
	PREC_ADD        // addition, subtraction
	PREC_PRODUCT    // multiplication
)

func (p *Parser) initExpressions() {
	p.prefixHandlers = map[token.TokenType]prefixParseFn{
		token.IDENT:  p.parseIdentifierLiteral,
		token.NULL:   p.parseNullLiteral,
		token.NUMBER: p.parseNumberLiteral,
		token.STRING: p.parseStringLiteral,
		token.LPAREN: p.parseParens,
	}
	p.infixHandlers = map[token.TokenType]infixParseFn{
		token.PLUS:     p.parseInfixExpression,
		token.MINUS:    p.parseInfixExpression,
		token.ASTERISK: p.parseInfixExpression,
		token.SLASH:    p.parseInfixExpression,
		token.ASSIGN:   p.parseAssigmentExpression,
	}
	p.precedence = map[token.TokenType]int{
		token.PLUS:     PREC_ADD,
		token.MINUS:    PREC_ADD,
		token.ASTERISK: PREC_PRODUCT,
		token.SLASH:    PREC_PRODUCT,
		token.ASSIGN:   PREC_ASSIGNMENT,
	}
}

// ===================
// Simple Pratt parser
// ===================
func (p *Parser) parseExpression() ast.Node { return p.parsePrecedence(PREC_LOWEST) }
func (p *Parser) parsePrecedence(precedence int) ast.Node {
	// must have a matching prefix parser, otherwise we cannot
	// parse anything!
	tok := p.consume()
	// fmt.Printf("tok=%+v\n", tok)
	prefixParser, ok := p.prefixHandlers[tok.Type]
	if !ok {
		p.errorToken("unrecognised expression: %q", tok.Type)
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
		return PREC_LOWEST
	}
	return prec
}

// ===========
// Expressions
// ===========

func (p *Parser) parseInfixExpression(left ast.Node) ast.Node {
	opToken := p.last(1)
	right := p.parsePrecedence(p.precedence[opToken.Type])
	return &ast.InfixExpression{
		Token: opToken,
		Op:    opToken.Literal,
		Left:  left,
		Right: right,
	}
}

func (p *Parser) parseAssigmentExpression(left ast.Node) ast.Node {
	// left should be an ident!
	leftIdent, ok := left.(*ast.IdentifierLiteral)
	if !ok {
		p.errorToken("cannot assign to %s", ast.NodeTypeAsString(left.Type()))
	}
	return &ast.AssignmentExpression{
		Token: p.last(1), // the '=' token
		Left:  leftIdent,
		Right: p.parsePrecedence(PREC_ASSIGNMENT),
	}
}

func (p *Parser) parseParens() ast.Node {
	// This is a grouping operator.
	expr := p.parseStatement()
	p.expect(token.RPAREN)
	return expr
}

// ========
// Literals
// ========

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

package parser

import (
	"jingle/ast"
	"jingle/scanner"
	"strconv"
)

type (
	// parse prefix expression: <OP><EXPR>
	prefixParseFn func() ast.Expression
	// parse infix expression: <EXPR><OP><EXPR>
	infixParseFn func(ast.Expression) ast.Expression
)

const (
	PREC_LOWEST     = iota
	PREC_ASSIGNMENT // assignment
	PREC_DOT        // attr access
	PREC_EQ         // ==, >=, !=
	PREC_AND_OR     // &&, ||
	PREC_ADD        // addition, subtraction
	PREC_PRODUCT    // multiplication
)

func (p *Parser) initExpressions() {
	p.prefixHandlers = map[scanner.TokenType]prefixParseFn{
		scanner.TokenIdent:    p.parseIdentifierLiteral,
		scanner.TokenNil:      p.parseNullLiteral,
		scanner.TokenNumber:   p.parseNumberLiteral,
		scanner.TokenString:   p.parseStringLiteral,
		scanner.TokenLParen:   p.parseParens,
		scanner.TokenBoolean:  p.parseBooleanLiteral,
		scanner.TokenFn:       p.parseFunctionLiteral,
		scanner.TokenLBracket: p.parseArrayLiteral,
	}
	p.infixHandlers = map[scanner.TokenType]infixParseFn{
		scanner.TokenPlus:  p.parseInfixExpression,
		scanner.TokenMinus: p.parseInfixExpression,
		scanner.TokenMul:   p.parseInfixExpression,
		scanner.TokenDiv:   p.parseInfixExpression,
		scanner.TokenGeq:   p.parseInfixExpression,
		scanner.TokenLeq:   p.parseInfixExpression,
		scanner.TokenSet:   p.parseAssigmentExpression,
		scanner.TokenOr:    p.parseOrExpression,
		scanner.TokenAnd:   p.parseAndExpression,
		scanner.TokenEq:    p.parseInfixExpression,
		scanner.TokenNeq:   p.parseInfixExpression,
		scanner.TokenDot:   p.parseAttrExpression,
	}
	p.precedence = map[scanner.TokenType]int{
		scanner.TokenPlus:  PREC_ADD,
		scanner.TokenMinus: PREC_ADD,
		scanner.TokenMul:   PREC_PRODUCT,
		scanner.TokenDiv:   PREC_PRODUCT,
		scanner.TokenSet:   PREC_ASSIGNMENT,
		scanner.TokenOr:    PREC_AND_OR,
		scanner.TokenAnd:   PREC_AND_OR,
		scanner.TokenEq:    PREC_EQ,
		scanner.TokenNeq:   PREC_EQ,
		scanner.TokenLeq:   PREC_EQ,
		scanner.TokenGeq:   PREC_EQ,
		scanner.TokenDot:   PREC_DOT,
	}
}

// ===================
// Simple Pratt parser
// ===================
func (p *Parser) parseExpression() ast.Expression { return p.parsePrecedence(PREC_LOWEST) }
func (p *Parser) parsePrecedence(precedence int) ast.Expression {
	// must have a matching prefix parser, otherwise we cannot
	// parse anything!
	tok := p.consume()
	prefixParser, ok := p.prefixHandlers[tok.Type]
	if !ok {
		p.errorToken("expected expression, got %s", tok.Type)
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

func (p *Parser) getPrecedence(tok scanner.TokenType) int {
	prec, ok := p.precedence[tok]
	if !ok {
		return PREC_LOWEST
	}
	return prec
}

// ===========
// Expressions
// ===========

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	// <left> <op> <right>
	opToken := p.last(1)
	right := p.parsePrecedence(p.precedence[opToken.Type])
	return &ast.InfixExpression{
		Token: opToken,
		Op:    opToken.Value,
		Left:  left,
		Right: right,
	}
}

func (p *Parser) parseAssigmentExpression(left ast.Expression) ast.Expression {
	// <left> = <right>
	if !ast.Assignable(left) {
		p.errorToken("cannot assign to %s", left.Type())
		return nil // not reachable
	}
	return &ast.AssignmentExpression{
		Token: p.last(1), // the '=' token
		Left:  left,
		Right: p.parsePrecedence(PREC_ASSIGNMENT - 1),
	}
}

func (p *Parser) parseParens() ast.Expression {
	// This is a grouping operator.
	expr := p.parseExpression()
	p.expect(scanner.TokenRParen)
	return expr
}

func (p *Parser) parseOrExpression(left ast.Expression) ast.Expression {
	opToken := p.last(1)
	right := p.parsePrecedence(p.precedence[opToken.Type])
	return &ast.OrExpression{
		Token: opToken,
		Left:  left,
		Right: right,
	}
}

func (p *Parser) parseAndExpression(left ast.Expression) ast.Expression {
	opToken := p.last(1)
	right := p.parsePrecedence(p.precedence[opToken.Type])
	return &ast.AndExpression{
		Token: opToken,
		Left:  left,
		Right: right,
	}
}

func (p *Parser) parseAttrExpression(left ast.Expression) ast.Expression {
	// <left>.IDENT = <expr>
	opToken := p.last(1)
	right := p.parsePrecedence(PREC_DOT + 1)
	if right.Type() != ast.IDENTIFIER_LITERAL {
		p.errorToken("unexpected %s", right.Type())
	}
	return &ast.AttrExpression{
		Token: opToken,
		Left:  left,
		Right: right.(*ast.IdentifierLiteral),
	}
}

// ========
// Literals
// ========

func (p *Parser) parseIdentifierLiteral() ast.Expression {
	return &ast.IdentifierLiteral{Token: p.last(1)}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{
		Token: p.last(1),
		Value: p.last(1).Value == "true",
	}
}

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NilLiteral{Token: p.last(1)}
}

func (p *Parser) parseNumberLiteral() ast.Expression {
	tok := p.last(1)
	val, err := strconv.ParseFloat(tok.Value, 64)
	if err != nil {
		p.errorToken("invalid number: %e", err)
	}
	return &ast.NumberLiteral{Token: tok, Value: val}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	tok := p.last(1)
	return &ast.StringLiteral{Token: tok, Value: tok.Value}
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	// fn → "fn" "(" params ")" stmt* "end"
	// params → nothing | ident ("," | "," params)?
	tok := p.last(1)
	fn := &ast.FunctionLiteral{
		Token:  tok,
		Params: []*ast.IdentifierLiteral{},
	}
	// parse the parameters
	p.expect(scanner.TokenLParen)
	for !p.match(scanner.TokenRParen) {
		p.expect(scanner.TokenIdent)
		ident := p.parseIdentifierLiteral()
		fn.Params = append(fn.Params, ident.(*ast.IdentifierLiteral))
		if !p.match(scanner.TokenComma) {
			// dont have a comma -- must be an RPAREN
			p.expect(scanner.TokenRParen)
			break
		}
	}
	fn.Body = p.parseBlock(scanner.TokenEnd)
	return fn
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	// list → "[" listElems "]"
	// listElems → nothing | <expr> ( "," | "," listElems )?
	tok := p.last(1)
	arr := &ast.ArrayLiteral{Token: tok, Elems: []ast.Node{}}
	for !p.match(scanner.TokenRBracket) {
		node := p.parseExpression()
		arr.Elems = append(arr.Elems, node)
		if !p.match(scanner.TokenComma) {
			// must be a RBracket
			p.expect(scanner.TokenRBracket)
			break
		}
	}
	return arr
}

package parser

import (
	// "fmt"
	"jingle/ast"
	"jingle/token"
	"strconv"
)

type (
	// parse prefix expression: <OP><EXPR>
	prefixParseFn func() ast.Node
	// parse infix expression: <EXPR><OP><EXPR>
	infixParseFn func(ast.Node) ast.Node
)

const (
	PREC_LOWEST     = iota
	PREC_ASSIGNMENT // assignment
	PREC_DOT        // attr access
	PREC_EQ         // ==
	PREC_AND_OR     // &&, ||
	PREC_ADD        // addition, subtraction
	PREC_PRODUCT    // multiplication
)

func (p *Parser) initExpressions() {
	p.prefixHandlers = map[token.TokenType]prefixParseFn{
		token.IDENT:    p.parseIdentifierLiteral,
		token.NULL:     p.parseNullLiteral,
		token.NUMBER:   p.parseNumberLiteral,
		token.STRING:   p.parseStringLiteral,
		token.LPAREN:   p.parseParens,
		token.TRUE:     p.parseBooleanLiteral,
		token.FALSE:    p.parseBooleanLiteral,
		token.FUNCTION: p.parseFunctionLiteral,
	}
	p.infixHandlers = map[token.TokenType]infixParseFn{
		token.PLUS:     p.parseInfixExpression,
		token.MINUS:    p.parseInfixExpression,
		token.ASTERISK: p.parseInfixExpression,
		token.SLASH:    p.parseInfixExpression,
		token.ASSIGN:   p.parseAssigmentExpression,
		token.OR:       p.parseOrExpression,
		token.AND:      p.parseAndExpression,
		token.EQ:       p.parseInfixExpression,
		token.NOT_EQ:   p.parseInfixExpression,
		token.DOT:      p.parseAttrExpression,
	}
	p.precedence = map[token.TokenType]int{
		token.PLUS:     PREC_ADD,
		token.MINUS:    PREC_ADD,
		token.ASTERISK: PREC_PRODUCT,
		token.SLASH:    PREC_PRODUCT,
		token.ASSIGN:   PREC_ASSIGNMENT,
		token.OR:       PREC_AND_OR,
		token.AND:      PREC_AND_OR,
		token.EQ:       PREC_EQ,
		token.NOT_EQ:   PREC_EQ,
		token.DOT:      PREC_DOT,
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
	// <left> <op> <right>
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
	// <left> = <right>
	switch left.Type() {
	// for now, left should be an ident.
	case ast.IDENTIFIER_LITERAL:
		return &ast.AssignmentExpression{
			Token: p.last(1), // the '=' token
			Left:  left,
			Right: p.parsePrecedence(PREC_ASSIGNMENT-1),
		}
	default:
		p.errorToken("cannot assign to %s", ast.NodeTypeAsString(left.Type()))
		return nil // not reachable
	}
}

func (p *Parser) parseParens() ast.Node {
	// This is a grouping operator.
	expr := p.parseExpression()
	p.expect(token.RPAREN)
	return expr
}

func (p *Parser) parseOrExpression(left ast.Node) ast.Node {
	opToken := p.last(1)
	right := p.parsePrecedence(p.precedence[opToken.Type])
	return &ast.OrExpression{
		Token: opToken,
		Left:  left,
		Right: right,
	}
}

func (p *Parser) parseAndExpression(left ast.Node) ast.Node {
	opToken := p.last(1)
	right := p.parsePrecedence(p.precedence[opToken.Type])
	return &ast.AndExpression{
		Token: opToken,
		Left:  left,
		Right: right,
	}
}

func (p *Parser) parseBlockExpression() ast.Node {
	// No token needed for a parseBlock.
	// block expressions:
	//       expr
	//       expr
	//       ...
	//    end
	lastHasSeparator := true
	block := &ast.BlockExpression{}
	block.Nodes = []ast.Node{}
	for !p.match(token.END) {
		if !lastHasSeparator {
			p.errorToken("expected a newline or semicolon")
		}
		block.Nodes = append(block.Nodes, p.parseStatement())
		lastHasSeparator = p.matchAny(token.SEP, token.SEMICOLON)
	}
	return block
}

func (p *Parser) parseAttrExpression(left ast.Node) ast.Node {
	// <left>.IDENT = <expr>
	opToken := p.last(1)
	right := p.parsePrecedence(PREC_DOT+1)
	if right.Type() != ast.IDENTIFIER_LITERAL {
		p.errorToken("unexpected %s", ast.NodeTypeAsString(right.Type()))
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

func (p *Parser) parseIdentifierLiteral() ast.Node {
	return &ast.IdentifierLiteral{Token: p.last(1)}
}

func (p *Parser) parseBooleanLiteral() ast.Node {
	return &ast.BooleanLiteral{
		Token: p.last(1),
		Value: p.last(1).Type == token.TRUE,
	}
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

func (p *Parser) parseFunctionLiteral() ast.Node {
	// "fn" "(" <ident>, ... ")" <block>
	tok := p.last(1)
	fn := &ast.FunctionLiteral{
		Token: tok,
		Params: []*ast.IdentifierLiteral{},
	}
	// parse the parameters
	p.expect(token.LPAREN)
	for !p.match(token.RPAREN) {
		p.expect(token.IDENT)
		ident := p.parseIdentifierLiteral()
		// p.consume()
		fn.Params = append(fn.Params, ident.(*ast.IdentifierLiteral))
		// try to match a comma
		if p.match(token.COMMA) {
			continue
		} else {
			// dont have a comma -- must be an RPAREN
			p.expect(token.RPAREN)
			break
		}
	}
	fn.Body = p.parseBlockExpression().(*ast.BlockExpression)
	return fn
}

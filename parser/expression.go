package parser

import (
	"jingle/ast"
	"jingle/scanner"
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
	PREC_EQ         // ==, >=, !=
	PREC_AND_OR     // &&, ||
	PREC_ADD        // addition, subtraction
	PREC_PRODUCT    // multiplication
)

func (p *Parser) initExpressions() {
	p.prefixHandlers = map[scanner.TokenType]prefixParseFn{
		scanner.TokenIdent:   p.parseIdentifierLiteral,
		scanner.TokenNil:     p.parseNullLiteral,
		scanner.TokenNumber:  p.parseNumberLiteral,
		scanner.TokenString:  p.parseStringLiteral,
		scanner.TokenLParen:  p.parseParens,
		scanner.TokenBoolean: p.parseBooleanLiteral,
		scanner.TokenFn:      p.parseFunctionLiteral,
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

func (p *Parser) parseInfixExpression(left ast.Node) ast.Node {
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

func (p *Parser) parseAssigmentExpression(left ast.Node) ast.Node {
	// <left> = <right>
	switch left.Type() {
	// for now, left should be an ident.
	case ast.IDENTIFIER_LITERAL:
		return &ast.AssignmentExpression{
			Token: p.last(1), // the '=' token
			Left:  left,
			Right: p.parsePrecedence(PREC_ASSIGNMENT - 1),
		}
	default:
		p.errorToken("cannot assign to %s", ast.NodeTypeAsString(left.Type()))
		return nil // not reachable
	}
}

func (p *Parser) parseParens() ast.Node {
	// This is a grouping operator.
	expr := p.parseExpression()
	p.expect(scanner.TokenRParen)
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
	for !p.match(scanner.TokenEnd) {
		if !lastHasSeparator {
			p.errorToken("expected newline or semicolon, got %s instead", p.current().Type)
		}
		block.Nodes = append(block.Nodes, p.parseStatement())
		lastHasSeparator = p.matchAny(scanner.TokenSeparator, scanner.TokenSemicolon)
	}
	return block
}

func (p *Parser) parseAttrExpression(left ast.Node) ast.Node {
	// <left>.IDENT = <expr>
	opToken := p.last(1)
	right := p.parsePrecedence(PREC_DOT + 1)
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
		Value: p.last(1).Value == "true",
	}
}

func (p *Parser) parseNullLiteral() ast.Node {
	return &ast.NilLiteral{Token: p.last(1)}
}

func (p *Parser) parseNumberLiteral() ast.Node {
	tok := p.last(1)
	val, err := strconv.ParseFloat(tok.Value, 64)
	if err != nil {
		p.errorToken("invalid number: %e", err)
	}
	return &ast.NumberLiteral{Token: tok, Value: val}
}

func (p *Parser) parseStringLiteral() ast.Node {
	tok := p.last(1)
	return &ast.StringLiteral{Token: tok, Value: tok.Value}
}

func (p *Parser) parseFunctionLiteral() ast.Node {
	// "fn" "(" <ident>, ... ")" <block>
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
		// p.consume()
		fn.Params = append(fn.Params, ident.(*ast.IdentifierLiteral))
		// try to match a comma
		if p.match(scanner.TokenComma) {
			continue
		} else {
			// dont have a comma -- must be an RPAREN
			p.expect(scanner.TokenRParen)
			break
		}
	}
	fn.Body = p.parseBlockExpression().(*ast.BlockExpression)
	return fn
}

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
	PREC_IF         // x if foo else bar
	PREC_INDEX      // a[b]
	PREC_EQ         // ==, >=, !=
	PREC_AND_OR     // and, or
	PREC_ADD        // addition, subtraction
	PREC_PRODUCT    // multiplication
	PREC_CALL       // func/method calls, attr get
)

func (p *Parser) initExpressions() {
	p.prefixHandlers = map[scanner.TokenType]prefixParseFn{
		scanner.TokenMinus:    p.parsePrefixExpression,
		scanner.TokenBang:     p.parsePrefixExpression,
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
		scanner.TokenPlus:     p.parseInfixExpression,
		scanner.TokenMinus:    p.parseInfixExpression,
		scanner.TokenMul:      p.parseInfixExpression,
		scanner.TokenDiv:      p.parseInfixExpression,
		scanner.TokenGeq:      p.parseInfixExpression,
		scanner.TokenLeq:      p.parseInfixExpression,
		scanner.TokenEq:       p.parseInfixExpression,
		scanner.TokenNeq:      p.parseInfixExpression,
		scanner.TokenSet:      p.parseAssigmentExpression,
		scanner.TokenOr:       p.parseOrExpression,
		scanner.TokenAnd:      p.parseAndExpression,
		scanner.TokenDot:      p.parseAttrExpression,
		scanner.TokenLBracket: p.parseIndexExpression,
		scanner.TokenLParen:   p.parseCallExpression,
		scanner.TokenIf:       p.parseIfElseExpression,
	}
	p.precedence = map[scanner.TokenType]int{
		scanner.TokenPlus:     PREC_ADD,
		scanner.TokenMinus:    PREC_ADD,
		scanner.TokenMul:      PREC_PRODUCT,
		scanner.TokenDiv:      PREC_PRODUCT,
		scanner.TokenSet:      PREC_ASSIGNMENT,
		scanner.TokenOr:       PREC_AND_OR,
		scanner.TokenAnd:      PREC_AND_OR,
		scanner.TokenEq:       PREC_EQ,
		scanner.TokenNeq:      PREC_EQ,
		scanner.TokenLeq:      PREC_EQ,
		scanner.TokenGeq:      PREC_EQ,
		scanner.TokenDot:      PREC_CALL,
		scanner.TokenLBracket: PREC_INDEX,
		scanner.TokenLParen:   PREC_CALL,
		scanner.TokenIf:       PREC_IF,
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
		p.error("expected expression, got %s", tok.Type)
	}
	left := prefixParser()
	for precedence < p.getPrecedence(p.peek().Type) {
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

func (p *Parser) parsePrefixExpression() ast.Expression {
	// prefix → ("!" | "-") expr
	opToken := p.previous()
	right := p.parseExpression()
	return &ast.PrefixExpression{
		Token: opToken,
		Op:    opToken.Value,
		Expr:  right,
	}
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	// infix → expr ("*" | "/" | "+" | "-" | ">" | "<" | "==" | "!=" | "<=" | ">=") expr
	opToken := p.previous()
	right := p.parsePrecedence(p.precedence[opToken.Type])
	return &ast.InfixExpression{
		Token: opToken,
		Op:    opToken.Value,
		Left:  left,
		Right: right,
	}
}

func (p *Parser) parseAssigmentExpression(left ast.Expression) ast.Expression {
	// assignment → expr "=" expr
	if reason, ok := ast.Assignable(left, false); !ok {
		p.errorToken(reason.GetToken(),
			"cannot assign to %s", left.Type())
	}
	return &ast.AssignmentExpression{
		Token: p.previous(), // the '=' token
		Left:  left,
		Right: p.parsePrecedence(PREC_ASSIGNMENT - 1),
	}
}

func (p *Parser) parseParens() ast.Expression {
	// parens → "(" expr ")"
	expr := p.parseExpression()
	p.expect(scanner.TokenRParen)
	return expr
}

func (p *Parser) parseOrExpression(left ast.Expression) ast.Expression {
	// or → expr "or" expr
	opToken := p.previous()
	right := p.parsePrecedence(p.precedence[opToken.Type])
	return &ast.OrExpression{
		Token: opToken,
		Left:  left,
		Right: right,
	}
}

func (p *Parser) parseAndExpression(left ast.Expression) ast.Expression {
	// and → expr "and" expr
	opToken := p.previous()
	right := p.parsePrecedence(p.precedence[opToken.Type])
	return &ast.AndExpression{
		Token: opToken,
		Left:  left,
		Right: right,
	}
}

func (p *Parser) parseAttrExpression(left ast.Expression) ast.Expression {
	// attr → expr "." ident
	opToken := p.previous()
	right := p.parsePrecedence(PREC_CALL)
	if right.Type() != ast.IDENTIFIER_LITERAL {
		p.error("unexpected %s", right.Type())
	}
	return &ast.AttrExpression{
		Token:  opToken,
		Target: left,
		Name:   right.(*ast.IdentifierLiteral),
	}
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	// index → expr "[" expr ("]" | "," args("]"))
	tok := p.previous()
	args := []ast.Expression{p.parseExpression()}
	if !p.match(scanner.TokenRBracket) {
		// more to come?
		args = append(args, p.parseArgs(scanner.TokenRBracket)...)
	}
	return &ast.IndexExpression{
		Token:  tok,
		Target: left,
		Args:   args,
	}
}

// parse args parses a `terminal`-delimited sequence of expressions,
// each separated by a comma.
func (p *Parser) parseArgs(terminal scanner.TokenType) []ast.Expression {
	// args → <terminal> | expr ("," (args)?)?
	args := []ast.Expression{}
	for !p.match(terminal) {
		args = append(args, p.parseExpression())
		if !p.match(scanner.TokenComma) {
			p.expect(terminal)
			break
		}
	}
	return args
}

func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	// call → expr "(" args(")")
	return &ast.CallExpression{
		Token:  p.previous(),
		Target: left,
		Args:   p.parseArgs(scanner.TokenRParen),
	}
}

func (p *Parser) parseIfElseExpression(left ast.Expression) ast.Expression {
	// ifElse → expr "if" expr ("else" expr)
	node := &ast.IfElseExpression{Token: p.previous()}
	node.Then = left
	node.Cond = p.parseExpression()
	if p.match(scanner.TokenElse) {
		node.Else = p.parseExpression()
	}
	return node
}

// ========
// Literals
// ========

func (p *Parser) parseIdentifierLiteral() ast.Expression {
	return &ast.IdentifierLiteral{Token: p.previous()}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{
		Token: p.previous(),
		Value: p.previous().Value == "true",
	}
}

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NilLiteral{Token: p.previous()}
}

func (p *Parser) parseNumberLiteral() ast.Expression {
	tok := p.previous()
	val, err := strconv.ParseFloat(tok.Value, 64)
	if err != nil {
		p.error("invalid number: %e", err)
	}
	return &ast.NumberLiteral{Token: tok, Value: val}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		Token: p.previous(),
		Value: p.previous().Value,
	}
}

func (p *Parser) parseParams() []*ast.IdentifierLiteral {
	// params → nothing | "ident" ("," | "," params)?
	params := []*ast.IdentifierLiteral{}
	for !p.match(scanner.TokenRParen) {
		p.expect(scanner.TokenIdent)
		ident := p.parseIdentifierLiteral()
		params = append(params, ident.(*ast.IdentifierLiteral))
		if !p.match(scanner.TokenComma) {
			// dont have a comma -- must be an RPAREN
			p.expect(scanner.TokenRParen)
			break
		}
	}
	return params
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	// fn → "fn" "(" params ")" stmt* "end"
	fn := &ast.FunctionLiteral{Token: p.previous()}
	p.expect(scanner.TokenLParen)
	fn.Params = p.parseParams()
	fn.Body = p.parseBlock(false, true, scanner.TokenEnd)
	return fn
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	// list → "[" args "]"
	return &ast.ArrayLiteral{
		Token: p.previous(),
		Elems: p.parseArgs(scanner.TokenRBracket),
	}
}

package parser

import (
	"fmt"
	"jingle/ast"
	"jingle/lexer"
	"jingle/token"
)

type Parser struct {
	l *lexer.Lexer

	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.NULL, p.parseNull)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LEQ, p.parseInfixExpression)
	p.registerInfix(token.GEQ, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.IS, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseOrExpression)
	p.registerInfix(token.AND, p.parseAndExpression)
	p.registerInfix(token.ASSIGN, nil)

	// Read two tokens, so curToken and peekToken are set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) registerPrefix(t token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[t] = fn
}

func (p *Parser) registerInfix(t token.TokenType, fn infixParseFn) {
	p.infixParseFns[t] = fn
}

// =====================
// Parser Error Handling
// =====================

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) pushError(s string, args ...interface{}) {
	msg := fmt.Sprintf(s, args...)
	msg = fmt.Sprintf("%s:%d:%d: %s",
		p.l.Filename,
		p.curToken.LineNo,
		p.curToken.Column,
		msg)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekError(t token.TokenType) {
	p.pushError("expected next token to be %s, got %s instead.",
		t, p.peekToken.Type)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	tok, err := p.l.NextToken()
	if err != nil {
		p.errors = append(p.errors, err.Error())
	}
	p.peekToken = tok
}

// =================
// Statement parsing
// =================

// ParseProgram() returns a *ast.Program representing the program
// being fed into the parser. This is the main entry point into the
// parser -- start reading from here!
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	prevHasSemicolon := true

	for !p.curTokenIs(token.EOF) {
		if !prevHasSemicolon {
			p.pushError("expected a semicolon before the last statement")
		}
		stmt, hasSemicolon := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
		prevHasSemicolon = hasSemicolon
	}

	return program
}

func (p *Parser) parseStatement() (ast.Statement, bool) {
	// Semicolon rules:
	// ----------------
	// LET, RETURN, and all other Infix statements _always_ require a
	// semicolon before the last statement. Otherwise, if we're in an
	// ExpressionStatement, then only functions and if statements do
	// not require a semicolon.
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		var stmt ast.Statement
		var hasSemicolon bool
		tok := p.curToken
		node := p.parseExpressionOrInfixStatement(LOWEST, true)
		switch node := node.(type) {
		case ast.Expression:
			// expression statement
			stmt = &ast.ExpressionStatement{Token: tok, Expression: node}
			hasSemicolon = p.consumeOptionalSemicolon()
			// check if it's a function or an if expr
			switch node.(type) {
			case *ast.IfExpression:
				hasSemicolon = true
			case *ast.FunctionLiteral:
				hasSemicolon = true
			}
		case ast.Statement:
			stmt = node
			hasSemicolon = p.consumeOptionalSemicolon()
		}
		return stmt, hasSemicolon
	}
}

func (p *Parser) parseLetStatement() (ast.Statement, bool) {
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil, true
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil, true
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	hasSemicolon := p.consumeOptionalSemicolon()
	return stmt, hasSemicolon
}

func (p *Parser) parseReturnStatement() (ast.Statement, bool) {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	hasSemicolon := p.consumeOptionalSemicolon()
	return stmt, hasSemicolon
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	prevHasSemicolon := true

	p.nextToken()
	for !p.curTokenIs(token.RBRACE) {
		if !prevHasSemicolon {
			p.pushError("expected a semicolon before the last statement")
		}
		stmt, hasSemicolon := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
		prevHasSemicolon = hasSemicolon
		if p.curTokenIs(token.EOF) {
			p.pushError("unexpected EOF")
			return nil
		}
	}

	return block
}

func (p *Parser) parseSetStatement(left ast.Expression) ast.Statement {
	assignableStmt, ok := left.(ast.Assignable)
	if !ok {
		p.pushError("cannot assign to %T", left)
	}
	stmt := &ast.SetStatement{Left: assignableStmt, Token: p.curToken}
	p.nextToken()
	stmt.Right = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseOrExpression(left ast.Expression) ast.Expression {
	expression := &ast.OrExpression{
		Token: p.curToken,
		Left:  left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) parseAndExpression(left ast.Expression) ast.Expression {
	expression := &ast.AndExpression{
		Token: p.curToken,
		Left:  left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

// ================
// Helper functions
// ================

func (p *Parser) consumeOptionalSemicolon() bool {
	// Optional semicolon
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
		return true
	}
	return false
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}
	p.nextToken()
	for {
		if p.curTokenIs(end) {
			return list
		}
		expr := p.parseExpression(LOWEST)
		list = append(list, expr)
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			continue
		}
		// Not a comma -- better be an END
		if !p.expectPeek(end) {
			return nil
		}
		return list
	}
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

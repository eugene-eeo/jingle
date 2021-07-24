package parser

import (
	"fmt"
	"jingle/ast"
	"jingle/lexer"
	"jingle/token"
	"strconv"
)

// Expression parsing
const (
	_ int = iota
	LOWEST
	SET         // =
	OR          // || or &&
	IS          // is
	EQUALS      // == or !=
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

var precedences = map[token.TokenType]int{
	token.ASSIGN:   SET,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.AND:      OR,
	token.OR:       OR,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LEQ:      LESSGREATER,
	token.GEQ:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
	token.IS:       IS,
}

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
	p.registerInfix(token.ASSIGN, p.parseSetExpression)
	p.registerInfix(token.OR, p.parseOrExpression)
	p.registerInfix(token.AND, p.parseAndExpression)

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
	// LET and RETURN _always_ requires a semicolon before the last
	// statement. Otherwise, if we're in an ExpressionStatement, then
	// only functions and if statements do not require a semicolon.
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		stmt, hasSemicolon := p.parseExpressionStatement()
		if !hasSemicolon && stmt != nil {
			exprStmt := stmt.(*ast.ExpressionStatement)
			switch exprStmt.Expression.(type) {
			case *ast.IfExpression:
				hasSemicolon = true
			case *ast.FunctionLiteral:
				hasSemicolon = true
			}
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

func (p *Parser) parseExpressionStatement() (ast.Statement, bool) {
	// defer untrace(trace("parseExpressionStatement"))
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
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

// ============================
// Pratt Parser for expressions
// ============================

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	// defer untrace(trace("parseExpression"))
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	// defer untrace(trace("parsePrefixExpression"))
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	// we're on top of an operator, e.g. p.curToken == '!'
	// defer untrace(trace("parseInfixExpression"))
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

// ==========================
// Different expression types
// ==========================

func (p *Parser) parseIdentifier() ast.Expression {
	// defer untrace(trace("parseIdentifier"))
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	// defer untrace(trace("parseStringLiteral"))
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	// defer untrace(trace("parseIntegerLiteral"))
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNull() ast.Expression {
	return &ast.Null{Token: p.curToken}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	// we're on top of an LPAREN
	p.nextToken()
	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expr := &ast.IfExpression{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expr.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expr.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		expr.Alternative = p.parseBlockStatement()
	}

	return expr
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fn := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	fn.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	fn.Body = p.parseBlockStatement()
	return fn
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // skip the comma
		if p.peekTokenIs(token.RPAREN) {
			// this conditional allows for trailing commas,
			// see parseExpressionList
			break
		}
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		// p.nextToken() // now we're on top of the ident
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	call := &ast.CallExpression{Token: p.curToken, Function: function}
	call.Arguments = p.parseExpressionList(token.RPAREN)
	return call
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	// we're on top of a [
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken() // jump over the [
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken, Pairs: []ast.HashPair{}}
	p.nextToken()
	for {
		if p.curTokenIs(token.RBRACE) {
			return hash
		}
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		val := p.parseExpression(LOWEST)
		hash.Pairs = append(hash.Pairs, ast.HashPair{key, val})
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			continue
		}
		// Not a comma -- better be an RBRACE!
		if !p.expectPeek(token.RBRACE) {
			return nil
		}
		return hash
	}
}

func (p *Parser) parseSetExpression(left ast.Expression) ast.Expression {
	assignableStmt, ok := left.(ast.Assignable)
	if !ok {
		p.pushError("cannot assign to %T", left)
	}
	expr := &ast.SetExpression{Left: assignableStmt, Token: p.curToken}
	p.nextToken()
	expr.Right = p.parseExpression(LOWEST)
	return expr
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

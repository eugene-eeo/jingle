package parser

import (
	"fmt"
	"jingle/ast"
	"jingle/token"
	"strconv"
)

// ============================
// Expression parsing machinery
// ============================

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

// =====================
// Generic Pratt parsing
// =====================

func (p *Parser) parseExpressionOrInfixStatement(
	precedence int,
	allowStatement bool,
) ast.Node {
	// defer untrace(trace("parseExpression"))
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix, found := p.infixParseFns[p.peekToken.Type]
		// fmt.Println(p.peekToken.Type, infix, found)
		if !found {
			return leftExp
		}
		if infix == nil {
			// found && infix == nil means that there is
			// an infix statement to be expected.
			if !allowStatement {
				p.pushError("unexpected statement")
				return nil
			} else {
				p.nextToken()
				return p.parseInfixStatement(p.curToken, leftExp)
			}
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseExpression(prec int) ast.Expression {
	expr := p.parseExpressionOrInfixStatement(prec, false)
	if expr == nil {
		return nil
	}
	return expr.(ast.Expression)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	p.pushError("no prefix expr handler for %s", t)
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

func (p *Parser) parseInfixStatement(tok token.Token, left ast.Expression) ast.Statement {
	switch tok.Type {
	case token.ASSIGN:
		return p.parseSetStatement(left)
	}
	p.pushError("no infix stmt handler for %s", tok.Type)
	return nil
}

// ===================
// Specifc expressions
// ===================

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
	idents := []*ast.Identifier{}
	p.nextToken()
	for {
		if p.curTokenIs(token.RPAREN) {
			return idents
		}
		if !p.curTokenIs(token.IDENT) {
			p.pushError("expected IDENT, got=%s", p.curToken.Type)
			return nil
		}
		// current Token is an IDENT
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		idents = append(idents, ident)
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			continue
		}
		// terminate here
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
		return idents
	}
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
		hash.Pairs = append(hash.Pairs, ast.HashPair{Key: key, Value: val})
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

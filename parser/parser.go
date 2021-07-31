// Package parser implements a parser for the jingle language.
package parser

import (
	"jingle/ast"
	"jingle/scanner"
)

// Parser parses the given slice of tokens, and produces either a
// usable AST or an error.
type Parser struct {
	filename string
	tokens   []scanner.Token // list of tokens from the scanner
	consumed int             // number of tokens consumed.
	errors   []error         // parser errors encountered.
	// precedences
	prefixHandlers map[scanner.TokenType]prefixParseFn
	infixHandlers  map[scanner.TokenType]infixParseFn
	precedence     map[scanner.TokenType]int
}

func New(filename string, tokens []scanner.Token) *Parser {
	p := &Parser{
		filename: filename,
		tokens:   tokens,
		consumed: 0,
	}
	p.initExpressions()
	return p
}

func (p *Parser) MustParse() *ast.Program {
	prog, err := p.Parse()
	if err != nil {
		panic(err)
	}
	return prog
}

// Program is the main entry point into the parser.
func (p *Parser) Parse() (program *ast.Program, err error) {
	// Internally, we use panic(ParserError) to signal that
	// there has been a parsing error.
	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(ParserError); ok {
				err = pe
				return
			}
			panic(r)
		}
	}()
	program = p.parseProgram()
	return
}

// ===============
// Utility methods
// ===============

// peek returns the current token we have yet to consume
func (p *Parser) peek() scanner.Token { return p.tokens[p.consumed] }
func (p *Parser) isAtEnd() bool       { return p.peek().Type == scanner.TokenEOF }

// previous returns the previously consumed token
func (p *Parser) previous() scanner.Token { return p.tokens[p.consumed-1] }
func (p *Parser) consume() scanner.Token {
	p.consumed++
	return p.previous()
}

// match looks ahead at the token stream, and consumes 1
// token if it matches any of the given types.
func (p *Parser) match(types ...scanner.TokenType) bool {
	peekType := p.peek().Type
	for _, typ := range types {
		if peekType == typ {
			p.consume()
			return true
		}
	}
	return false
}

// expect is like match, but raises an error.
func (p *Parser) expect(t scanner.TokenType) {
	if !p.match(t) {
		p.error("expected %s, got %s instead", t, p.peek().Type)
	}
}

// ===============
// Actual Parsing!
// ===============

func (p *Parser) parseProgram() *ast.Program {
	prog := &ast.Program{}
	block := p.parseBlock(scanner.TokenEOF)
	prog.Statements = block.Statements
	return prog
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.peek().Type {
	case scanner.TokenLet:
		return p.parseLetStatement()
	case scanner.TokenFor:
		return p.parseForStatement()
	case scanner.TokenIf:
		return p.parseIfStatement()
	default:
		return &ast.ExpressionStatement{Expr: p.parseExpression()}
	}
}

func (p *Parser) parseBlock(terminal ...scanner.TokenType) *ast.Block {
	// block → ("sep")? blockStmts <terminal>
	// blockStmts → nothing | stmt ("sep" blockStmts)?
	lastHasSeparator := true
	block := &ast.Block{}
	block.Statements = []ast.Statement{}
	p.match(scanner.TokenSeparator) // initial whitespace -- ignore
	for !p.match(terminal...) {
		if !lastHasSeparator {
			p.error("expected newline or semicolon after statement")
		}
		block.Statements = append(block.Statements, p.parseStatement())
		lastHasSeparator = p.match(scanner.TokenSeparator)
	}
	block.Terminal = p.previous()
	return block
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	// let → "let" expr
	node := &ast.LetStatement{Token: p.consume()}
	node.Binding = p.parseExpression()
	if reason, ok := ast.Assignable(node.Binding, true); !ok {
		p.errorToken(reason.GetToken(),
			"cannot assign to %s", reason.Type())
	}
	return node
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	// if → "if" expr "then" block ("end" | ("else" block "end"))
	node := &ast.IfStatement{Token: p.consume()}
	node.Cond = p.parseExpression()
	p.expect(scanner.TokenThen)
	thenBlock := p.parseBlock(scanner.TokenEnd, scanner.TokenElse)
	node.Then = thenBlock
	if thenBlock.Terminal.Type == scanner.TokenElse {
		elseBlock := p.parseBlock(scanner.TokenEnd)
		node.Else = elseBlock
	}
	return node
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	// for → "for" expr "in" expr "do" stmts... "end"
	// note: expr has to be assignable
	node := &ast.ForStatement{Token: p.consume()}
	node.Binding = p.parseExpression()
	if reason, ok := ast.Assignable(node.Binding, true); !ok {
		p.errorToken(reason.GetToken(),
			"cannot assign to %s", reason.Type())
	}
	p.expect(scanner.TokenIn)
	node.Iterable = p.parseExpression()
	p.expect(scanner.TokenDo)
	node.Body = p.parseBlock(scanner.TokenEnd)
	return node
}

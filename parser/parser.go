// Package parser implements a parser for the jingle language.
package parser

import (
	"jingle/ast"
	"jingle/scanner"
)

var EOFToken = scanner.Token{Type: scanner.TokenEOF}

type Parser struct {
	filename string
	tokens   []scanner.Token // list of tokens from the scanner
	read     int             // number of tokens read
	errors   []ParserError   // parser errors encountered.
	// precedences
	prefixHandlers map[scanner.TokenType]prefixParseFn
	infixHandlers  map[scanner.TokenType]infixParseFn
	precedence     map[scanner.TokenType]int
}

func New(filename string, tokens []scanner.Token) *Parser {
	p := &Parser{
		filename: filename,
		tokens:   tokens,
		read:     0,
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
	// Internally, we use panic(...) but this method will wrap
	// these panics into an error for the public API.
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

func (p *Parser) last(i int) scanner.Token { return p.tokens[p.read-i] }
func (p *Parser) current() scanner.Token   { return p.lookAhead(0) }

func (p *Parser) lookAhead(distance int) scanner.Token {
	if p.read + distance > len(p.tokens) {
		return EOFToken
	}
	return p.tokens[p.read+distance]
}

func (p *Parser) expect(t scanner.TokenType) scanner.Token {
	if tokType := p.current().Type; tokType != t {
		// for the error message.
		p.consume()
		p.errorToken("expected %s, got %s instead", t, tokType)
	}
	return p.consume()
}

func (p *Parser) consume() scanner.Token {
	p.read++
	return p.last(1)
}

// isLookAhead looks ahead at the token stream, and returns
// true if the lookahead stream matches the given types.
func (p *Parser) lookAheadMatches(types ...scanner.TokenType) bool {
	for i, tokenType := range types {
		if p.lookAhead(i).Type != tokenType {
			return false
		}
	}
	return true
}

// match looks ahead at the token stream, and consumes
// tokens if all of them match the given types.
func (p *Parser) match(types ...scanner.TokenType) bool {
	if !p.lookAheadMatches(types...) {
		return false
	}
	// consume them
	for i := 0; i < len(types); i++ {
		p.consume()
	}
	return true
}

// matchAny consumes 1 token if it is _any_ of the given types.
func (p *Parser) matchAny(types ...scanner.TokenType) bool {
	curType := p.current().Type
	for _, tokType := range types {
		if curType == tokType {
			p.consume()
			return true
		}
	}
	return false
}

// ===============
// Actual Parsing!
// ===============

func (p *Parser) parseProgram() *ast.Program {
	prog := &ast.Program{}
	prog.Statements = []ast.Statement{}
	p.match(scanner.TokenSeparator) // initial whitespace -- ignore
	hasSep := true
	for !p.match(scanner.TokenEOF) {
		if !hasSep {
			p.errorToken("expected newline or semicolon, got %s instead", p.current().Type)
		}
		prog.Statements = append(prog.Statements, p.parseStatement())
		hasSep = p.matchAny(scanner.TokenSeparator, scanner.TokenSemicolon)
	}
	return prog
}

func (p *Parser) parseStatement() ast.Statement {
	switch {
	case p.match(scanner.TokenFor):
		return p.parseForStatement()
	}
	return &ast.ExpressionStatement{Expr: p.parseExpression()}
}

func (p *Parser) parseBlock(terminal scanner.TokenType) *ast.Block {
	// block → blockStmts <terminal>
	// blockStmts → nothing | stmt ("sep" blockStmts)?
	lastHasSeparator := true
	block := &ast.Block{}
	block.Statements = []ast.Statement{}
	for !p.match(terminal) {
		if !lastHasSeparator {
			p.errorToken("expected newline or semicolon, got %s instead", p.current().Type)
		}
		block.Statements = append(block.Statements, p.parseStatement())
		lastHasSeparator = p.matchAny(scanner.TokenSeparator, scanner.TokenSemicolon)
	}
	block.Terminal = p.last(1)
	return block
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	// for → "for" expr "in" expr "do" stmts... "end"
	// note: expr has to be assignable
	node := &ast.ForStatement{Token: p.last(1)}
	node.Binding = p.parseExpression()
	if (!ast.Assignable(node.Binding)) {
		p.errorToken("cannot assign to %s", node.Binding.Type())
	}
	p.expect(scanner.TokenIn)
	node.Iterable = p.parseExpression()
	p.expect(scanner.TokenDo)
	node.Body = p.parseBlock(scanner.TokenEnd)
	return node
}

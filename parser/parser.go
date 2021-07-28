// package parser implements a parser for the jingle language.
// the goals are for it to be easily extensible (so that it is
// easy to add new syntax elements). The parser is a recursive
// descent parser with arbitrary lookahead, and is lifted from
// the magpie language (https://github.com/munificent/magpie).
package parser

import (
	"jingle/ast"
	"jingle/lexer"
	"jingle/token"
)

var EmptyToken = token.Token{Type: token.ILLEGAL}

type Parser struct {
	lexer  *lexer.Lexer
	tokens []token.Token // list of tokens we've read so far.
	read   int           // number of tokens read
	// precedences
	prefixHandlers map[token.TokenType]prefixParseFn
	infixHandlers  map[token.TokenType]infixParseFn
	precedence     map[token.TokenType]int
}

func New(lx *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  lx,
		tokens: []token.Token{},
		read:   0,
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

func (p *Parser) last(i int) token.Token { return p.tokens[p.read-i] }
func (p *Parser) current() token.Token   { return p.lookAhead(0) }

func (p *Parser) lookAhead(distance int) token.Token {
	// [t1 ]
	//     ^read=1
	// lookAhead(0) => read 1
	// [t1 t2 t3 t4]
	//     ^read
	// lookAhead(1) => no reads
	size := len(p.tokens)
	for distance >= size-p.read {
		tok := p.lexer.NextToken()
		if err := p.lexer.Error(); err != nil {
			p.errorErr(err)
		}
		p.tokens = append(p.tokens, tok)
		size++
	}
	return p.tokens[p.read+distance]
}

func (p *Parser) expect(t token.TokenType) token.Token {
	if tokType := p.current().Type; tokType != t {
		p.consume()
		p.errorToken("expected %s, got %s instead", t, tokType)
	}
	return p.consume()
}

func (p *Parser) consume() token.Token {
	p.lookAhead(0)
	p.read++
	return p.last(1)
}

// isLookAhead looks ahead at the token stream, and returns
// true if the lookahead stream matches the given types.
func (p *Parser) lookAheadMatches(types ...token.TokenType) bool {
	for i, tokenType := range types {
		if p.lookAhead(i).Type != tokenType {
			return false
		}
	}
	return true
}

// match looks ahead at the token stream, and consumes
// tokens if all of them match the given types.
func (p *Parser) match(types ...token.TokenType) bool {
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
func (p *Parser) matchAny(types ...token.TokenType) bool {
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
	prog.Nodes = []ast.Node{}
	p.match(token.SEP) // initial whitespace -- ignore
	hasSep := true
	for !p.match(token.EOF) {
		if !hasSep {
			p.errorToken("expected newline or semicolon, got %s instead", p.current().Type)
		}
		prog.Nodes = append(prog.Nodes, p.parseStatement())
		hasSep = p.matchAny(token.SEP, token.SEMICOLON)
	}
	return prog
}

func (p *Parser) parseStatement() ast.Node {
	switch {
	case p.match(token.LET):
		return p.parseLetStatement()
	}
	return p.parseExpression()
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	letToken := p.last(1)
	p.expect(token.IDENT)
	left := p.parseIdentifierLiteral()
	p.expect(token.ASSIGN)
	right := p.parseExpression()
	return &ast.LetStatement{
		Token: letToken,
		Left:  left.(*ast.IdentifierLiteral),
		Right: right,
	}
}

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
	block := p.parseBlock(false, false, scanner.TokenEOF)
	prog.Statements = block.Statements
	return prog
}

func (p *Parser) parseStatement() ast.Statement {
	// stmt → let | for | if | exprstmt
	// exprstmt → expr
	switch p.peek().Type {
	case scanner.TokenLet:
		return p.parseLetStatement()
	case scanner.TokenFor:
		return p.parseForStatement()
	case scanner.TokenIf:
		return p.parseIfStatement()
	case scanner.TokenClass:
		return p.parseClassStatement()
	default:
		return &ast.ExpressionStatement{Expr: p.parseExpression()}
	}
}

func (p *Parser) parseBlock(
	isClass bool,
	isFunc bool,
	terminal ...scanner.TokenType,
) *ast.Block {
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
		var stmt ast.Statement
		switch p.peek().Type {
		case scanner.TokenDef:
			if !isClass {
				p.error("method declaration outside of class")
			}
			stmt = p.parseMethodDeclaration()
		case scanner.TokenReturn:
			if !isFunc {
				p.error("return statement outside of function")
			}
			stmt = p.parseReturnStatement()
		default:
			stmt = p.parseStatement()
		}
		block.Statements = append(block.Statements, stmt)
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
	thenBlock := p.parseBlock(false, false, scanner.TokenEnd, scanner.TokenElse)
	node.Then = thenBlock
	if thenBlock.Terminal.Type == scanner.TokenElse {
		elseBlock := p.parseBlock(false, false, scanner.TokenEnd)
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
	node.Body = p.parseBlock(false, false, scanner.TokenEnd)
	return node
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	// return → "return" expr
	node := &ast.ReturnStatement{Token: p.consume()}
	node.Expr = p.parseExpression()
	return node
}

func (p *Parser) parseClassStatement() *ast.ClassStatement {
	// class → "class" ident ( "<" expr )? classDecls "end"
	// classDecls → nothing | "sep" | (methodDecl | stmt) ( "sep" | "sep" classDecls )?
	class := &ast.ClassStatement{Token: p.consume()}
	p.expect(scanner.TokenIdent)
	class.Name = p.parseIdentifierLiteral().(*ast.IdentifierLiteral)
	if p.match(scanner.TokenLt) {
		// subclass
		class.SuperClass = p.parseExpression()
	}
	class.Body = p.parseBlock(true, false, scanner.TokenEnd)
	return class
}

func (p *Parser) parseMethodDeclaration() *ast.MethodDeclaration {
	// methodDecl → "def" methodName "(" params ")" block "end"
	// we're on top of a 'def' token.
	meth := &ast.MethodDeclaration{Token: p.previous()}
	meth.Name = p.parseMethodName()
	p.expect(scanner.TokenLParen)
	meth.Params = p.parseParams()
	meth.Body = p.parseBlock(false, true, scanner.TokenEnd)
	return meth
}

func (p *Parser) parseMethodName() string {
	// methodName → (ident
	//               | "[" "]" ("=")?
	//               | "+" | "-" | "*" | "/"
	//               | ">" | ">=" | "<" | "<=" | "==" | "!=")
	tok := p.consume()
	switch tok.Type {
	case scanner.TokenIdent:
		return tok.Value
	case scanner.TokenLBracket:
		p.expect(scanner.TokenRBracket)
		if p.match(scanner.TokenSet) {
			return "[]="
		}
		return "[]"
	case scanner.TokenPlus:
		return "+"
	case scanner.TokenMinus:
		return "-"
	case scanner.TokenMul:
		return "*"
	case scanner.TokenDiv:
		return "/"
	case scanner.TokenGt:
		return ">"
	case scanner.TokenGeq:
		return ">="
	case scanner.TokenLt:
		return "<"
	case scanner.TokenLeq:
		return "<="
	case scanner.TokenEq:
		return "=="
	case scanner.TokenNeq:
		return "!="
	}
	p.error("invalid method name")
	return "" // unreachable
}

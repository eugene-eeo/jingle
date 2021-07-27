package parser

import "jingle/ast"

type (
	// parse for a prefix expression: <OP><EXPR>
	prefixParseFn = func() ast.Node
	// parse for an infix expression: <EXPR><OP><EXPR>
	infixParseFn = func(ast.Node) ast.Node
)

func (p *Parser) initExpressions() {
}

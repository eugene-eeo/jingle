package ast

import (
	"bytes"
	"fmt"
	"jingle/token"
	"strconv"
)

// Node is a generic AST node. Furthermore a node can also be a
// statement node -- but all 'statements' are expressions.
type Node interface {
	Type() NodeType
	Start() token.Token
	End() token.Token
	String() string // used for debugging
}

type Statement interface {
	Node
	statementNode()
}

// ===========================
// Statements
// ===========================

type Program struct {
	StartToken token.Token
	EndToken   token.Token
	Nodes      []Node
}

func (node *Program) statementNode()     {}
func (node *Program) Type() NodeType     { return PROGRAM }
func (node *Program) Start() token.Token { return node.StartToken }
func (node *Program) End() token.Token   { return node.EndToken }
func (node *Program) String() string {
	var out bytes.Buffer
	last := len(node.Nodes) - 1
	for i, stmt := range node.Nodes {
		out.WriteString(stmt.String())
		if i != last {
			out.WriteString("\n")
		}
	}
	return out.String()
}

type LetStatement struct {
	StartToken token.Token
	EndToken   token.Token
	Left       *IdentifierLiteral
	Right      Node
}

func (node *LetStatement) statementNode() {}
func (node *LetStatement) Type() NodeType { return LET_STATEMENT }

// func (node *LetStatement) Type() NodeType     { return LET_STATEMENT }
func (node *LetStatement) Start() token.Token { return node.StartToken }
func (node *LetStatement) End() token.Token   { return node.EndToken }
func (node *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString("let ")
	out.WriteString(node.Left.String())
	out.WriteString(" = ")
	out.WriteString(node.Right.String())
	return out.String()
}

// ===========================
// Expressions
// ===========================

// ===========================
// Literals
// ===========================

type NullLiteral struct {
	Token token.Token
}

func (node *NullLiteral) Type() NodeType     { return NULL_LITERAL }
func (node *NullLiteral) Start() token.Token { return node.Token }
func (node *NullLiteral) End() token.Token   { return node.Token }
func (node *NullLiteral) String() string     { return "null" }

type BoolLiteral struct {
	Token token.Token
	Value bool
}

func (node *BoolLiteral) Type() NodeType     { return BOOL_LITERAL }
func (node *BoolLiteral) Start() token.Token { return node.Token }
func (node *BoolLiteral) End() token.Token   { return node.Token }
func (node *BoolLiteral) String() string     { return "null" }

type IdentifierLiteral struct {
	Token token.Token
}

func (node *IdentifierLiteral) Type() NodeType     { return IDENTIFIER_LITERAL }
func (node *IdentifierLiteral) Start() token.Token { return node.Token }
func (node *IdentifierLiteral) End() token.Token   { return node.Token }
func (node *IdentifierLiteral) String() string     { return node.Token.Literal }
func (node *IdentifierLiteral) Name() string {
	return node.Token.Literal
}

type NumberLiteral struct {
	Token token.Token
	Value float64
}

func (node *NumberLiteral) Type() NodeType     { return NUMBER_LITERAL }
func (node *NumberLiteral) Start() token.Token { return node.Token }
func (node *NumberLiteral) End() token.Token   { return node.Token }
func (node *NumberLiteral) String() string     { return strconv.FormatFloat(node.Value, 'G', -1, 64) }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (node *StringLiteral) Type() NodeType     { return STRING_LITERAL }
func (node *StringLiteral) Start() token.Token { return node.Token }
func (node *StringLiteral) End() token.Token   { return node.Token }
func (node *StringLiteral) String() string     { return fmt.Sprintf("%q", node.Value) }

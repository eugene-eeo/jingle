package ast

import (
	"bytes"
	"fmt"
	"jingle/token"
)

// Node is a generic AST node. Furthermore a node can also be a
// statement node -- but all 'statements' are expressions.
type Node interface {
	Type() NodeType
	String() string // used for debugging
}

type Statement interface {
	Node
	statementNode()
}

// ===========================
// 'Statements'
// ===========================

type Program struct {
	Nodes []Node
}

func (node *Program) statementNode() {}
func (node *Program) Type() NodeType { return PROGRAM }
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
	Token token.Token
	Left  *IdentifierLiteral
	Right Node
}

func (node *LetStatement) statementNode() {}
func (node *LetStatement) Type() NodeType { return LET_STATEMENT }
func (node *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(node.Token.Literal + " ")
	out.WriteString(node.Left.String())
	out.WriteString(" = ")
	out.WriteString(node.Right.String())
	return out.String()
}

// ===========================
// Expressions
// ===========================

type InfixExpression struct {
	Token token.Token // the <op> token
	Op    string
	Left  Node
	Right Node
}

func (node *InfixExpression) Type() NodeType { return INFIX_EXPRESSION }
func (node *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Left.String())
	out.WriteString(" " + node.Token.Literal + " ")
	out.WriteString(node.Right.String())
	out.WriteString(")")
	return out.String()
}

type AssignmentExpression struct {
	Token token.Token // the '=' token
	Left  Node
	Right Node
}

func (node *AssignmentExpression) Type() NodeType { return INFIX_EXPRESSION }
func (node *AssignmentExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Left.String())
	out.WriteString(" " + node.Token.Literal + " ")
	out.WriteString(node.Right.String())
	out.WriteString(")")
	return out.String()
}

// ===========================
// Literals
// ===========================

type NullLiteral struct {
	Token token.Token
}

func (node *NullLiteral) Type() NodeType { return NULL_LITERAL }
func (node *NullLiteral) String() string { return "null" }

type BoolLiteral struct {
	Token token.Token
	Value bool
}

func (node *BoolLiteral) Type() NodeType { return BOOL_LITERAL }
func (node *BoolLiteral) String() string { return "null" }

type IdentifierLiteral struct {
	Token token.Token
}

func (node *IdentifierLiteral) Type() NodeType { return IDENTIFIER_LITERAL }
func (node *IdentifierLiteral) String() string { return node.Token.Literal }
func (node *IdentifierLiteral) Name() string {
	return node.Token.Literal
}

type NumberLiteral struct {
	Token token.Token
	Value float64
}

func (node *NumberLiteral) Type() NodeType { return NUMBER_LITERAL }
func (node *NumberLiteral) String() string { return node.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (node *StringLiteral) Type() NodeType { return STRING_LITERAL }
func (node *StringLiteral) String() string { return fmt.Sprintf("%q", node.Value) }

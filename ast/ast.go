package ast

import (
	"bytes"
	"fmt"
	"jingle/token"
	"strings"
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

type OrExpression struct {
	// The reason we need these is to implement short-circuiting
	// expressions -- it is easier for the evaluator to do this.
	Token token.Token // the `||` token
	Op    string
	Left  Node
	Right Node
}

func (node *OrExpression) Type() NodeType { return OR_EXPRESSION }
func (node *OrExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Left.String())
	out.WriteString(" " + node.Token.Literal + " ")
	out.WriteString(node.Right.String())
	out.WriteString(")")
	return out.String()
}

type AndExpression struct {
	Token token.Token // the `&&` token
	Op    string
	Left  Node
	Right Node
}

func (node *AndExpression) Type() NodeType { return AND_EXPRESSION }
func (node *AndExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Left.String())
	out.WriteString(" " + node.Token.Literal + " ")
	out.WriteString(node.Right.String())
	out.WriteString(")")
	return out.String()
}

type BlockExpression struct {
	Nodes []Node
}

func (node BlockExpression) Type() NodeType { return BLOCK_EXPRESSION }
func (node BlockExpression) String() string {
	var out bytes.Buffer
	out.WriteString(" ")
	for _, stmt := range node.Nodes {
		out.WriteString(stmt.String())
		out.WriteString("; ")
	}
	out.WriteString("end")
	return out.String()
}

type AttrExpression struct {
	Token token.Token // the '.' token
	Left  Node
	Right *IdentifierLiteral
}

func (node AttrExpression) Type() NodeType { return ATTR_EXPRESSION }
func (node AttrExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Left.String())
	out.WriteString(")")
	out.WriteString(node.Token.Literal)
	out.WriteString(node.Right.String())
	return out.String()
}

// ===========================
// Literals
// ===========================

type NullLiteral struct {
	Token token.Token // the 'null' token
}

func (node *NullLiteral) Type() NodeType { return NULL_LITERAL }
func (node *NullLiteral) String() string { return node.Token.Literal }

type BooleanLiteral struct {
	Token token.Token // true/false token
	Value bool
}

func (node *BooleanLiteral) Type() NodeType { return BOOL_LITERAL }
func (node *BooleanLiteral) String() string { return node.Token.Literal }

type IdentifierLiteral struct {
	Token token.Token // ident token
}

func (node *IdentifierLiteral) Type() NodeType { return IDENTIFIER_LITERAL }
func (node *IdentifierLiteral) String() string { return node.Token.Literal }
func (node *IdentifierLiteral) Name() string {
	return node.Token.Literal
}

type NumberLiteral struct {
	Token token.Token // number token
	Value float64
}

func (node *NumberLiteral) Type() NodeType { return NUMBER_LITERAL }
func (node *NumberLiteral) String() string { return node.Token.Literal }

type StringLiteral struct {
	Token token.Token // string token
	Value string
}

func (node *StringLiteral) Type() NodeType { return STRING_LITERAL }
func (node *StringLiteral) String() string { return fmt.Sprintf("%q", node.Value) }

type FunctionLiteral struct {
	Token  token.Token // the 'fn' token
	Params []*IdentifierLiteral
	Body   *BlockExpression
}

func (node *FunctionLiteral) Type() NodeType { return FUNCTION_LITERAL }
func (node *FunctionLiteral) String() string {
	var buf bytes.Buffer
	params := []string{}
	for _, param := range node.Params {
		params = append(params, param.String())
	}

	buf.WriteString(node.Token.Literal)
	buf.WriteString("(")
	buf.WriteString(strings.Join(params, ", "))
	buf.WriteString(")")
	buf.WriteString(node.Body.String())
	return buf.String()
}

package ast

import (
	"bytes"
	"fmt"
	"jingle/scanner"
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
	Token    scanner.Token
	Bindings []Node
}

func (node *LetStatement) statementNode() {}
func (node *LetStatement) Type() NodeType { return LET_STATEMENT }
func (node *LetStatement) String() string {
	var out bytes.Buffer
	bindings := []string{}
	for _, b := range node.Bindings {
		bindings = append(bindings, b.String())
	}
	out.WriteString(node.Token.Value + " ")
	out.WriteString(strings.Join(bindings, ", "))
	return out.String()
}

type ForStatement struct {
	Token    scanner.Token // the 'for' token
	Binding  *IdentifierLiteral
	Iterable Node
	Body     *Block
}

func (node ForStatement) Type() NodeType { return FOR_STATEMENT }
func (node ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString(node.Token.Value)
	out.WriteString(" ")
	out.WriteString(node.Binding.String())
	out.WriteString(" in ")
	out.WriteString(node.Iterable.String())
	out.WriteString(" do")
	out.WriteString(node.Body.String())
	return out.String()
}

type Block struct {
	Nodes []Node
}

func (node Block) Type() NodeType { return BLOCK_EXPRESSION }
func (node Block) String() string {
	var out bytes.Buffer
	out.WriteString(" ")
	for _, stmt := range node.Nodes {
		out.WriteString(stmt.String())
		out.WriteString("; ")
	}
	out.WriteString("end")
	return out.String()
}

// ===========================
// Expressions
// ===========================

type InfixExpression struct {
	Token scanner.Token // the <op> token
	Op    string
	Left  Node
	Right Node
}

func (node *InfixExpression) Type() NodeType { return INFIX_EXPRESSION }
func (node *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Left.String())
	out.WriteString(" " + node.Token.Value + " ")
	out.WriteString(node.Right.String())
	out.WriteString(")")
	return out.String()
}

type AssignmentExpression struct {
	Token scanner.Token // the '=' token
	Left  Node
	Right Node
}

func (node *AssignmentExpression) Type() NodeType { return ASSIGNMENT_EXPRESSION }
func (node *AssignmentExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Left.String())
	out.WriteString(" " + node.Token.Value + " ")
	out.WriteString(node.Right.String())
	out.WriteString(")")
	return out.String()
}

type OrExpression struct {
	// The reason we need these is to implement short-circuiting
	// expressions -- it is easier for the evaluator to do this.
	Token scanner.Token // the `||` token
	Op    string
	Left  Node
	Right Node
}

func (node *OrExpression) Type() NodeType { return OR_EXPRESSION }
func (node *OrExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Left.String())
	out.WriteString(" " + node.Token.Value + " ")
	out.WriteString(node.Right.String())
	out.WriteString(")")
	return out.String()
}

type AndExpression struct {
	Token scanner.Token // the `&&` token
	Op    string
	Left  Node
	Right Node
}

func (node *AndExpression) Type() NodeType { return AND_EXPRESSION }
func (node *AndExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Left.String())
	out.WriteString(" " + node.Token.Value + " ")
	out.WriteString(node.Right.String())
	out.WriteString(")")
	return out.String()
}

type AttrExpression struct {
	Token scanner.Token // the '.' token
	Left  Node
	Right *IdentifierLiteral
}

func (node AttrExpression) Type() NodeType { return ATTR_EXPRESSION }
func (node AttrExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Left.String())
	out.WriteString(")")
	out.WriteString(node.Token.Value)
	out.WriteString(node.Right.String())
	return out.String()
}

// ===========================
// Literals
// ===========================

type NilLiteral struct {
	Token scanner.Token // the 'nil' token
}

func (node *NilLiteral) Type() NodeType { return NIL_LITERAL }
func (node *NilLiteral) String() string { return node.Token.Value }

type BooleanLiteral struct {
	Token scanner.Token // true/false token
	Value bool
}

func (node *BooleanLiteral) Type() NodeType { return BOOLEAN_LITERAL }
func (node *BooleanLiteral) String() string { return node.Token.Value }

type IdentifierLiteral struct {
	Token scanner.Token // ident token
}

func (node *IdentifierLiteral) Type() NodeType { return IDENTIFIER_LITERAL }
func (node *IdentifierLiteral) String() string { return node.Token.Value }
func (node *IdentifierLiteral) Name() string {
	return node.Token.Value
}

type NumberLiteral struct {
	Token scanner.Token // number token
	Value float64
}

func (node *NumberLiteral) Type() NodeType { return NUMBER_LITERAL }
func (node *NumberLiteral) String() string { return node.Token.Value }

type StringLiteral struct {
	Token scanner.Token // string token
	Value string
}

func (node *StringLiteral) Type() NodeType { return STRING_LITERAL }
func (node *StringLiteral) String() string { return fmt.Sprintf("%q", node.Value) }

type FunctionLiteral struct {
	Token  scanner.Token // the 'fn' token
	Params []*IdentifierLiteral
	Body   *Block
}

func (node *FunctionLiteral) Type() NodeType { return FUNCTION_LITERAL }
func (node *FunctionLiteral) String() string {
	var buf bytes.Buffer
	params := []string{}
	for _, param := range node.Params {
		params = append(params, param.String())
	}

	buf.WriteString(node.Token.Value)
	buf.WriteString("(")
	buf.WriteString(strings.Join(params, ", "))
	buf.WriteString(")")
	buf.WriteString(node.Body.String())
	return buf.String()
}

type ArrayLiteral struct {
	Token scanner.Token // the 'fn' token
	Elems []Node
}

func (node *ArrayLiteral) Type() NodeType { return ARRAY_LITERAL }
func (node *ArrayLiteral) String() string {
	var buf bytes.Buffer
	elems := []string{}
	for _, elem := range node.Elems {
		elems = append(elems, elem.String())
	}

	buf.WriteString(node.Token.Value)
	buf.WriteString(strings.Join(elems, ", "))
	buf.WriteString("]")
	return buf.String()
}

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
	GetToken() scanner.Token
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// ===========================
// Statements
// ===========================

type Program struct {
	Token      scanner.Token
	Statements []Statement
}

func (node *Program) statementNode()          {}
func (node *Program) Type() NodeType          { return PROGRAM }
func (node *Program) GetToken() scanner.Token { return node.Token }
func (node *Program) String() string {
	var out bytes.Buffer
	for _, stmt := range node.Statements {
		out.WriteString(stmt.String())
	}
	return out.String()
}

type ExpressionStatement struct{ Expr Expression }

func (node *ExpressionStatement) statementNode()          {}
func (node *ExpressionStatement) Type() NodeType          { return EXPRESSION_STATEMENT }
func (node *ExpressionStatement) GetToken() scanner.Token { return node.Expr.GetToken() }
func (node *ExpressionStatement) String() string {
	return node.Expr.String() + ";"
}

type LetStatement struct {
	Token   scanner.Token // the 'let' token
	Binding Expression
}

func (node *LetStatement) statementNode()          {}
func (node *LetStatement) Type() NodeType          { return LET_STATEMENT }
func (node *LetStatement) GetToken() scanner.Token { return node.Token }
func (node *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(node.Token.Value)
	out.WriteString(" ")
	out.WriteString(node.Binding.String())
	return out.String()
}

type ForStatement struct {
	Token    scanner.Token // the 'for' token
	Binding  Expression
	Iterable Expression
	Body     *Block
}

func (node *ForStatement) statementNode()          {}
func (node *ForStatement) Type() NodeType          { return FOR_STATEMENT }
func (node *ForStatement) GetToken() scanner.Token { return node.Token }
func (node *ForStatement) String() string {
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

type WhileStatement struct {
	Token     scanner.Token // the 'while' token
	Condition Expression
	Body      *Block
}

func (node *WhileStatement) statementNode()          {}
func (node *WhileStatement) Type() NodeType          { return WHILE_STATEMENT }
func (node *WhileStatement) GetToken() scanner.Token { return node.Token }
func (node *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString(node.Token.Value)
	out.WriteString(" ")
	out.WriteString(node.Condition.String())
	out.WriteString(" ")
	out.WriteString(node.Body.String())
	return out.String()
}

type Block struct {
	Statements []Statement
	Terminal   scanner.Token
}

func (node *Block) statementNode()          {}
func (node *Block) Type() NodeType          { return BLOCK_STATEMENT }
func (node *Block) GetToken() scanner.Token { return node.Terminal }
func (node *Block) String() string {
	var out bytes.Buffer
	out.WriteString(" ")
	for _, stmt := range node.Statements {
		out.WriteString(stmt.String())
	}
	out.WriteString(" ")
	out.WriteString(node.Terminal.Value)
	return out.String()
}

type IfStatement struct {
	Token scanner.Token // the <if> token
	Cond  Expression
	Then  Node
	Else  Node
}

func (node *IfStatement) statementNode()          {}
func (node *IfStatement) Type() NodeType          { return IF_STATEMENT }
func (node *IfStatement) GetToken() scanner.Token { return node.Token }
func (node *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString(node.Token.Value)
	out.WriteString(" ")
	out.WriteString(node.Cond.String())
	if node.Else != nil {
		out.WriteString(" else ")
		out.WriteString(node.Else.String())
	}
	return out.String()
}

type MethodName struct {
	Token scanner.Token
	Name  string
}

type MethodDeclaration struct {
	Token      scanner.Token // the 'def' token
	MethodName MethodName
	Params     []*IdentifierLiteral
	Body       *Block
}

func (node *MethodDeclaration) statementNode()          {}
func (node *MethodDeclaration) Type() NodeType          { return METHOD_DECLARATION }
func (node *MethodDeclaration) GetToken() scanner.Token { return node.Token }
func (node *MethodDeclaration) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, param := range node.Params {
		params = append(params, param.String())
	}
	out.WriteString(node.Token.Value)
	out.WriteString(" ")
	out.WriteString(node.MethodName.Name)
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString(node.Body.String())
	return out.String()
}

type ClassStatement struct {
	Token      scanner.Token // the 'class' token
	Name       *IdentifierLiteral
	SuperClass Expression
	Body       *Block
}

func (node *ClassStatement) statementNode()          {}
func (node *ClassStatement) Type() NodeType          { return CLASS_STATEMENT }
func (node *ClassStatement) GetToken() scanner.Token { return node.Token }
func (node *ClassStatement) String() string {
	var out bytes.Buffer
	out.WriteString(node.Token.Value)
	out.WriteString(" ")
	out.WriteString(node.Name.String())
	if node.SuperClass != nil {
		out.WriteString(" < ")
		out.WriteString(node.SuperClass.String())
	}
	out.WriteString(node.Body.String())
	return out.String()
}

type ReturnStatement struct {
	Token scanner.Token // the 'return' token
	Expr  Expression
}

func (node *ReturnStatement) statementNode()          {}
func (node *ReturnStatement) Type() NodeType          { return RETURN_STATEMENT }
func (node *ReturnStatement) GetToken() scanner.Token { return node.Token }
func (node *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(node.Token.Value)
	out.WriteString(" ")
	out.WriteString(node.Expr.String())
	return out.String()
}

// ===========================
// Expressions
// ===========================

type PrefixExpression struct {
	Token scanner.Token // the <op> token
	Op    string
	Expr  Expression
}

func (node *PrefixExpression) expressionNode()         {}
func (node *PrefixExpression) Type() NodeType          { return PREFIX_EXPRESSION }
func (node *PrefixExpression) GetToken() scanner.Token { return node.Token }
func (node *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Token.Value)
	out.WriteString(node.Expr.String())
	out.WriteString(")")
	return out.String()
}

type InfixExpression struct {
	Token scanner.Token // the <op> token
	Op    string
	Left  Expression
	Right Expression
}

func (node *InfixExpression) expressionNode()         {}
func (node *InfixExpression) Type() NodeType          { return INFIX_EXPRESSION }
func (node *InfixExpression) GetToken() scanner.Token { return node.Token }
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
	Left  Expression
	Right Expression
}

func (node *AssignmentExpression) expressionNode()         {}
func (node *AssignmentExpression) Type() NodeType          { return ASSIGNMENT_EXPRESSION }
func (node *AssignmentExpression) GetToken() scanner.Token { return node.Token }
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
	Left  Expression
	Right Expression
}

func (node *OrExpression) expressionNode()         {}
func (node *OrExpression) Type() NodeType          { return OR_EXPRESSION }
func (node *OrExpression) GetToken() scanner.Token { return node.Token }
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
	Left  Expression
	Right Expression
}

func (node *AndExpression) expressionNode()         {}
func (node *AndExpression) Type() NodeType          { return AND_EXPRESSION }
func (node *AndExpression) GetToken() scanner.Token { return node.Token }
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
	Token  scanner.Token // the '.' token
	Target Expression
	Name   *IdentifierLiteral
}

func (node *AttrExpression) expressionNode()         {}
func (node *AttrExpression) Type() NodeType          { return ATTR_EXPRESSION }
func (node *AttrExpression) GetToken() scanner.Token { return node.Token }
func (node *AttrExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Target.String())
	out.WriteString(")")
	out.WriteString(node.Token.Value)
	out.WriteString(node.Name.String())
	return out.String()
}

type IndexExpression struct {
	Token  scanner.Token // the '[' token
	Target Expression
	Args   []Expression
}

func (node *IndexExpression) expressionNode()         {}
func (node *IndexExpression) Type() NodeType          { return INDEX_EXPRESSION }
func (node *IndexExpression) GetToken() scanner.Token { return node.Token }
func (node *IndexExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, arg := range node.Args {
		args = append(args, arg.String())
	}
	out.WriteString("(")
	out.WriteString(node.Target.String())
	out.WriteString(")")
	out.WriteString(node.Token.Value)
	out.WriteString(strings.Join(args, ","))
	out.WriteString("]")
	return out.String()
}

type CallExpression struct {
	Token  scanner.Token // the '(' token
	Target Expression
	Args   []Expression
}

func (node *CallExpression) expressionNode()         {}
func (node *CallExpression) Type() NodeType          { return CALL_EXPRESSION }
func (node *CallExpression) GetToken() scanner.Token { return node.Token }
func (node *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, arg := range node.Args {
		args = append(args, arg.String())
	}
	out.WriteString(node.Target.String())
	out.WriteString(node.Token.Value)
	out.WriteString(strings.Join(args, ","))
	out.WriteString(")")
	return out.String()
}

type IfElseExpression struct {
	Token scanner.Token // the 'if' token
	Cond  Expression
	Then  Expression
	Else  Expression
}

func (node *IfElseExpression) expressionNode()         {}
func (node *IfElseExpression) Type() NodeType          { return IF_ELSE_EXPRESSION }
func (node *IfElseExpression) GetToken() scanner.Token { return node.Token }
func (node *IfElseExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(node.Then.String())
	out.WriteString(" ")
	out.WriteString(node.Token.Value)
	out.WriteString(" ")
	out.WriteString(node.Cond.String())
	if node.Else != nil {
		out.WriteString(" else ")
		out.WriteString(node.Else.String())
	}
	out.WriteString(")")
	return out.String()
}

// ===========================
// Literals
// ===========================

type NilLiteral struct {
	Token scanner.Token // the 'nil' token
}

func (node *NilLiteral) expressionNode()         {}
func (node *NilLiteral) Type() NodeType          { return NIL_LITERAL }
func (node *NilLiteral) GetToken() scanner.Token { return node.Token }
func (node *NilLiteral) String() string          { return node.Token.Value }

type BooleanLiteral struct {
	Token scanner.Token // true/false token
	Value bool
}

func (node *BooleanLiteral) expressionNode()         {}
func (node *BooleanLiteral) Type() NodeType          { return BOOLEAN_LITERAL }
func (node *BooleanLiteral) GetToken() scanner.Token { return node.Token }
func (node *BooleanLiteral) String() string          { return node.Token.Value }

type IdentifierLiteral struct {
	Token scanner.Token // ident token
}

func (node *IdentifierLiteral) expressionNode()         {}
func (node *IdentifierLiteral) Type() NodeType          { return IDENTIFIER_LITERAL }
func (node *IdentifierLiteral) GetToken() scanner.Token { return node.Token }
func (node *IdentifierLiteral) String() string          { return node.Token.Value }
func (node *IdentifierLiteral) Name() string {
	return node.Token.Value
}

type NumberLiteral struct {
	Token scanner.Token // number token
	Value float64
}

func (node *NumberLiteral) expressionNode()         {}
func (node *NumberLiteral) Type() NodeType          { return NUMBER_LITERAL }
func (node *NumberLiteral) GetToken() scanner.Token { return node.Token }
func (node *NumberLiteral) String() string          { return node.Token.Value }

type StringLiteral struct {
	Token scanner.Token // string token
	Value string
}

func (node *StringLiteral) expressionNode()         {}
func (node *StringLiteral) Type() NodeType          { return STRING_LITERAL }
func (node *StringLiteral) GetToken() scanner.Token { return node.Token }
func (node *StringLiteral) String() string          { return fmt.Sprintf("%q", node.Value) }

type FunctionLiteral struct {
	Token  scanner.Token // the 'fn' token
	Params []*IdentifierLiteral
	Body   *Block
}

func (node *FunctionLiteral) expressionNode()         {}
func (node *FunctionLiteral) Type() NodeType          { return FUNCTION_LITERAL }
func (node *FunctionLiteral) GetToken() scanner.Token { return node.Token }
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
	Elems []Expression
}

func (node *ArrayLiteral) expressionNode()         {}
func (node *ArrayLiteral) Type() NodeType          { return ARRAY_LITERAL }
func (node *ArrayLiteral) GetToken() scanner.Token { return node.Token }
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

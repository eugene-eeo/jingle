package ast

//go:generate stringer -type=NodeType

type NodeType uint

const (
	_ = NodeType(iota)
	PROGRAM
	// Statements
	LET_STATEMENT
	FOR_STATEMENT

	// Expressions
	INFIX_EXPRESSION
	ASSIGNMENT_EXPRESSION
	OR_EXPRESSION
	AND_EXPRESSION
	BLOCK_EXPRESSION
	ATTR_EXPRESSION

	// Literals
	NIL_LITERAL
	BOOLEAN_LITERAL
	IDENTIFIER_LITERAL
	NUMBER_LITERAL
	STRING_LITERAL
	FUNCTION_LITERAL
	ARRAY_LITERAL
)

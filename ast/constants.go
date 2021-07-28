package ast

type NodeType uint

const (
	_ = NodeType(iota)
	PROGRAM
	// Statements
	LET_STATEMENT

	// Expressions
	INFIX_EXPRESSION
	ASSIGNMENT_EXPRESSION
	OR_EXPRESSION
	AND_EXPRESSION

	// Literals
	NULL_LITERAL
	BOOL_LITERAL
	IDENTIFIER_LITERAL
	NUMBER_LITERAL
	STRING_LITERAL
)

func NodeTypeAsString(t NodeType) string {
	switch t {
	case PROGRAM:
		return "PROGRAM"
	case LET_STATEMENT:
		return "LET_STATEMENT"
	case INFIX_EXPRESSION:
		return "INFIX_EXPRESSION"
	case ASSIGNMENT_EXPRESSION:
		return "ASSIGNMENT_EXPRESSION"
	case OR_EXPRESSION:
		return "OR_EXPRESSION"
	case AND_EXPRESSION:
		return "AND_EXPRESSION"
	case NULL_LITERAL:
		return "NULL_LITERAL"
	case BOOL_LITERAL:
		return "BOOL_LITERAL"
	case IDENTIFIER_LITERAL:
		return "IDENTIFIER_LITERAL"
	case NUMBER_LITERAL:
		return "NUMBER_LITERAL"
	case STRING_LITERAL:
		return "STRING_LITERAL"
	}
	return "<Unknown>"
}

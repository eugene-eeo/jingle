package ast

// Assignable returns if the given expr is assignable.
//    x = b
//    x.a = b
func Assignable(node Node, isDeclaration bool) (Node, bool) {
	switch node := node.(type) {
	case *AssignmentExpression:
		return Assignable(node.Left, isDeclaration)
	case *AttrExpression:
		return node, !isDeclaration
	case *IndexExpression:
		return node, !isDeclaration
	case *IdentifierLiteral:
		return node, true
	}
	return node, false
}

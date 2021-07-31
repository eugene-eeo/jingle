package ast

// Assignable returns if the given expr is assignable
//    x = b
//    x.a = b
//    [x, b] = [a, 1]
//    ^--- array literals are assignable.
func Assignable(node Node, isDeclaration bool) (Node, bool) {
	switch node := node.(type) {
	case *AssignmentExpression:
		return Assignable(node.Left, isDeclaration)
	case *AttrExpression:
		return node, !isDeclaration
	case *IdentifierLiteral:
		return node, true
	case *ArrayLiteral:
		for _, x := range node.Elems {
			if reason, ok := Assignable(x, isDeclaration); !ok {
				return reason, ok
			}
		}
		return node, true
	}
	return node, false
}

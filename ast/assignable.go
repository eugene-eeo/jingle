package ast

// Assignable returns if the given node is assignable
//    x = b
//    x.a = b
//    [x, 1] = [a, 1]
//    ^--- array literals are assignable.
func Assignable(node Node) bool {
	switch node := node.(type) {
	case *AssignmentExpression:
		return true
	case *AttrExpression:
		return true
	case *IdentifierLiteral:
		return true
	case *ArrayLiteral:
		for _, x := range node.Elems {
			if !Assignable(x) {
				return false
			}
		}
		return true
	}
	return false
}

package ast

// Assignable returns if the given expr is assignable
//    x = b
//    x.a = b
//    [x, b] = [a, 1]
//    ^--- array literals are assignable.
func Assignable(node Node) bool {
	switch node := node.(type) {
	case *AssignmentExpression:
		// We can do a recursive check here, but for performance we
		// assume that whatever code parses AssignmentExpressions
		// already checks this. Otherwise we will end up calling this
		// O(n^2) times.
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

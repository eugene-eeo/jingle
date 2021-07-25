package evaluator

import "jingle/object"

// Where we define how the language operators work.
// For example, what happens when 1 + 1.

// =================
// Generic Operators
// =================

// evalIndexExpression returns left[index]
func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(
			left.(*object.Array),
			index.(*object.Integer),
		)
	case left.Type() == object.STRING_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalStringIndexExpression(
			left.(*object.String),
			index.(*object.Integer),
		)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left.(*object.Hash), index)
	default:
		return newError("operator not supported: %s[%s]", left.Type(), index.Type())
	}
}

// evalInfixExpression returns left `operator` right
func evalInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	switch {
	case operator == "is":
		return nativeBoolToBooleanObject(left == right)
	case operator == "==" && left == right:
		// Fast case: pointer equality.
		return TRUE
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(
			operator,
			left.(*object.Integer),
			right.(*object.Integer),
		)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(
			operator,
			left.(*object.String),
			right.(*object.String),
		)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	}
	return newError("unknown operator: %s %s %s",
		left.Type(),
		operator,
		right.Type(),
	)
}

// evalSetIndex returns the result of setting an index
// to a particular value, i.e. obj[idx] = val
func evalSetIndex(obj, idx, val object.Object) object.Object {
	switch {
	case obj.Type() == object.ARRAY_OBJ && idx.Type() == object.INTEGER_OBJ:
		obj := obj.(*object.Array)
		idx := idx.(*object.Integer)
		obj.Elements[idx.Value] = val
	case obj.Type() == object.HASH_OBJ:
		obj := obj.(*object.Hash)
		hashableIdx, ok := idx.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", idx.Type())
		}
		obj.Pairs.Set(hashableIdx, val)
	default:
		return newError("invalid operation: %s[%s] = %s",
			obj.Type(),
			idx.Type(),
			val.Type(),
		)
	}
	return NULL
}

// ======================
// Type-specific dispatch
// ======================

func evalIntegerInfixExpression(operator string, left, right *object.Integer) object.Object {
	a := left.Value
	b := right.Value
	switch operator {
	case "+":
		return &object.Integer{Value: a + b}
	case "-":
		return &object.Integer{Value: a - b}
	case "/":
		return &object.Integer{Value: a / b}
	case "*":
		return &object.Integer{Value: a * b}
	case "==":
		return nativeBoolToBooleanObject(a == b)
	case "!=":
		return nativeBoolToBooleanObject(a != b)
	case "<":
		return nativeBoolToBooleanObject(a < b)
	case ">":
		return nativeBoolToBooleanObject(a > b)
	case "<=":
		return nativeBoolToBooleanObject(a <= b)
	case ">=":
		return nativeBoolToBooleanObject(a >= b)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right *object.String) object.Object {
	a := left.Value
	b := right.Value
	switch operator {
	case "+":
		return &object.String{Value: a + b}
	// case "*":
	// 	return &object.Integer{Value: a * b}
	case "==":
		return nativeBoolToBooleanObject(a == b)
	case "!=":
		return nativeBoolToBooleanObject(a != b)
	case "<":
		return nativeBoolToBooleanObject(a < b)
	case ">":
		return nativeBoolToBooleanObject(a > b)
	case "<=":
		return nativeBoolToBooleanObject(a <= b)
	case ">=":
		return nativeBoolToBooleanObject(a >= b)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringIndexExpression(
	str *object.String,
	index *object.Integer,
) object.Object {
	idx := index.Value
	max := int64(len(str.Value)) - 1
	if idx < 0 || idx > max {
		return NULL
	}
	return &object.String{Value: string(str.Value[idx])}
}

func evalArrayIndexExpression(array *object.Array, index *object.Integer) object.Object {
	idx := index.Value
	max := int64(len(array.Elements)) - 1
	if idx < 0 || idx > max {
		return NULL
	}
	return array.Elements[idx]
}

func evalHashIndexExpression(hash *object.Hash, key object.Object) object.Object {
	hashKey, ok := key.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", key.Type())
	}
	value, ok := hash.Pairs.Get(hashKey)
	if !ok {
		return NULL
	}
	return value
}

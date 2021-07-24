package evaluator

import (
	"fmt"
	"jingle/ast"
	"jingle/object"
)

var (
	// Constants -- only one of each in any execution.
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// ==== Statements ====
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		value := Eval(node.ReturnValue, env)
		if isError(value) {
			return value
		}
		return &object.ReturnValue{Value: value}

	case *ast.LetStatement:
		// Warning: this returns <nil> explicitly: remember
		// to handle it!
		value := Eval(node.Value, env)
		if isError(value) {
			return value
		}
		env.Let(node.Name.Value, value)
		return nil

	// ==== Expressions ====
	case *ast.SetExpression:
		switch left := node.Left.(type) {
		case *ast.Identifier:
			if !env.Has(left.Value) {
				return newError("identifier not found: %s", left.Value)
			}
			value := Eval(node.Right, env)
			if isError(value) {
				return value
			}
			env.Set(left.Value, value)
			return NULL
		}
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.CallExpression:
		fn := Eval(node.Function, env)
		if isError(fn) {
			return fn
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(fn, args)

	case *ast.FunctionLiteral:
		return &object.Function{
			Parameters: node.Parameters,
			// This, along with applyFunction(...) gives us the ability to do closures.
			// This is because the `Env' field of the function evaluated inside us would
			// have _it's_ Env field = ours + some extension.
			Env:  env,
			Body: node.Body,
		}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	case *ast.Identifier:
		return evalIdentifier(node, env)

	// Container types
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.Null:
		return NULL

	}
	return nil
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, stmt := range program.Statements {
		result = Eval(stmt, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	// We cannot reuse evalProgram for evalBlockStatement -- this is because
	// we wrap return values in a object.ReturnValue{...}, but evalProgram
	// will unwrap it. So this means that nested blocks do not work:
	//   if (true) {
	//     if (true) { return 10 }
	//     return 1
	//   }
	// produces 1 instead. When evalProgram(...) eventually receives the
	// wrapped ReturnValue, _it_ will unwrap it instead. Moreover, block
	// statements have block scoping:
	//   let u = 1;
	//   if (true) {
	//     let u = 2;
	//     u = 3;
	//   }
	//   u <-- 1
	var result object.Object = NULL
	var blockEnv *object.Environment = env
	for _, stmt := range block.Statements {
		// Optimisation: don't create a new environment until we encounter
		// a let statement. In particular, this means that most blocks will
		// not require a separate environment.
		if !blockEnv.IsBlock() {
			if _, ok := stmt.(*ast.LetStatement); ok {
				blockEnv = object.NewEnclosedEnvironment(env, true)
			}
		}
		result = Eval(stmt, blockEnv)
		switch result := result.(type) {
		case *object.ReturnValue:
			// Note that the ReturnValue is not unwrapped.
			return result
		case *object.Error:
			return result
		}
	}
	// This is another difference from evalProgram: it returns null by
	// default, if the last statement is a let or set -- contrast this
	// to evalProgram(), which can return nil.
	if result == nil {
		return NULL
	}
	return result
}

func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(node.Condition, env)
	if isError(condition) {
		return condition
	} else if isTruthy(condition) {
		return Eval(node.Consequence, env)
	} else if node.Alternative != nil {
		return Eval(node.Alternative, env)
	} else {
		return NULL
	}
}

// ==========================
// Function calling machinery
// ==========================

func evalExpressions(
	exprs []ast.Expression,
	env *object.Environment,
) []object.Object {
	result := make([]object.Object, len(exprs))
	for i, e := range exprs {
		value := Eval(e, env)
		if isError(value) {
			return []object.Object{value}
		}
		result[i] = value
	}
	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		// Create a new environment that extends _fn.Env_ (not the `canonical'
		// environment) -- this allows environment chains for closures.
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		// No ceremony needed...
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env, false)
	for paramIdx, param := range fn.Parameters {
		var value object.Object = NULL
		if paramIdx < len(args) {
			value = args[paramIdx]
		}
		env.Set(param.Value, value)
	}
	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if rv, ok := obj.(*object.ReturnValue); ok {
		return rv.Value
	}
	return obj
}

// ==================
// 'Easy' expressions
// ==================

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	// Allow chance to override builtins!
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: %s", node.Value)
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalMinusOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	v := right.(*object.Integer).Value
	return &object.Integer{Value: -v}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case operator == "is":
		// Pointer equality
		return nativeBoolToBooleanObject(left == right)
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
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		left := left.(*object.Boolean).Value
		right := right.(*object.Boolean).Value
		switch operator {
		case "||":
			return nativeBoolToBooleanObject(left || right)
		case "&&":
			return nativeBoolToBooleanObject(left && right)
		}
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	}
	return newError("unknown operator: %s %s %s",
		left.Type(), operator, right.Type())
}

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

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(
			left.(*object.Array),
			index.(*object.Integer),
		)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left.(*object.Hash), index)
	default:
		return newError("operator not supported: %s[%s]", left.Type(), index.Type())
	}
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

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	hash := &object.Hash{}
	hash.Pairs = object.NewHashTable()
	for _, entry := range node.Pairs {
		key := Eval(entry.Key, env)
		if isError(key) {
			return key
		}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}
		value := Eval(entry.Value, env)
		if isError(value) {
			return value
		}
		hash.Pairs.Set(hashKey, value)
	}
	return hash
}

// ==============
// Helper methods
// ==============

func nativeBoolToBooleanObject(b bool) *object.Boolean {
	if b {
		return TRUE
	} else {
		return FALSE
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case TRUE:
		return true
	case NULL:
		return false
	case FALSE:
		return false
	default:
		return true
	}
}

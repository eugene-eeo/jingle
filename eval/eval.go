// Package eval implemenets a simple tree-walk evaluator for
// the jingle ast.
package eval

import (
	"fmt"
	"jingle/ast"
)

type Context struct {
	scope *Scope
	g     *GlobalObjects
}

func NewContext() *Context {
	ctx := &Context{}
	ctx.g = NewGlobalObjects(ctx)
	ctx.scope = NewGlobalScope(ctx.g)
	return ctx
}

func isError(v Value) bool {
	_, ok := v.(*Error)
	return ok
}

// lookup finds a variable in the scope stack.
func (ctx *Context) lookup(name string) (Value, bool) {
	return ctx.scope.Get(name)
}

func (ctx *Context) maybeBind(obj Value, this Value) Value {
	switch obj := obj.(type) {
	case *NativeFunction:
		return obj.Bind(this)
	}
	return obj
}

// lookupAttr looks up an attribute in `obj`.
func (ctx *Context) lookupAttr(obj Value, attr string) (Value, bool) {
	// first try to find it on the object itself.
	switch x_obj := obj.(type) {
	case *Object:
		if val, ok := x_obj.attrs[attr]; ok {
			return ctx.maybeBind(val, obj), true
		}
	case *Class:
		// For classes, we first look at the class attributes.
		// Failing that, we try to find a method in the class.
		if val, ok := x_obj.attrs[attr]; ok {
			return ctx.maybeBind(val, obj), true
		}
	}
	// now try to fetch it from the class.
	// failing that, the superclass, and so on.
	for klass := obj.Klass(); klass != nil; klass = klass.super {
		// we chose arbitrarily to let attributes win over methods.
		if val, ok := klass.attrs[attr]; ok {
			return ctx.maybeBind(val, obj), true
		}
		if val, ok := klass.methods[attr]; ok {
			return ctx.maybeBind(val, obj), true
		}
	}
	return nil, false
}

// call calls the given target with the given arguments.
func (ctx *Context) call(target Value, args []Value) Value {
	switch target := target.(type) {
	case *NativeFunction:
		return target.Call(args)
	}
	// TODO: return error
	return nil
}

func (ctx *Context) Eval(node ast.Node) Value {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return ctx.evalProgram(node)
	case *ast.ExpressionStatement:
		return ctx.Eval(node.Expr)
	// Expressions
	case *ast.AttrExpression:
		target := ctx.Eval(node.Target)
		if isError(target) {
			return target
		}
		val, ok := ctx.lookupAttr(target, node.Name.Name())
		if !ok {
			return &Error{Reason: ctx.g.NewString(fmt.Sprintf(
				"object does not have attr %s",
				node.Name.Name(),
			))}
		}
		return val
	case *ast.CallExpression:
		target := ctx.Eval(node.Target)
		if isError(target) {
			return target
		}
		args := make([]Value, len(node.Args))
		for i, arg_node := range node.Args {
			args[i] = ctx.Eval(arg_node)
			if isError(args[i]) {
				return args[i]
			}
		}
		return ctx.call(target, args)
	// Literals
	case *ast.IdentifierLiteral:
		val, ok := ctx.lookup(node.Name())
		if !ok {
			return &Error{Reason: ctx.g.NewString(fmt.Sprintf(
				"name %s is undefined",
				node.Name(),
			))}
		}
		return val
	case *ast.StringLiteral:
		return ctx.g.NewString(node.Value)
	case *ast.BooleanLiteral:
		if node.Value {
			return ctx.g.TRUE
		} else {
			return ctx.g.FALSE
		}
	case *ast.NilLiteral:
		return ctx.g.NIL
	default:
		panic(fmt.Sprintf("not implemented yet: %T", node))
	}
	return nil
}

func (ctx *Context) evalProgram(prog *ast.Program) Value {
	var rv Value = ctx.g.NIL
	for _, x := range prog.Statements {
		rv = ctx.Eval(x)
	}
	return rv
}

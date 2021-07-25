package evaluator

import (
	"fmt"
	"jingle/object"
)

// === Util functions ====
func wantNArgs(args []object.Object, n int) object.Object {
	if len(args) != n {
		return newError("wrong number of arguments. got=%d, want=%d",
			len(args), n)
	}
	return nil
}

func wantType(arg string, obj object.Object, t object.ObjectType) object.Object {
	if obj.Type() != t {
		return newError("%s has wrong type. got=%s, want=%s",
			arg, obj.Type(), t)
	}
	return nil
}

// === Actual builtins ===

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := wantNArgs(args, 1); err != nil {
				return err
			}
			switch obj := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(obj.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(obj.Elements))}
			case *object.Hash:
				return &object.Integer{Value: int64(obj.Pairs.Size())}
			default:
				return newError("argument to `len` not supported, got %s", obj.Type())
			}
		},
	},
	"type": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := wantNArgs(args, 1); err != nil {
				return err
			}
			return &object.String{Value: string(args[0].Type())}
		},
	},
	"inspect": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := wantNArgs(args, 1); err != nil {
				return err
			}
			return &object.String{Value: args[0].Inspect()}
		},
	},
	"hashable": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := wantNArgs(args, 1); err != nil {
				return err
			}
			_, ok := args[0].(object.Hashable)
			return nativeBoolToBooleanObject(ok)
		},
	},
	"puts": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
}

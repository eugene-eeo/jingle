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
	"copy": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := wantNArgs(args, 1); err != nil {
				return err
			}
			switch obj := args[0].(type) {
			case *object.Array:
				elems := make([]object.Object, len(obj.Elements))
				copy(elems, obj.Elements)
				return &object.Array{Elements: elems}
			case *object.Hash:
				hash := &object.Hash{}
				hash.Pairs = object.NewHashTable()
				obj.Pairs.Iter(func(k object.Hashable, v object.Object) bool {
					hash.Pairs.Set(k, v)
					return true
				})
				return hash
			default:
				return newError("argument to `copy` not supported, got %s", obj.Type())
			}
		},
	},
	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) <= 1 {
				return newError("`push` takes >= 2 arguments. got=%d", len(args))
			}
			if err := wantType("1st argument to `push`", args[0], object.ARRAY_OBJ); err != nil {
				return err
			}
			arr := args[0].(*object.Array)
			arr.Elements = append(arr.Elements, args[1:]...)
			return arr
		},
	},
	"delete": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := wantNArgs(args, 2); err != nil {
				return err
			}
			switch obj := args[0].(type) {
			case *object.Array:
				if err := wantType("2nd argument to `delete`", args[1], object.INTEGER_OBJ); err != nil {
					return err
				}
				arr := args[0].(*object.Array)
				idx := args[1].(*object.Integer).Value
				sz := int64(len(arr.Elements))
				if idx < 0 || idx > sz-1 {
					return newError("invalid index for `delete`")
				}
				copy(arr.Elements[idx:], arr.Elements[idx+1:])
				arr.Elements[len(arr.Elements)-1] = nil
				arr.Elements = arr.Elements[:len(arr.Elements)-1]
				return arr
			default:
				return newError("argument to `delete` not supported, got %s", obj.Type())
			}
		},
	},
	"insert": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if err := wantNArgs(args, 3); err != nil {
				return err
			}
			if err := wantType("1st argument to `insert`", args[0], object.ARRAY_OBJ); err != nil {
				return err
			}
			if err := wantType("2nd argument to `insert`", args[1], object.INTEGER_OBJ); err != nil {
				return err
			}
			arr := args[0].(*object.Array)
			idx := args[1].(*object.Integer).Value
			obj := args[2]
			sz := int64(len(arr.Elements))
			if idx < 0 || idx > sz {
				return newError("invalid index for `insert`")
			}
			arr.Elements = append(arr.Elements, nil)
			copy(arr.Elements[idx+1:], arr.Elements[idx:])
			arr.Elements[idx] = obj
			return arr
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

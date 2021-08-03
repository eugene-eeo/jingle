package eval

import "fmt"

// Value represents any Jingle value.
// We take inspiration from Ruby's object implementation.
// Values should know their own class -- we will use Go's
// type assertions to convert generic Values into more
// specific ones.
type Value interface {
	Klass() *Class
}

// GlobalObjects represents the base set of globals that are needed
// to bootstrap the object system.
type GlobalObjects struct {
	ctx *Context // pointer to the current context
	// Classes
	Object         *Class // the Object class
	Class          *Class // the Class class
	NativeFunction *Class // the NativeFunction class
	Nil            *Class // the class of nil, Nil
	Boolean        *Class // class of booleans, Boolean
	String         *Class // String class
	Error          *Class // Error class
	// Literals
	TRUE  *Boolean
	FALSE *Boolean
	NIL   *Nil
}

func NewGlobalObjects(ctx *Context) *GlobalObjects {
	g := &GlobalObjects{}
	g.ctx = ctx
	g.Object = g.NewClass("Object", nil) // Object has no supertype.
	g.Class = g.NewClass("Class", g.Object)
	g.Object.klass = g.Class
	g.Class.klass = g.Class
	g.NativeFunction = g.NewClass("NativeFunction", g.Object)

	// define Object methods here (class', attrs)
	g.Object.methods["class'"] = g.NewNativeFunction(func(ref *NativeFunction, args []Value) Value {
		return ref.this.Klass()
	})
	g.Object.methods["inspect"] = g.NewNativeFunction(func(ref *NativeFunction, args []Value) Value {
		return g.NewString(fmt.Sprintf("<object %p>", ref.this))
	})
	// define Class methods here (new)
	g.Class.methods["inspect"] = g.NewNativeFunction(func(ref *NativeFunction, args []Value) Value {
		return g.NewString(fmt.Sprintf(
			"<class %s>",
			ref.this.(*Class).name,
		))
	})
	g.Class.methods["get_method"] = g.NewNativeFunction(func (ref *NativeFunction, args []Value) Value {
		name := args[0].(*String).s
		return ref.this.(*Class).methods[name]
	})
	// define NativeFunction methods here
	g.NativeFunction.methods["bind"] = g.NewNativeFunction(func (ref *NativeFunction, args []Value) Value {
		return ref.this.(*NativeFunction).Bind(args[0])
	})

	g.String = g.NewClass("String", g.Object)
	g.Nil = g.NewClass("Nil", g.Object)
	g.Nil.methods["inspect"] = g.NewNativeFunction(func(ref *NativeFunction, args []Value) Value {
		return g.NewString("nil")
	})

	g.Boolean = g.NewClass("Boolean", g.Object)
	g.Boolean.methods["inspect"] = g.NewNativeFunction(func(ref *NativeFunction, args []Value) Value {
		switch ref.this {
		case g.TRUE:
			return g.NewString("true")
		case g.FALSE:
			return g.NewString("false")
		}
		return nil
	})

	g.NIL = &Nil{Basic: Basic{klass: g.Nil}}
	g.TRUE = &Boolean{Basic: Basic{klass: g.Boolean}, b: true}
	g.FALSE = &Boolean{Basic: Basic{klass: g.Boolean}, b: false}
	return g
}

// Embedded in each concrete Value.
type Basic struct{ klass *Class }

func (b Basic) Klass() *Class {
	return b.klass
}

// Class represents a (user-defined) class.
// super is the superclass -- note that all classes except
// for the base class have a non-nil `super`.
type Class struct {
	Basic
	name    string
	attrs   map[string]Value // my attributes.
	methods map[string]Value // my methods.
	super   *Class
}

func (g *GlobalObjects) NewClass(name string, super *Class) *Class {
	return &Class{
		Basic:   Basic{klass: g.Class},
		name:    name,
		attrs:   map[string]Value{},
		methods: map[string]Value{},
		super:   super,
	}
}

// Object represents an instance of a user-defined class,
// one that is not covered by the types below.
type Object struct {
	Basic
	attrs map[string]Value
}

func (g *GlobalObjects) NewObject(klass *Class) *Object {
	return &Object{
		Basic: Basic{klass: klass},
		attrs: map[string]Value{},
	}
}

// Boolean *value* (true or false)
type Boolean struct {
	Basic
	b bool
}

// Nil *value*
type Nil struct{ Basic }

// String *value*
type String struct {
	Basic
	s string
}

func (g *GlobalObjects) NewString(str string) *String {
	return &String{Basic: Basic{klass: g.String}, s: str}
}

func (s String) String() string { return s.s }

// NativeFunction is a function written in Go.
type NativeFunction struct {
	Basic
	ctx   *Context
	this  Value
	attrs map[string]Value
	fn    func(*NativeFunction, []Value) Value
}

func (nf *NativeFunction) Bind(this Value) *NativeFunction {
	if nf.this != nil {
		return nf
	}
	return &NativeFunction{
		Basic: nf.Basic,
		this:  this,
		attrs: nf.attrs,
		fn:    nf.fn,
	}
}

func (nf *NativeFunction) Call(args []Value) Value {
	return nf.fn(nf, args)
}

func (g *GlobalObjects) NewNativeFunction(fn func(*NativeFunction, []Value) Value) *NativeFunction {
	return &NativeFunction{
		Basic: Basic{klass: g.NativeFunction},
		ctx:   g.ctx,
		this:  nil,
		attrs: map[string]Value{},
		fn:    fn,
	}
}

// Error wraps around a reason object, and is an error
// meant to be unwrapped. If there is no code to catch the error,
// the error propagates up the stack. This is NOT the Error _class_
// in the language.
type Error struct {
	Basic
	Reason Value
}

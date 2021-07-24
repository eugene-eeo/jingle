package object

func NewEnclosedEnvironment(outer *Environment, isBlock bool) *Environment {
	env := NewEnvironment()
	env.outer = outer
	env.isBlock = isBlock
	return env
}

func NewEnvironment() *Environment {
	return &Environment{
		isBlock: false,
		outer:   nil,
		store:   make(map[string]Object),
	}
}

type Environment struct {
	store   map[string]Object
	outer   *Environment
	isBlock bool
}

func (e *Environment) IsBlock() bool {
	return e.isBlock
}

func (e *Environment) Has(name string) bool {
	_, ok := e.store[name]
	if !ok && e.isBlock {
		return e.outer.Has(name)
	}
	return ok
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	// manually unwrap the call stack ourselves.
	for !ok && e.outer != nil {
		e = e.outer
		obj, ok = e.store[name]
	}
	return obj, ok
}

func (e *Environment) Let(name string, obj Object) Object {
	e.store[name] = obj
	return obj
}

func (e *Environment) Set(name string, obj Object) Object {
	if e.isBlock {
		// A bit more complicated -- we have to see if we have it,
		// otherwise find an environment that has it and let that
		// environment do the setting...
		if _, ok := e.store[name]; !ok {
			// Go to the parent. Since .Set(...) is only called
			// after a call to .Has(...), this is safe.
			return e.outer.Set(name, obj)
		}
	}
	return e.Let(name, obj)
}

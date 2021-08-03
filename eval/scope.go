package eval

type Scope struct {
	values map[string]Value
	outer  *Scope
}

func NewScope(outer *Scope) *Scope {
	return &Scope{
		values: map[string]Value{},
		outer:  outer,
	}
}

func (s *Scope) Get(name string) (Value, bool) {
	if v, ok := s.values[name]; ok {
		return v, ok
	}
	if s.outer != nil {
		return s.outer.Get(name)
	}
	return nil, false
}

func (s *Scope) Set(name string, v Value) Value {
	s.values[name] = v
	return v
}

func NewGlobalScope(g *GlobalObjects) *Scope {
	s := NewScope(nil)
	s.values["Object"] = g.Object
	s.values["Class"] = g.Class
	s.values["Boolean"] = g.Boolean
	s.values["String"] = g.String
	return s
}

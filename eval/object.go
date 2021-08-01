package eval

// Object represents any Jingle value.
type Object interface {
	SuperType() Object
	Type() Object
	Get(name string) Object
	Set(name string, obj Object) Object
}

// Constants
var (
	OBJECT = &ObjectClass{}
	CLASS  = &Class{}
	NIL    = &Nil{}
)

// Base Object. Object itself is a class, but has no supertypes.
// Any instance of a class is an instance of an object.
type ObjectClass struct{}

func (oc *ObjectClass) SuperType() Object                  { return nil }
func (oc *ObjectClass) Type() Object                       { return CLASS }
func (oc *ObjectClass) Get(name string) Object             { return nil }
func (oc *ObjectClass) Set(name string, obj Object) Object { return nil }

// A Class. Every _class_ is an instance of Class.
// A class itself is an Object, so it's supertype is Object.
type Class struct{}

func (c *Class) SuperType() Object                  { return OBJECT }
func (c *Class) Type() Object                       { return CLASS }
func (c *Class) Get(name string) Object             { return nil }
func (c *Class) Set(name string, obj Object) Object { return nil }

// The `nil' constant.
type Nil struct{}

func (n *Nil) SuperType() Object                  { return nil }
func (n *Nil) Type() Object                       { return NIL }
func (n *Nil) Get(name string) Object             { return nil }
func (n *Nil) Set(name string, obj Object) Object { return nil }

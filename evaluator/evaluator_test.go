package evaluator_test

import (
	"jingle/evaluator"
	"jingle/lexer"
	"jingle/object"
	"jingle/parser"
	"testing"
)

// ==============
// Test Functions
// ==============

type Null struct{}
type Error struct{ message interface{} }
type String struct{ value string }
type Array []interface{}
type Hash []struct {
	key   interface{}
	value interface{}
}

func TestExpressionLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"5", 5},
		{"10", 10},
		{"true", true},
		{"false", false},
		{"null", Null{}},
		{`"string"`, String{"string"}},
		{`"\0string\""`, String{"\u0000string\""}},
		{`"\"string\""`, String{`"string"`}},
		{"[1,2,3]", Array{1, 2, 3}},
		{`[1,"string",null,true]`, Array{1, String{"string"}, Null{}, true}},
	}
	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testObject(t, evaluated, tt.expected) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"!true", false},
		{"!false", true},
		{"!null", true},
		{"!!false", false},
		{"!!true", true},
		{"!5", false},
		{"!!5", true},
		{"-5", -5},
		{"--5", 5},
		{"-true", Error{"unknown operator: -BOOLEAN"}},
		{"-[1]", Error{"unknown operator: -ARRAY"}},
		{"!([])", false},
		{"!({})", false},
		{`!("")`, false},
	}
	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testObject(t, evaluated, tt.expected) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"5 + 5", 10},
		{"(5 + 5) * 10", 100},
		{"(5 + 5) * 15 == 150", true},
		{"5 / 5", 1},
		{"5 * 5", 25},
		{"5 - 5", 0},
		{"5 > 5", false},
		{"5 > 0", true},
		{"5 < 0", false},
		{"0 < 5", true},
		{"5 == 5", true},
		{"5 == 0", false},
		{"5 != 0", true},
		{"5 != 5", false},
		{"5 <= 1", false},
		{"5 <= 5", true},
		{"5 <= 10", true},
		{"10 >= 5", true},
		{"5 >= 5", true},
		{"1 >= 5", false},
		{"true != true", false},
		{"true == true", true},
		{"true != false", true},
		{"false != true", true},
		{"false == false", true},
		{"false != false", false},
		{"(1 < 2) == false", false},
		{"(1 == 2) == false", true},
		{"true || false", true},
		{"false || true", true},
		{"true || true", true},
		{"false || false", false},
		{"false && false", false},
		{"false && true", false},
		{"true && false", false},
		{"true && true", true},
		{"null == null", true},
		{"null != null", false},
		{"5 + 5 > 9", true},
		{"5 + 5 < 9 || true", true},
		{"5 + 5 < 11 && null == null", true},
		{`1 == null`, false},
		{`true is true`, true},
		{`false is false`, true},
		{`null is null`, true},
		{`[1] is [1]`, false},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testObject(t, evaluated, tt.expected) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestStringOps(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`"a" + "b"`, String{"ab"}},
		{`"a" + "b" == "ab"`, true},
		{`"a" + "b" != "ab"`, false},
		{`"a" > "b"`, false},
		{`"a" >= "b"`, false},
		{`"a" < "b"`, true},
		{`"a" <= "b"`, true},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testObject(t, evaluated, tt.expected) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestIfStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 1 }", 1},
		{"if (false) { 1 }", Null{}},
		{"if (null) { 1 }", Null{}},
		{"if (null) { 1 } else { 2 }", 2},
		{"if (1 < 2) { 1 }", 1},
		{"if (1 > 2) { 1 }", Null{}},
		{"if (1 < 2) { 1 } else { 2 }", 1},
		{"if (1 < 2) { true; 1 } else { 2 }", 1},
		{"if (1 > 2) { 1 } else { 2 }", 2},
		{"if (1 > 2) { true; 1 } else { false; 2 }", 2},
		{"if (null) { true; 1 } else { false; 2 }", 2},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testObject(t, evaluated, tt.expected) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestScopes(t *testing.T) {
	// Test block scoping. Each block introduces a new scope: new variables
	// declared with `let` within the block would not affect variables
	// outside of the block.
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"let t = 2; if (true) { t = 1; }; t;", 1},
		{"let t = 2; if (true) { let t = 1; }; t;", 2},
		{"let t = 2; if (true) { t = 1; let t = 3; }; t;", 1},
		{`
let f = fn() {
    let u = 1;
    let g = fn() { u = 2 };
    g
};
let g = f();
g();
		`, Error{"identifier not found: u"}},
		{`
let f = fn() {
    let t = 1;
    let g = fn() {
    	let u = 2;
    	let v = 3;
    	u = 1;
    	v
    };
    return g;
};
let g = f();
g();
		`, 3},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testObject(t, evaluated, tt.expected) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"2; return 10; 9;", 10},
		{"2; return 1 * 4; 9;", 4},
		{"if (10 > 1) { if (10 > 1) { return 10; }; return 1; }", 10},
		{"let f = fn() { return 5 }; let a = fn() { f(); return 10; }; a()", 10},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testObject(t, evaluated, tt.expected) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input        string
		errorMessage string
	}{
		{"5 + true;", "type mismatch: INTEGER + BOOLEAN"},
		{"5 + true; 5;", "type mismatch: INTEGER + BOOLEAN"},
		{"(5 + true) + 5;", "type mismatch: INTEGER + BOOLEAN"},
		{"-true;", "unknown operator: -BOOLEAN"},
		{"true + false;", "unknown operator: BOOLEAN + BOOLEAN"},
		{"5; true + false; 5", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) { true + false; }", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) { if (true) { return true + false; }; return 1; }", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (1 + true) { 1 }", "type mismatch: INTEGER + BOOLEAN"},
		{"(if (1) { return true + false; }) + 2", "unknown operator: BOOLEAN + BOOLEAN"},
		{"foobar", "identifier not found: foobar"},
		{"let foobar = 1; x = 1", "identifier not found: x"},
		{`"a" - "b"`, "unknown operator: STRING - STRING"},
		{"let a = 5 + true;", "type mismatch: INTEGER + BOOLEAN"},
		{"let a = 5; a = u + 1;", "identifier not found: u"},
		{"!(true + 5)", "type mismatch: BOOLEAN + INTEGER"},
		{"u(1)", "identifier not found: u"},
		{"let f = fn() {}; f(z)", "identifier not found: z"},
		{"u[1]", "identifier not found: u"},
		{"let u = [1]; u[z]", "identifier not found: z"},
		{"[1, u, z];", "identifier not found: u"},
	}
	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testErrorObject(t, evaluated, tt.errorMessage) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"let a = 5; a;", 5},
		{"let a = 5; a * a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; a = 10; a;", 10},
	}
	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testObject(t, evaluated, tt.expected) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestFunctionObject(t *testing.T) {
	input := `fn(x) { x + 2 }`
	obj := testEval(t, input)
	fn, ok := obj.(*object.Function)
	if !ok {
		t.Fatalf("wrong type: expected *object.Function, got=%T", fn)
	}
	if len(fn.Parameters) != 1 {
		t.Fatalf("wrong number of params: expected %d, got=%d",
			1, len(fn.Parameters))
	}
	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0].String())
	}
	expectedBody := "{ (x + 2); }"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionCall(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"let a = fn() { return 5; }; a();", 5},
		{"let a = fn(x) { return x; }; a();", Null{}},
		// Set and Let will return Null instead of exploding
		{"let a = fn(x) { x = 1; }; a();", Null{}},
		{"let a = fn(x) { let u = 1; }; a();", Null{}},
		{"let a = fn(x, y) { if (x == 1) { return y; }; return x }; a(1, 2);", 2},
		{"let a = fn(x, y) { if (x == 1) { return y; }; return x }; a(0, 2);", 0},
		{"let u = 1; let a = fn() { return u; }; a();", 1},
		{"let adder = fn(n) { fn(x) { n + x } }; let addOne = adder(1); addOne(2);", 3},
		// Cannot modify variables outside of our block scope using =
		{"let u = 1; let f = fn(n) { u = 3; return n; }; f(2);", Error{"identifier not found: u"}},
		// Cannot call non-functions
		{"let u = 1; u(2);", Error{"not a function: INTEGER"}},
	}
	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testObject(t, evaluated, tt.expected) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("a")`, 1},
		{`len(1)`, Error{}},
		{`len("one", "two")`, Error{}},
		{`len(["a"])`, 1},
		{`len([1,2,3])`, 3},

		// Meta stuff
		{`type(1)`, String{"INTEGER"}},
		{`type(1, 2)`, Error{}},
		{`inspect(1)`, String{`1`}},
		{`inspect(1, 2)`, Error{}},
		{`hashable({})`, false},
		{`hashable([])`, false},
		{`hashable(1)`, true},
		{`hashable(1, 2)`, Error{}},
		{`hashable(fn() {})`, true},

		// Arrays
		{`copy()`, Error{}},
		{`copy(1)`, Error{}},
		{`copy(["a"], 2)`, Error{}},
		{`copy(["a", 1, null])`, Array{String{"a"}, 1, Null{}}},

		{`push("a", 1)`, Error{}},
		{`push(["a"])`, Error{}},
		{`push(["a"], "b")`, Array{String{"a"}, String{"b"}}},
		{`push(["a"], "b", 2)`, Array{String{"a"}, String{"b"}, 2}},
		{`push([], "b", 2, null)`, Array{String{"b"}, 2, Null{}}},

		{`delete([], 1, 2)`, Error{}},
		{`delete("abc", 1)`, Error{}},
		{`delete([], "a")`, Error{}},
		{`delete(["a", "b", 2, 3, 4], 1)`, Array{String{"a"}, 2, 3, 4}},
		{`delete([0, 1, 2, 3], 0)`, Array{1, 2, 3}},
		{`delete([0, 1, 2, 3], 1)`, Array{0, 2, 3}},
		{`delete([0, 1, 2, 3], 1 + 1)`, Array{0, 1, 3}},
		{`delete([0, 1, 2, 3], 3)`, Array{0, 1, 2}},
		{`delete([0, 1, 2, 3], -1)`, Error{}},
		{`delete([0, 1, 2, 3], 100)`, Error{}},

		{`insert([], 0)`, Error{}},
		{`insert("a", "a", {})`, Error{}},
		{`insert([], "a", {})`, Error{}},
		{`insert([], 1, "a")`, Error{}},
		{`insert([], 0, "b")`, Array{String{"b"}}},
		{`insert([1], 0, "b")`, Array{String{"b"}, 1}},
		{`insert([1, 2], 1, "b")`, Array{1, String{"b"}, 2}},
		{`insert([1, 2], 2, "b")`, Array{1, 2, String{"b"}}},

		// Hash tables
		{`copy({"a": 1})`, Hash{{String{"a"}, 1}}},
	}
	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testObject(t, evaluated, tt.expected) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`1[0]`, Error{"operator not supported: INTEGER[INTEGER]"}},
		{`[1,2,3][0]`, 1},
		{`[1,2,3][1]`, 2},
		{`[1,2,3][2]`, 3},
		{`[1,2,3][3]`, Null{}},
		{`[1,2,3][-1]`, Null{}},
		{`{"a": 1}["a"]`, 1},
		{`{"a": 1}["b"]`, Null{}},
		{`{"a": 1}[[1,2,3]]`, Error{"unusable as hash key: ARRAY"}},
		{`{"a": 1}[{"a": 2}]`, Error{"unusable as hash key: HASH"}},
		{`{puts: 1}[puts]`, 1},
		{`
let f = fn() {};
let h = {f: 1};
h[f]`, 1},
	}
	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testObject(t, evaluated, tt.expected) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestHashLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`{}`, Hash{}},
		{`{"a": true}`, Hash{{String{"a"}, true}}},
		{`{null: 1, true: false,}`, Hash{{Null{}, 1}, {true, false}}},
		{`{1: "a"}`, Hash{{1, String{"a"}}}},
		{`{[1,2,3]: "a"}`, Error{"unusable as hash key: ARRAY"}},
		{`{{"a":1}: "a"}`, Error{"unusable as hash key: HASH"}},
		{`{u: "a"}`, Error{"identifier not found: u"}},
		{`{"u": a}`, Error{"identifier not found: a"}},
	}
	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		if !testObject(t, evaluated, tt.expected) {
			t.Errorf("tests[%d]: failed evaluating %q", i, tt.input)
		}
	}
}

func TestShortCircuit(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true || explode()) { 1 }", 1},
		{"if (false && explode()) { explode() } else { 2 }", 2},
		{"if (1) { 2 } else { explode() }", 2},
		{"true || a", true},
		{"1 || a", 1},
		{"2 || b", 2},
		{"false && b", false},
		{"{} && b", Error{nil}},
	}
	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		env := object.NewEnvironment()
		env.Set("explode", &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				return &object.Error{Message: "HEY!"}
			},
		})
		program := p.ParseProgram()
		checkParserErrors(t, p)
		if !testObject(t, evaluator.Eval(program, env), tt.expected) {
			t.Errorf("tests[%d]: failed", i)
		}
	}
}

// =================
// Testing Utilities
// =================

func testObject(t *testing.T, obj object.Object, expected interface{}) bool {
	switch expected := expected.(type) {
	case int64:
		return testIntegerObject(t, obj, expected)
	case int:
		return testIntegerObject(t, obj, int64(expected))
	case bool:
		return testBooleanObject(t, obj, expected)
	case Error:
		return testErrorObject(t, obj, expected.message)
	case Null:
		return testNullObject(t, obj)
	case String:
		return testStringObject(t, obj, expected.value)
	case Array:
		return testArrayObject(t, obj, expected)
	case Hash:
		return testHashObject(t, obj, expected)
	}
	t.Fatalf("unhandled type for expected. got=%T", expected)
	return false
}

func interfaceToObject(it interface{}) object.Object {
	switch v := it.(type) {
	case int64:
		return &object.Integer{Value: v}
	case int:
		return &object.Integer{Value: int64(v)}
	case bool:
		if v {
			return evaluator.TRUE
		} else {
			return evaluator.FALSE
		}
	// case Error:
	// 	return &object.Error{}
	case Null:
		return evaluator.NULL
	case String:
		return &object.String{Value: v.value}
	case Array:
		elems := []object.Object{}
		for _, thing := range v {
			elems = append(elems, interfaceToObject(thing))
		}
		return &object.Array{Elements: elems}
	case Hash:
		pairs := object.NewHashTable()
		for _, thing := range v {
			pairs.Set(
				interfaceToObject(thing.key).(object.Hashable),
				interfaceToObject(thing.value),
			)
		}
		return &object.Hash{Pairs: pairs}
	}
	panic("unhandled type")
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not *object.Integer. got=%T(%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("expected result.Value=%d, got=%d", expected, result.Value)
		return false
	}
	return true
}

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object is not *object.String. got=%T(%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("expected result.Value=%q, got=%q", expected, result.Value)
		return false
	}
	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not *object.Boolean. got=%T(%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("expected result.Value=%t, got=%t", expected, result.Value)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	_, ok := obj.(*object.Null)
	if !ok {
		t.Errorf("object is not *object.Null. got=%T(%+v)", obj, obj)
		return false
	}
	return true
}

func testArrayObject(t *testing.T, obj object.Object, array Array) bool {
	arr, ok := obj.(*object.Array)
	if !ok {
		t.Errorf("object is not *object.Array. got=%T(%+v)", obj, obj)
		return false
	}
	if len(arr.Elements) != len(array) {
		t.Errorf("wrong number of elements: expected=%d, got=%d",
			len(array), len(arr.Elements))
		return false
	}
	passed := true
	for i, obj := range arr.Elements {
		if !testObject(t, obj, array[i]) {
			t.Errorf("arr[%d]: invalid item", i)
			passed = false
		}
	}
	return passed
}

func testHashObject(t *testing.T, obj object.Object, expected Hash) bool {
	hash, ok := obj.(*object.Hash)
	if !ok {
		t.Errorf("object is not *object.Hash. got=%T(%+v)", obj, obj)
		return false
	}
	if hash.Pairs.Size() != uint64(len(expected)) {
		t.Errorf("wrong number of elements: expected=%d, got=%d",
			len(expected), hash.Pairs.Size())
		return false
	}
	passed := true
	for _, test := range expected {
		value, ok := hash.Pairs.Get(interfaceToObject(test.key).(object.Hashable))
		if !ok {
			t.Errorf("hash[%T(%+v)]: not in map", test.key, test.key)
			passed = false
		}
		if !testObject(t, value, test.value) {
			t.Errorf("hash[%T(%+v)]: invalid value", test.value, test.value)
			passed = false
		}
	}
	return passed
}

func testErrorObject(t *testing.T, obj object.Object, msg interface{}) bool {
	err, ok := obj.(*object.Error)
	if !ok {
		t.Errorf("object is not *object.Error. got=%T(%+v)", obj, obj)
		return false
	}
	if msgString, ok := msg.(string); ok && err.Message != msgString {
		t.Errorf("wrong error message: expected=%q, got=%q", msg, err.Message)
		return false
	}
	return true
}

func testEval(t *testing.T, input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	env := object.NewEnvironment()
	program := p.ParseProgram()
	checkParserErrors(t, p)
	return evaluator.Eval(program, env)
}

func checkParserErrors(t *testing.T, p *parser.Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

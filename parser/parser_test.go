package parser_test

import (
	"fmt"
	"jingle/ast"
	"jingle/lexer"
	"jingle/parser"
	"testing"
)

type Null struct{}
type String struct{ value string }
type Array []interface{}
type Hash []struct {
	Key   interface{}
	Value interface{}
}
type ExprTestFunc func(ast.Expression) bool

// ==============
// Test Utilities
// ==============

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

func testParserErrors(t *testing.T, p *parser.Parser, hasErrors bool) bool {
	// Very different test -- we don't fail when the parser
	// errors out, we want to inspect the error values.
	errors := p.Errors()
	if len(errors) == 0 && hasErrors {
		t.Errorf("expected to have parser errors, got none")
		return false
	}
	if len(errors) != 0 && !hasErrors {
		t.Errorf("parser has %d errors", len(errors))
		for _, msg := range errors {
			t.Errorf("parser error: %q", msg)
		}
		return false
	}
	return true
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s",
			value, integ.TokenLiteral())
		return false
	}
	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %s. got %s", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got %s", value, ident.Value)
		return false
	}
	return true
}

func testStringLiteral(t *testing.T, exp ast.Expression, value string) bool {
	str, ok := exp.(*ast.StringLiteral)
	if !ok {
		t.Errorf("exp not *ast.String. got=%T(%+v)", exp, exp)
		return false
	}
	if str.Value != value {
		t.Errorf("str.Value not %q. got %q", value, str.Value)
		return false
	}
	if str.TokenLiteral() != value {
		t.Errorf("str.TokenLiteral() not %q. got %q", value, str.TokenLiteral())
		return false
	}
	return true
}

func testNull(t *testing.T, exp ast.Expression) bool {
	nullexpr := exp.(*ast.Null)
	if nullexpr.TokenLiteral() != "null" {
		t.Errorf("nullexpr.TokenLiteral not 'null'. got %s", nullexpr.TokenLiteral())
		return false
	}
	return true
}

func testBoolean(t *testing.T, exp ast.Expression, value bool) bool {
	bexpr, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}

	if bexpr.Value != value {
		t.Errorf("bexpr.Value not %t. got %t", value, bexpr.Value)
		return false
	}

	var vstring string
	if value {
		vstring = "true"
	} else {
		vstring = "false"
	}
	if bexpr.TokenLiteral() != vstring {
		t.Errorf("bexpr.TokenLiteral not %s. got %s",
			vstring, bexpr.TokenLiteral())
		return false
	}
	return true
}

func testArrayLiteral(t *testing.T, exp ast.Expression, value Array) bool {
	array, ok := exp.(*ast.ArrayLiteral)
	if !ok {
		t.Errorf("exp not *ast.ArrayLiteral. got=%T(%+v)", exp, exp)
		return false
	}
	if array.TokenLiteral() != "[" {
		t.Errorf("array.TokenLiteral() not %q. got=%q", "[", array.TokenLiteral())
		return false
	}
	if len(array.Elements) != len(value) {
		t.Errorf("len(array.Elements) not %d. got=%d",
			len(value), len(array.Elements))
		return false
	}
	for i, test := range value {
		if !testLiteralExpression(t, array.Elements[i], test) {
			return false
		}
	}
	return true
}

func testHashLiteral(t *testing.T, exp ast.Expression, value Hash) bool {
	hash, ok := exp.(*ast.HashLiteral)
	if !ok {
		t.Errorf("exp not *ast.ArrayLiteral. got=%T(%+v)", exp, exp)
		return false
	}
	if hash.TokenLiteral() != "{" {
		t.Errorf("hash.TokenLiteral() not %q. got=%q", "{", hash.TokenLiteral())
		return false
	}
	if len(hash.Pairs) != len(value) {
		t.Errorf("len(hash.Pairs) not %d. got=%d",
			len(value), len(hash.Pairs))
		return false
	}
	for i, test := range value {
		if !testLiteralExpression(t, hash.Pairs[i].Key, test.Key) || !testLiteralExpression(t, hash.Pairs[i].Value, test.Value) {
			return false
		}
	}
	return true
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBoolean(t, exp, v)
	case Null:
		return testNull(t, exp)
	case String:
		return testStringLiteral(t, exp, v.value)
	case Array:
		return testArrayLiteral(t, exp, v)
	case ExprTestFunc:
		return v(exp)
	case Hash:
		return testHashLiteral(t, exp, v)
	}
	t.Errorf("type of expected not handled. got=%T", expected)
	return false
}

// Test the parsing of a single expression statement.
// That is, programs that look like this:
//   1;
//   foobar;
//   true;
func testParseExpressionStatement(
	t *testing.T,
	input string,
	value interface{},
) bool {
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Errorf("expected 1 statement, got=%d", len(program.Statements))
		return false
	}

	exprStmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("expected 1 *ast.ExpressionStatement, got=%T", exprStmt)
		return false
	}

	return testLiteralExpression(t, exprStmt.Expression, value)
}

func testLetStatement(t *testing.T, s ast.Statement, name string, value interface{}) bool {
	if s.TokenLiteral() != "let" {
		t.Fatalf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}
	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}
	if !testIdentifier(t, letStmt.Name, name) {
		return false
	}
	val := letStmt.Value
	if !testLiteralExpression(t, val, value) {
		return false
	}
	return true
}

func testReturnStatement(t *testing.T, s ast.Statement, value interface{}) bool {
	if s.TokenLiteral() != "return" {
		t.Fatalf("s.TokenLiteral not 'return'. got=%q", s.TokenLiteral())
		return false
	}
	retStmt, ok := s.(*ast.ReturnStatement)
	if !ok {
		t.Errorf("s not *ast.ReturnStatement. got=%T", s)
		return false
	}
	val := retStmt.ReturnValue
	if !testLiteralExpression(t, val, value) {
		return false
	}
	return true
}

func testSetStatement(t *testing.T, s ast.Statement, name string, value interface{}) bool {
	if s.TokenLiteral() != "=" {
		t.Fatalf("s.TokenLiteral not '='. got=%q", s.TokenLiteral())
		return false
	}
	setStmt, ok := s.(*ast.SetStatement)
	if !ok {
		t.Errorf("s not *ast.SetStatement. got=%T", s)
		return false
	}
	if !testIdentifier(t, setStmt.Name, name) {
		return false
	}
	val := setStmt.Value
	if !testLiteralExpression(t, val, value) {
		return false
	}
	return true
}

func testPrefixExpression(
	t *testing.T,
	exp ast.Expression,
	operator string,
	right interface{},
) bool {
	pexp, ok := exp.(*ast.PrefixExpression)
	if !ok {
		t.Fatalf("stmt is not ast.PrefixExpression. got=%T", exp)
	}

	if pexp.Operator != operator {
		t.Fatalf("pexp.Operator is not '%s'. got=%s",
			operator, pexp.Operator)
	}

	if !testLiteralExpression(t, pexp.Right, right) {
		return false
	}
	return true
}

func testIndexExpression(
	t *testing.T,
	exp ast.Expression,
	left interface{},
	index interface{},
) bool {
	indexExpr, ok := exp.(*ast.IndexExpression)
	if !ok {
		t.Errorf("exp is not an *ast.IndexExpression. got=%T(%s)", exp, exp)
		return false
	}
	if !testLiteralExpression(t, indexExpr.Left, left) {
		return false
	}
	if indexExpr.TokenLiteral() != "[" {
		t.Errorf("indexExpr.TokenLiteral() is not '['. got=%q",
			indexExpr.TokenLiteral())
		return false
	}
	if !testLiteralExpression(t, indexExpr.Index, index) {
		return false
	}
	return true
}

func testInfixExpression(
	t *testing.T,
	exp ast.Expression,
	left interface{},
	operator string,
	right interface{},
) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not an ast.InfixExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("opExp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if opExp.TokenLiteral() != operator {
		t.Errorf("opExp.TokenLiteral() is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

// ============
// Actual Tests
// ============

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
		{"let u = null;", "u", Null{}},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}
		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier, tt.expectedValue) {
			return
		}
	}
}

func TestSetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"x = 5;", "x", 5},
		{"y = true;", "y", true},
		{"foobar = y;", "foobar", "y"},
		{"u = null;", "u", Null{}},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}
		stmt := program.Statements[0]
		if !testSetStatement(t, stmt, tt.expectedIdentifier, tt.expectedValue) {
			return
		}
	}
}

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		input string
		value interface{}
	}{
		{"return 500;", 500},
		{"return true;", true},
		{"return null;", Null{}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 3 statements. got=%d",
				len(program.Statements))
		}

		testReturnStatement(t, program.Statements[0], tt.value)
	}
}

func TestExpressionStatements(t *testing.T) {
	tests := []struct {
		input string
		value interface{}
	}{
		{"foobar;", "foobar"},
		{"null;", Null{}},
		{"10;", 10},
		{"123456789;", 123456789},
		{"true;", true},
		{"false;", false},
		{"\"abc\";", String{"abc"}},
		{`"abc \"quotes\"\r\n\t\0"`, String{"abc \"quotes\"\r\n\t\u0000"}},
	}

	for d, tt := range tests {
		if !testParseExpressionStatement(t, tt.input, tt.value) {
			t.Errorf("tests[%d]: failed to parse %q correctly", d, tt.input)
		}
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		right    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!a;", "!", "a"},
		{"-b;", "-", "b"},
		{"!true;", "!", true},
		{"!false;", "!", false},
		{"!null;", "!", Null{}},
		{"!\"abc\";", "!", String{"abc"}},
		{"-\"abc\";", "-", String{"abc"}},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			// t.Errorf("program.Statements %+v\n", program.Statements)
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not a ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}
		testPrefixExpression(t, stmt.Expression, tt.operator, tt.right)
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input    string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 >= 5;", 5, ">=", 5},
		{"5 <= 5;", 5, "<=", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"a == 5;", "a", "==", 5},
		{"alice * bob;", "alice", "*", "bob"},
		{"true == true;", true, "==", true},
		{"false == false;", false, "==", false},
		{"true != false;", true, "!=", false},
		{"a != true;", "a", "!=", true},
		{"a != null;", "a", "!=", Null{}},
		{"true && true;", true, "&&", true},
		{"true || true;", true, "||", true},
		{`"abc" || "def";`, String{"abc"}, "||", String{"def"}},
		{`"abc" + "def";`, String{"abc"}, "+", String{"def"}},
		{`"abc" <= "def";`, String{"abc"}, "<=", String{"def"}},
		{`a is a;`, "a", "is", "a"},
		{`a is b;`, "a", "is", "b"},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}
		testInfixExpression(t, stmt.Expression, tt.left, tt.operator, tt.right)
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"true", "true;"},
		{"false", "false;"},
		{"hello", "hello;"},
		{"3 > 5 == false", "((3 > 5) == false);"},
		{"a < 5 == true", "((a < 5) == true);"},
		{"-a * b", "((-a) * b);"},
		{"!-a", "(!(-a));"},
		{"a + b + c", "((a + b) + c);"},
		{"a + b - c", "((a + b) - c);"},
		{"a * b * c", "((a * b) * c);"},
		{"a / b / c", "((a / b) / c);"},
		{"a * b + c", "((a * b) + c);"},
		{"a + b * c", "(a + (b * c));"},
		{"a / b + c", "((a / b) + c);"},
		{"a + b / c", "(a + (b / c));"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f);"},
		{"-a + b", "((-a) + b);"},
		{"!a + b", "((!a) + b);"},
		{"a + !b", "(a + (!b));"},
		{"a + b != c", "((a + b) != c);"},
		{"a == !b + c", "(a == ((!b) + c));"},
		{"1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4);"},
		{"(5 + 5) * 2", "((5 + 5) * 2);"},
		{"2 / (5 + 5)", "(2 / (5 + 5));"},
		{"-(5 + 5)", "(-(5 + 5));"},
		{"!(true == true)", "(!(true == true));"},
		{"if (x < y) { x } else { y } == 5", "(if (x < y) { x; } else { y; } == 5);"},
		{"if (x < y) { x } == 5", "(if (x < y) { x; } == 5);"},
		{"fn(x, y) { x + y } == true", "(fn(x, y) { (x + y); } == true);"},
		{"a + add(b * c) + d", "((a + add((b * c))) + d);"},
		{"add(a, b, 1, 2 * 3, add(4, 5))", "add(a, b, 1, (2 * 3), add(4, 5));"},
		{"map(lst, fn(x) { x + 1 })", "map(lst, fn(x) { (x + 1); });"},
		{"a + b == null", "((a + b) == null);"},
		{"a == b || b == c", "((a == b) || (b == c));"},
		{"a == b && b == c", "((a == b) && (b == c));"},
		{`"a" + "b" == "ab"`, `(("a" + "b") == "ab");`},
		{`"a" + "b" != "bc"`, `(("a" + "b") != "bc");`},
		{`a * [1, 2, 3, 4][b * c] * d`, `((a * ([1, 2, 3, 4][(b * c)])) * d);`},
		{`add(vec[1], u * vec[2])`, `add((vec[1]), (u * (vec[2])));`},
		{`1 is 2`, `(1 is 2);`},
		{`1 is 2 == true`, `(1 is (2 == true));`},
		{`1 is 2 * 2`, `(1 is (2 * 2));`},
		{`1 is 2 || 2 is 3`, `((1 is 2) || (2 is 3));`},
		{`1 is !2 && 2 is 3`, `((1 is (!2)) && (2 is 3));`},
		{`[1] is [1]`, `([1] is [1]);`},
		{`[2] is [1]`, `([2] is [1]);`},
		{`{} is {}`, `({} is {});`},
		{`[1,2,3][1][2][3]`, `((([1, 2, 3][1])[2])[3]);`},
		{`[1,2,3][1][2]`, `(([1, 2, 3][1])[2]);`},
		{`[1,2,3][1]`, `([1, 2, 3][1]);`},
	}

	for d, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("tests[%d]: expected=%q, got=%q", d,
				tt.expected, actual)
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not an ast.IfExpression. got=%T",
			stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Fatalf("exp.Consequence.Statements does not contain %d statements. got=%d",
			1, len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if exp.Alternative != nil {
		t.Errorf("exp.Alternative was not nil. got=%+v", exp.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not an ast.IfExpression. got=%T",
			stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Fatalf("exp.Consequence.Statements does not contain %d statements. got=%d",
			1, len(exp.Consequence.Statements))
	}
	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Consequence.Statements[0])
	}
	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if len(exp.Alternative.Statements) != 1 {
		t.Fatalf("exp.Alternative.Statements does not contain %d statements. got=%d",
			1, len(exp.Alternative.Statements))
	}
	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Alternative.Statements[0])
	}
	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestFunctionLiteral(t *testing.T) {
	input := `fn(x, y) { x + y; }`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T",
			stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function.Parameters does not contain %d params. got=%d",
			2, len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements does not contain %d statements. got=%d",
			1, len(function.Body.Statements))
	}

	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function.Body.Statements[0] is not ast.ExpressionStatement. got=%T",
			function.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{"fn() {}", []string{}},
		{"fn(x) {}", []string{"x"}},
		{"fn(x,) {}", []string{"x"}},
		{"fn(x, y) {}", []string{"x", "y"}},
		{"fn(x, y,) {}", []string{"x", "y"}},
		{"fn(x, y, z) {}", []string{"x", "y", "z"}},
		{"fn(x, y, z,) {}", []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		fn := stmt.Expression.(*ast.FunctionLiteral)

		if len(fn.Parameters) != len(tt.expectedParams) {
			t.Errorf("expected fn.Parameters to be %d. got=%d",
				len(tt.expectedParams), len(fn.Parameters))
		}

		for i, ident := range tt.expectedParams {
			testLiteralExpression(t, fn.Parameters[i], ident)
		}
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d",
			1, len(program.Statements))
	}

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	call := stmt.Expression.(*ast.CallExpression)
	if !testIdentifier(t, call.Function, "add") {
		return
	}

	if len(call.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d", len(call.Arguments))
	}

	testLiteralExpression(t, call.Arguments[0], 1)
	testInfixExpression(t, call.Arguments[1], 2, "*", 3)
	testInfixExpression(t, call.Arguments[2], 4, "+", 5)
}

func TestMultiStatements(t *testing.T) {
	input := `
let a = 1;
a = 2;
return a;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Errorf("incorrect number of statements. expected=%d, got=%d",
			3, len(program.Statements))
		return
	}
	let := program.Statements[0].(*ast.LetStatement)
	set := program.Statements[1].(*ast.SetStatement)
	ret := program.Statements[2].(*ast.ReturnStatement)

	testLetStatement(t, let, "a", 1)
	testSetStatement(t, set, "a", 2)
	testReturnStatement(t, ret, "a")
}

func TestParsingArrayLiterals(t *testing.T) {
	tests := []struct {
		input string
		value Array
	}{
		{"[]", Array{}},
		{"[1,2]", Array{1, 2}},
		{"[1,2,]", Array{1, 2}},
		{"[a,1,true,null,\"abc\"]", Array{"a", 1, true, Null{}, String{"abc"}}},
		{"[1 + 1, 1 * 2]", Array{
			ExprTestFunc(func(node ast.Expression) bool {
				return testInfixExpression(t, node, 1, "+", 1)
			}),
			ExprTestFunc(func(node ast.Expression) bool {
				return testInfixExpression(t, node, 1, "*", 2)
			}),
		}},
	}
	for i, tt := range tests {
		if !testParseExpressionStatement(t, tt.input, tt.value) {
			t.Errorf("tests[%d]: cannot parse %q correctly",
				i, tt.input)
			continue
		}
	}
}

func TestParsingIndexExpressions(t *testing.T) {
	tests := []struct {
		input string
		left  interface{}
		index interface{}
	}{
		{"[][1]", Array{}, 1},
		{"1[2 + 2]", 1, ExprTestFunc(func(exp ast.Expression) bool {
			return testInfixExpression(t, exp, 2, "+", 2)
		})},
	}
	for i, tt := range tests {
		if !testParseExpressionStatement(
			t, tt.input,
			ExprTestFunc(func(exp ast.Expression) bool {
				return testIndexExpression(t, exp, tt.left, tt.index)
			}),
		) {
			t.Errorf("tests[%d]: cannot parse %q correctly",
				i, tt.input)
			continue
		}
	}
}

func TestParsingHashLiterals(t *testing.T) {
	tests := []struct {
		input string
		hash  interface{}
	}{
		{"{}", Hash{}},
		{"{1: 2,}", Hash{{1, 2}}},
		{"{1: 2, 2: 3}", Hash{{1, 2}, {2, 3}}},
		{"{1: 2, 2: 3,}", Hash{{1, 2}, {2, 3}}},
		{"{[1]: 2, 2: 3,}", Hash{{Array{1}, 2}, {2, 3}}},
		{"{[1]: 2, [2]: 3,}", Hash{{Array{1}, 2}, {Array{2}, 3}}},
		{"{[1]: {1: 2}, [2]: 3,}", Hash{{Array{1}, Hash{{1, 2}}}, {Array{2}, 3}}},
		{`{"a": 1, "b": null, true: 1, b: false}`, Hash{
			{String{"a"}, 1},
			{String{"b"}, Null{}},
			{true, 1},
			{"b", false},
		}},
		{`{1 + 2: true || false,}`, Hash{{
			ExprTestFunc(func(e ast.Expression) bool { return testInfixExpression(t, e, 1, "+", 2) }),
			ExprTestFunc(func(e ast.Expression) bool { return testInfixExpression(t, e, true, "||", false) }),
		}}},
	}
	for i, tt := range tests {
		if !testParseExpressionStatement(t, tt.input, tt.hash) {
			t.Errorf("tests[%d]: cannot parse %q correctly",
				i, tt.input)
			continue
		}
	}
}

func TestSemicolonRules(t *testing.T) {
	tests := []struct {
		input     string
		hasErrors bool
	}{
		// === Let and set statements ===
		{"let u = 1 false", true},
		{"u = 2 false", true},
		// Within a block...
		{"if (false) { let u = 1 [1,2,3] }", true},
		{"if (false) { u = 1 [1,2,3] }", true},
		// === Return statements ===
		{"return 1 false", true},
		{"if (false) { return [1,2] 1 }", true},
		// === ExpressionStatement (If) ===
		{"if (false) { 2 } 1", false},
		// === ExpressionStatement (Function) ===
		{"fn (x') { } 1", false},
		// === ExpressionStatement (Others) ===
		{"1 1", true},
		{"{} 1", true},
		{"[] 1", true},
		{"true 1", true},
	}

	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		p.ParseProgram()
		if !testParserErrors(t, p, tt.hasErrors) {
			t.Errorf("tests[%d]: expected parsing hasErrors=%t",
				i, tt.hasErrors)
		}
	}
}

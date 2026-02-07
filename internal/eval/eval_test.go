package eval

import (
	"os"
	"testing"

	"github.com/boattime/awsl/internal/lexer"
	"github.com/boattime/awsl/internal/parser"
)

// testEval parses and evaluates the input, returning the result.
func testEval(input string) Object {
	l := lexer.New(input)
	p := parser.New(l)
	env := NewEnvironment(os.Stdout)
	program := p.ParseProgram()
	return Eval(program, env)
}

// testIntegerObject checks that obj is an Integer with the expected value.
func testIntegerObject(t *testing.T, obj Object, expected int64) bool {
	t.Helper()

	result, ok := obj.(*Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}

	return true
}

// testFloatObject checks that obj is a Float with the expected value.
func testFloatObject(t *testing.T, obj Object, expected float64) bool {
	t.Helper()

	result, ok := obj.(*Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%f, want=%f", result.Value, expected)
		return false
	}

	return true
}

// testStringObject checks that obj is a String with the expected value.
func testStringObject(t *testing.T, obj Object, expected string) bool {
	t.Helper()

	result, ok := obj.(*String)
	if !ok {
		t.Errorf("object is not String. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%q, want=%q", result.Value, expected)
		return false
	}

	return true
}

// testBooleanObject checks that obj is a Boolean with the expected value.
func testBooleanObject(t *testing.T, obj Object, expected bool) bool {
	t.Helper()

	result, ok := obj.(*Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
		return false
	}

	return true
}

// testNullObject checks that obj is NULL.
func testNullObject(t *testing.T, obj Object) bool {
	t.Helper()

	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}

	return true
}

// testErrorObject checks that obj is an Error containing the expected message.
func testErrorObject(t *testing.T, obj Object, expectedMessage string) bool {
	t.Helper()

	errObj, ok := obj.(*Error)
	if !ok {
		t.Errorf("object is not Error. got=%T (%+v)", obj, obj)
		return false
	}

	if errObj.Message != expectedMessage {
		t.Errorf("wrong error message. got=%q, want=%q", errObj.Message, expectedMessage)
		return false
	}

	return true
}

func TestEvalIntegerLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5;", 5},
		{"10;", 10},
		{"0;", 0},
		{"9223372036854775807;", 9223372036854775807},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestEvalFloatLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14;", 3.14},
		{"0.5;", 0.5},
		{"100.001;", 100.001},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testFloatObject(t, evaluated, tt.expected)
		})
	}
}

func TestEvalStringLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello";`, "hello"},
		{`"";`, ""},
		{`"hello world";`, "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestEvalBooleanLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestEvalNullLiteral(t *testing.T) {
	evaluated := testEval("null;")
	testNullObject(t, evaluated)
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true;", false},
		{"!false;", true},
		{"!!true;", true},
		{"!!false;", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestBangOperatorWithNull(t *testing.T) {
	evaluated := testEval("!null;")
	testBooleanObject(t, evaluated, true)
}

func TestBangOperatorWithInteger(t *testing.T) {
	evaluated := testEval("!5;")
	testBooleanObject(t, evaluated, false)
}

func TestCallExpressionBuiltin(t *testing.T) {
	env := NewEnvironment(os.Stdout)
	env.Set("add", &Builtin{
		Name: "add",
		Fn: func(env *Environment, args ...Object) Object {
			if len(args) != 2 {
				return &Error{Message: "add requires 2 arguments"}
			}
			a, ok1 := args[0].(*Integer)
			b, ok2 := args[1].(*Integer)
			if !ok1 || !ok2 {
				return &Error{Message: "add requires integers"}
			}
			return &Integer{Value: a.Value + b.Value}
		},
	})

	l := lexer.New("add(2, 3);")
	p := parser.New(l)
	program := p.ParseProgram()
	result := Eval(program, env)

	testIntegerObject(t, result, 5)
}

func TestCallExpressionNotAFunction(t *testing.T) {
	evaluated := testEval("x = 5; x();")
	testErrorObject(t, evaluated, "not a function: INTEGER")
}

func TestCallExpressionUndefinedFunction(t *testing.T) {
	evaluated := testEval("foo();")
	testErrorObject(t, evaluated, "undefined variable: foo")
}

func TestCallExpressionArgumentError(t *testing.T) {
	env := NewEnvironment(os.Stdout)
	env.Set("identity", &Builtin{
		Name: "identity",
		Fn: func(env *Environment, args ...Object) Object {
			return args[0]
		},
	})

	l := lexer.New("identity(x);")
	p := parser.New(l)
	program := p.ParseProgram()
	result := Eval(program, env)

	testErrorObject(t, result, "undefined variable: x")
}

func TestMinusPrefixOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"-5;", -5},
		{"-10;", -10},
		{"--5;", 5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestMinusPrefixOperatorFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"-3.14;", -3.14},
		{"--2.5;", 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testFloatObject(t, evaluated, tt.expected)
		})
	}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5 + 5;", 10},
		{"5 - 5;", 0},
		{"5 * 5;", 25},
		{"10 / 2;", 5},
		{"5 + 5 + 5 + 5 - 10;", 10},
		{"2 * 2 * 2 * 2 * 2;", 32},
		{"5 * 2 + 10;", 20},
		{"5 + 2 * 10;", 25},
		{"20 + 2 * -10;", 0},
		{"50 / 2 * 2 + 10;", 60},
		{"2 * (5 + 10);", 30},
		{"3 * 3 * 3 + 10;", 37},
		{"3 * (3 * 3) + 10;", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10;", 50},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestFloatArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1.5 + 2.5;", 4.0},
		{"5.0 - 2.0;", 3.0},
		{"2.0 * 3.0;", 6.0},
		{"10.0 / 4.0;", 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testFloatObject(t, evaluated, tt.expected)
		})
	}
}

func TestStringConcatenation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello" + " " + "world";`, "hello world"},
		{`"foo" + "bar";`, "foobar"},
		{`"" + "test";`, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestIntegerComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1 < 2;", true},
		{"1 > 2;", false},
		{"1 < 1;", false},
		{"1 > 1;", false},
		{"1 <= 2;", true},
		{"1 >= 2;", false},
		{"1 <= 1;", true},
		{"1 >= 1;", true},
		{"1 == 1;", true},
		{"1 != 1;", false},
		{"1 == 2;", false},
		{"1 != 2;", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestFloatComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1.0 < 2.0;", true},
		{"1.0 > 2.0;", false},
		{"1.5 == 1.5;", true},
		{"1.5 != 2.5;", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestBooleanComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true == true;", true},
		{"false == false;", true},
		{"true == false;", false},
		{"true != false;", true},
		{"false != true;", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestStringComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`"hello" == "hello";`, true},
		{`"hello" == "world";`, false},
		{`"hello" != "world";`, true},
		{`"" == "";`, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestNullComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"null == null;", true},
		{"null != null;", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestGroupedExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"(5);", 5},
		{"(5 + 5);", 10},
		{"(5 + 5) * 2;", 20},
		{"((5 + 5) * 2);", 20},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestMultipleStatements(t *testing.T) {
	input := `
		5;
		10;
		15;
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 15)
}

func TestDivisionByZeroError(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{"10 / 0;", "division by zero"},
		{"5 + 5; 10 / 0;", "division by zero"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testErrorObject(t, evaluated, tt.expectedMessage)
		})
	}
}

func TestDivisionByZeroErrorFloat(t *testing.T) {
	evaluated := testEval("10.0 / 0.0;")
	testErrorObject(t, evaluated, "division by zero")
}

func TestTypeMismatchError(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{`5 + "hello";`, "type mismatch: INTEGER + STRING"},
		{`"hello" - "world";`, "unknown operator: STRING - STRING"},
		{`5 - true;`, "type mismatch: INTEGER - BOOLEAN"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testErrorObject(t, evaluated, tt.expectedMessage)
		})
	}
}

func TestUnknownPrefixOperatorError(t *testing.T) {
	evaluated := testEval(`-"hello";`)
	testErrorObject(t, evaluated, "unknown operator: -STRING")
}

func TestErrorStopsEvaluation(t *testing.T) {
	input := `
		10 / 0;
		5 + 5;
	`
	evaluated := testEval(input)

	errObj, ok := evaluated.(*Error)
	if !ok {
		t.Fatalf("expected Error, got %T", evaluated)
	}

	if errObj.Message != "division by zero" {
		t.Errorf("wrong error message. got=%q", errObj.Message)
	}
}

func TestBooleanSingletonsUsed(t *testing.T) {
	trueResult := testEval("true;")
	if trueResult != TRUE {
		t.Error("expected TRUE singleton")
	}

	falseResult := testEval("false;")
	if falseResult != FALSE {
		t.Error("expected FALSE singleton")
	}

	compResult := testEval("5 > 3;")
	if compResult != TRUE {
		t.Error("expected TRUE singleton from comparison")
	}
}

func TestNullSingletonUsed(t *testing.T) {
	result := testEval("null;")
	if result != NULL {
		t.Error("expected NULL singleton")
	}
}

func TestEmptyProgram(t *testing.T) {
	evaluated := testEval("")
	testNullObject(t, evaluated)
}

func TestAssignment(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"x = 5; x;", 5},
		{"x = 5 * 5; x;", 25},
		{"a = 5; b = a; b;", 5},
		{"a = 5; b = a; c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestAssignmentReturnsNull(t *testing.T) {
	evaluated := testEval("x = 42;")
	testNullObject(t, evaluated)
}

func TestAssignmentOverwrite(t *testing.T) {
	input := `
		x = 10;
		x = 20;
		x;
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 20)
}

func TestAssignmentDifferentTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{"integer", "x = 42; x;", int64(42)},
		{"float", "x = 3.14; x;", float64(3.14)},
		{"string", `x = "hello"; x;`, "hello"},
		{"boolean", "x = true; x;", true},
		{"null", "x = null; x;", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			switch expected := tt.expected.(type) {
			case int64:
				testIntegerObject(t, evaluated, expected)
			case float64:
				testFloatObject(t, evaluated, expected)
			case string:
				testStringObject(t, evaluated, expected)
			case bool:
				testBooleanObject(t, evaluated, expected)
			case nil:
				testNullObject(t, evaluated)
			}
		})
	}
}

func TestIfStatementTruthy(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"x = 0; if (true) { x = 10; } x;", 10},
		{"x = 0; if (5 > 3) { x = 10; } x;", 10},
		{"x = 0; if (1 == 1) { x = 42; } x;", 42},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestIfStatementFalsy(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"x = 5; if (false) { x = 10; } x;", 5},
		{"x = 5; if (3 > 5) { x = 10; } x;", 5},
		{"x = 5; if (1 == 2) { x = 42; } x;", 5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestIfElseStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"x = 0; if (true) { x = 10; } else { x = 20; } x;", 10},
		{"x = 0; if (false) { x = 10; } else { x = 20; } x;", 20},
		{"x = 0; if (5 > 10) { x = 10; } else { x = 20; } x;", 20},
		{"x = 0; if (10 > 5) { x = 10; } else { x = 20; } x;", 10},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestIfStatementReturnsNull(t *testing.T) {
	tests := []string{
		"if (true) { 5; }",
		"if (false) { 5; }",
		"if (true) { 5; } else { 10; }",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			evaluated := testEval(input)
			testNullObject(t, evaluated)
		})
	}
}

func TestIfStatementWithNestedBlocks(t *testing.T) {
	input := `
		x = 0;
		if (true) {
			if (true) {
				x = 42;
			}
		}
		x;
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 42)
}

func TestIfStatementConditionError(t *testing.T) {
	evaluated := testEval("if (undefined_var) { 5; }")
	testErrorObject(t, evaluated, "undefined variable: undefined_var")
}

func TestBlockStatementMultipleStatements(t *testing.T) {
	input := `
		x = 0;
		if (true) {
			x = 1;
			x = x + 1;
			x = x + 1;
		}
		x;
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 3)
}

func TestLogicalOperatorsPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true || false && false;", true},
		{"false || true && true;", true},
		{"false && true || true;", true},
		{"false && false || false;", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestComplexLogicalExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"x = 0; if (5 > 3 && 10 > 5 || false) { x = 42; } x;", 42},
		{"x = 0; if (false || 5 > 3 && 10 > 5) { x = 42; } x;", 42},
		{"x = 0; if (1 == 2 && 3 == 3 || 4 == 4) { x = 42; } x;", 42},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestUndefinedVariable(t *testing.T) {
	evaluated := testEval("foobar;")
	testErrorObject(t, evaluated, "undefined variable: foobar")
}

func TestUndefinedVariableInExpression(t *testing.T) {
	evaluated := testEval("x = 5; x + y;")
	testErrorObject(t, evaluated, "undefined variable: y")
}

func TestAssignmentWithExpressionUsingVariable(t *testing.T) {
	input := `
		x = 10;
		x = x + 5;
		x;
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 15)
}

func TestListLiteralEmpty(t *testing.T) {
	evaluated := testEval("[];")
	list, ok := evaluated.(*List)
	if !ok {
		t.Fatalf("expected *List, got %T", evaluated)
	}
	if len(list.Elements) != 0 {
		t.Errorf("expected 0 elements, got %d", len(list.Elements))
	}
}

func TestListLiteralIntegers(t *testing.T) {
	evaluated := testEval("[1, 2, 3];")
	list, ok := evaluated.(*List)
	if !ok {
		t.Fatalf("expected *List, got %T", evaluated)
	}
	if len(list.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list.Elements))
	}

	testIntegerObject(t, list.Elements[0], 1)
	testIntegerObject(t, list.Elements[1], 2)
	testIntegerObject(t, list.Elements[2], 3)
}

func TestListLiteralMixedTypes(t *testing.T) {
	evaluated := testEval(`[1, "hello", true, null];`)
	list, ok := evaluated.(*List)
	if !ok {
		t.Fatalf("expected *List, got %T", evaluated)
	}
	if len(list.Elements) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(list.Elements))
	}

	testIntegerObject(t, list.Elements[0], 1)
	testStringObject(t, list.Elements[1], "hello")
	testBooleanObject(t, list.Elements[2], true)
	testNullObject(t, list.Elements[3])
}

func TestForStatementBasic(t *testing.T) {
	input := `
		sum = 0;
		for (x in [1, 2, 3]) {
			sum = sum + x;
		}
		sum;
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 6)
}

func TestForStatementEmptyList(t *testing.T) {
	input := `
		sum = 0;
		for (x in []) {
			sum = sum + 1;
		}
		sum;
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 0)
}

func TestForStatementStrings(t *testing.T) {
	input := `
		result = "";
		for (s in ["a", "b", "c"]) {
			result = result + s;
		}
		result;
	`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "abc")
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "simple return",
			input:    "fn test() { return 10; } test();",
			expected: 10,
		},
		{
			name:     "return with expression",
			input:    "fn test() { return 5 + 5; } test();",
			expected: 10,
		},
		{
			name:     "early return",
			input:    "fn test() { return 10; return 20; } test();",
			expected: 10,
		},
		{
			name:     "return in if",
			input:    "fn test(x) { if (x > 5) { return 1; } return 0; } test(10);",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestObjectLiteralEmpty(t *testing.T) {
	evaluated := testEval("{};")
	hash, ok := evaluated.(*Hash)
	if !ok {
		t.Fatalf("expected *Hash, got %T", evaluated)
	}
	if len(hash.Pairs) != 0 {
		t.Errorf("expected 0 pairs, got %d", len(hash.Pairs))
	}
}

func TestObjectLiteralBasic(t *testing.T) {
	evaluated := testEval(`{name: "Alice", age: 30};`)
	hash, ok := evaluated.(*Hash)
	if !ok {
		t.Fatalf("expected *Hash, got %T", evaluated)
	}
	if len(hash.Pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(hash.Pairs))
	}

	nameVal, ok := hash.Get("name")
	if !ok {
		t.Fatal("expected 'name' key to exist")
	}
	testStringObject(t, nameVal, "Alice")

	ageVal, ok := hash.Get("age")
	if !ok {
		t.Fatal("expected 'age' key to exist")
	}
	testIntegerObject(t, ageVal, 30)
}

func TestObjectLiteralMixedTypes(t *testing.T) {
	evaluated := testEval(`{
		str: "hello",
		num: 42,
		float: 3.14,
		bool: true,
		nothing: null
	};`)
	hash, ok := evaluated.(*Hash)
	if !ok {
		t.Fatalf("expected *Hash, got %T", evaluated)
	}
	if len(hash.Pairs) != 5 {
		t.Fatalf("expected 5 pairs, got %d", len(hash.Pairs))
	}

	testStringObject(t, hash.Pairs["str"], "hello")
	testIntegerObject(t, hash.Pairs["num"], 42)
	testFloatObject(t, hash.Pairs["float"], 3.14)
	testBooleanObject(t, hash.Pairs["bool"], true)
	testNullObject(t, hash.Pairs["nothing"])
}

func TestObjectLiteralWithExpressions(t *testing.T) {
	evaluated := testEval(`{
		sum: 5 + 5,
		product: 3 * 4,
		concat: "hello" + " world"
	};`)
	hash, ok := evaluated.(*Hash)
	if !ok {
		t.Fatalf("expected *Hash, got %T", evaluated)
	}

	testIntegerObject(t, hash.Pairs["sum"], 10)
	testIntegerObject(t, hash.Pairs["product"], 12)
	testStringObject(t, hash.Pairs["concat"], "hello world")
}

func TestObjectLiteralWithVariables(t *testing.T) {
	input := `
		x = 10;
		y = "test";
		{a: x, b: y};
	`
	evaluated := testEval(input)
	hash, ok := evaluated.(*Hash)
	if !ok {
		t.Fatalf("expected *Hash, got %T", evaluated)
	}

	testIntegerObject(t, hash.Pairs["a"], 10)
	testStringObject(t, hash.Pairs["b"], "test")
}

func TestObjectLiteralNested(t *testing.T) {
	evaluated := testEval(`{
		outer: {
			inner: 42
		}
	};`)
	hash, ok := evaluated.(*Hash)
	if !ok {
		t.Fatalf("expected *Hash, got %T", evaluated)
	}

	outerVal, ok := hash.Get("outer")
	if !ok {
		t.Fatal("expected 'outer' key to exist")
	}

	innerHash, ok := outerVal.(*Hash)
	if !ok {
		t.Fatalf("expected *Hash, got %T", outerVal)
	}

	innerVal, ok := innerHash.Get("inner")
	if !ok {
		t.Fatal("expected 'inner' key to exist")
	}
	testIntegerObject(t, innerVal, 42)
}

func TestObjectLiteralWithList(t *testing.T) {
	evaluated := testEval(`{items: [1, 2, 3]};`)
	hash, ok := evaluated.(*Hash)
	if !ok {
		t.Fatalf("expected *Hash, got %T", evaluated)
	}

	itemsVal, ok := hash.Get("items")
	if !ok {
		t.Fatal("expected 'items' key to exist")
	}

	list, ok := itemsVal.(*List)
	if !ok {
		t.Fatalf("expected *List, got %T", itemsVal)
	}

	if len(list.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list.Elements))
	}
}

func TestObjectLiteralValueError(t *testing.T) {
	evaluated := testEval(`{key: undefined_var};`)
	testErrorObject(t, evaluated, "undefined variable: undefined_var")
}

func TestObjectLiteralAssignment(t *testing.T) {
	input := `
		obj = {name: "test", value: 100};
		obj;
	`
	evaluated := testEval(input)
	hash, ok := evaluated.(*Hash)
	if !ok {
		t.Fatalf("expected *Hash, got %T", evaluated)
	}

	testStringObject(t, hash.Pairs["name"], "test")
	testIntegerObject(t, hash.Pairs["value"], 100)
}

func TestListIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"[1, 2, 3][0];", 1},
		{"[1, 2, 3][1];", 2},
		{"[1, 2, 3][2];", 3},
		{"x = [10, 20, 30]; x[0];", 10},
		{"x = [10, 20, 30]; x[1];", 20},
		{"x = [10, 20, 30]; x[2];", 30},
		{"[1, 2, 3][1 + 1];", 3},
		{"x = 0; [10, 20, 30][x];", 10},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestListNegativeIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"[1, 2, 3][-1];", 3},
		{"[1, 2, 3][-2];", 2},
		{"[1, 2, 3][-3];", 1},
		{"[10, 20, 30, 40, 50][-1];", 50},
		{"[10, 20, 30, 40, 50][-5];", 10},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestListIndexOutOfBounds(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{"[1, 2, 3][3];", "index out of bounds: 3 (length: 3)"},
		{"[1, 2, 3][10];", "index out of bounds: 10 (length: 3)"},
		{"[1, 2, 3][-4];", "index out of bounds: -4 (length: 3)"},
		{"[][0];", "index out of bounds: 0 (length: 0)"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testErrorObject(t, evaluated, tt.expectedMessage)
		})
	}
}

func TestListIndexMixedTypes(t *testing.T) {
	input := `[1, "hello", true, null][1];`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "hello")
}

func TestListIndexReturnsBoolean(t *testing.T) {
	evaluated := testEval(`[true, false][0];`)
	testBooleanObject(t, evaluated, true)
}

func TestListIndexReturnsNull(t *testing.T) {
	evaluated := testEval(`[null][0];`)
	testNullObject(t, evaluated)
}

func TestStringIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"[0];`, "h"},
		{`"hello"[1];`, "e"},
		{`"hello"[4];`, "o"},
		{`s = "world"; s[0];`, "w"},
		{`s = "test"; s[2];`, "s"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestStringNegativeIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"[-1];`, "o"},
		{`"hello"[-2];`, "l"},
		{`"hello"[-5];`, "h"},
		{`"abc"[-1];`, "c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestStringIndexOutOfBounds(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{`"hello"[5];`, "index out of bounds: 5 (length: 5)"},
		{`"hello"[100];`, "index out of bounds: 100 (length: 5)"},
		{`"hello"[-6];`, "index out of bounds: -6 (length: 5)"},
		{`""[0];`, "index out of bounds: 0 (length: 0)"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testErrorObject(t, evaluated, tt.expectedMessage)
		})
	}
}

func TestHashIndexExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:     "string value",
			input:    `{name: "Alice"}["name"];`,
			expected: "Alice",
		},
		{
			name:     "integer value",
			input:    `{age: 30}["age"];`,
			expected: int64(30),
		},
		{
			name:     "boolean value",
			input:    `{active: true}["active"];`,
			expected: true,
		},
		{
			name:     "with variable",
			input:    `obj = {key: "value"}; obj["key"];`,
			expected: "value",
		},
		{
			name:     "nested access",
			input:    `{outer: {inner: 42}}["outer"];`,
			expected: "hash", // special case - check it's a hash
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			switch expected := tt.expected.(type) {
			case string:
				if expected == "hash" {
					if _, ok := evaluated.(*Hash); !ok {
						t.Errorf("expected *Hash, got %T", evaluated)
					}
				} else {
					testStringObject(t, evaluated, expected)
				}
			case int64:
				testIntegerObject(t, evaluated, expected)
			case bool:
				testBooleanObject(t, evaluated, expected)
			}
		})
	}
}

func TestHashIndexMissingKey(t *testing.T) {
	tests := []string{
		`{}["missing"];`,
		`{name: "Alice"}["age"];`,
		`{a: 1, b: 2}["c"];`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			evaluated := testEval(input)
			testNullObject(t, evaluated)
		})
	}
}

func TestHashIndexWithStringVariable(t *testing.T) {
	input := `
		obj = {name: "test"};
		key = "name";
		obj[key];
	`
	evaluated := testEval(input)
	testStringObject(t, evaluated, "test")
}

func TestHashNestedIndexAccess(t *testing.T) {
	input := `{outer: {inner: 42}}["outer"]["inner"];`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 42)
}

func TestIndexExpressionTypeErrors(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{`[1, 2, 3]["string"];`, "index operator not supported: LIST[STRING]"},
		{`"hello"["string"];`, "index operator not supported: STRING[STRING]"},
		{`{key: "value"}[0];`, "index operator not supported: HASH[INTEGER]"},
		{`42[0];`, "index operator not supported: INTEGER[INTEGER]"},
		{`true[0];`, "index operator not supported: BOOLEAN[INTEGER]"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testErrorObject(t, evaluated, tt.expectedMessage)
		})
	}
}

func TestIndexExpressionLeftError(t *testing.T) {
	evaluated := testEval("undefined_list[0];")
	testErrorObject(t, evaluated, "undefined variable: undefined_list")
}

func TestIndexExpressionIndexError(t *testing.T) {
	evaluated := testEval("[1, 2, 3][undefined_index];")
	testErrorObject(t, evaluated, "undefined variable: undefined_index")
}

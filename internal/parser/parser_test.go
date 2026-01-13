package parser

import (
	"testing"

	"github.com/boattime/awsl/internal/ast"
	"github.com/boattime/awsl/internal/lexer"
)

// parseProgram creates a parser, parses the input, and fails if there are errors.
func parseProgram(t *testing.T, input string) *ast.Program {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if p.HasErrors() {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	return program
}

// parseProgramWithErrors parses input and returns both the program and errors.
func parseProgramWithErrors(t *testing.T, input string) (*ast.Program, []*Error) {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	return program, p.Errors()
}

// requireStatementCount checks that the program has the expected number of statements.
func requireStatementCount(t *testing.T, program *ast.Program, expected int) {
	t.Helper()
	if len(program.Statements) != expected {
		t.Fatalf("expected %d statements, got %d", expected, len(program.Statements))
	}
}

// requireExpressionStatement asserts the statement is an ExpressionStatement
// and returns its expression.
func requireExpressionStatement(t *testing.T, stmt ast.Statement) ast.Expression {
	t.Helper()
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected *ast.ExpressionStatement, got %T", stmt)
	}
	return exprStmt.Expression
}

// testIntegerLiteral checks that an expression is an integer literal with the expected value.
func testIntegerLiteral(t *testing.T, expr ast.Expression, expected int64) {
	t.Helper()
	intLit, ok := expr.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected *ast.IntegerLiteral, got %T", expr)
	}
	if intLit.Value != expected {
		t.Errorf("expected value %d, got %d", expected, intLit.Value)
	}
}

// testIdentifier checks that an expression is an identifier with the expected value.
func testIdentifier(t *testing.T, expr ast.Expression, expected string) {
	t.Helper()
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected *ast.Identifier, got %T", expr)
	}
	if ident.Value != expected {
		t.Errorf("expected value %q, got %q", expected, ident.Value)
	}
}

// testBooleanLiteral checks that an expression is a boolean literal with the expected value.
func testBooleanLiteral(t *testing.T, expr ast.Expression, expected bool) {
	t.Helper()
	boolLit, ok := expr.(*ast.BooleanLiteral)
	if !ok {
		t.Fatalf("expected *ast.BooleanLiteral, got %T", expr)
	}
	if boolLit.Value != expected {
		t.Errorf("expected value %t, got %t", expected, boolLit.Value)
	}
}

// testStringLiteral checks that an expression is a string literal with the expected value.
func testStringLiteral(t *testing.T, expr ast.Expression, expected string) {
	t.Helper()
	strLit, ok := expr.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected *ast.StringLiteral, got %T", expr)
	}
	if strLit.Value != expected {
		t.Errorf("expected value %q, got %q", expected, strLit.Value)
	}
}

func TestIntegerLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"42;", 42},
		{"0;", 0},
		{"12345;", 12345},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			requireStatementCount(t, program, 1)

			expr := requireExpressionStatement(t, program.Statements[0])
			testIntegerLiteral(t, expr, tt.expected)
		})
	}
}

func TestFloatLiteral(t *testing.T) {
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
			program := parseProgram(t, tt.input)
			requireStatementCount(t, program, 1)

			expr := requireExpressionStatement(t, program.Statements[0])
			floatLit, ok := expr.(*ast.FloatLiteral)
			if !ok {
				t.Fatalf("expected *ast.FloatLiteral, got %T", expr)
			}
			if floatLit.Value != tt.expected {
				t.Errorf("expected value %f, got %f", tt.expected, floatLit.Value)
			}
		})
	}
}

func TestStringLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello";`, "hello"},
		{`"us-west-2";`, "us-west-2"},
		{`"";`, ""},
		{`"with spaces";`, "with spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			requireStatementCount(t, program, 1)

			expr := requireExpressionStatement(t, program.Statements[0])
			testStringLiteral(t, expr, tt.expected)
		})
	}
}

func TestBooleanLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			requireStatementCount(t, program, 1)

			expr := requireExpressionStatement(t, program.Statements[0])
			testBooleanLiteral(t, expr, tt.expected)
		})
	}
}

func TestNullLiteral(t *testing.T) {
	program := parseProgram(t, "null;")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	_, ok := expr.(*ast.NullLiteral)
	if !ok {
		t.Fatalf("expected *ast.NullLiteral, got %T", expr)
	}
}

func TestIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"foo;", "foo"},
		{"bar_baz;", "bar_baz"},
		{"myVar123;", "myVar123"},
		{"_private;", "_private"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			requireStatementCount(t, program, 1)

			expr := requireExpressionStatement(t, program.Statements[0])
			testIdentifier(t, expr, tt.expected)
		})
	}
}

func TestPrefixExpression(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		value    int64
	}{
		{"-5;", "-", 5},
		{"-15;", "-", 15},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			requireStatementCount(t, program, 1)

			expr := requireExpressionStatement(t, program.Statements[0])
			prefix, ok := expr.(*ast.PrefixExpression)
			if !ok {
				t.Fatalf("expected *ast.PrefixExpression, got %T", expr)
			}

			if prefix.Operator != tt.operator {
				t.Errorf("expected operator %q, got %q", tt.operator, prefix.Operator)
			}

			testIntegerLiteral(t, prefix.Right, tt.value)
		})
	}
}

func TestBangPrefixExpression(t *testing.T) {
	program := parseProgram(t, "!true;")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	prefix, ok := expr.(*ast.PrefixExpression)
	if !ok {
		t.Fatalf("expected *ast.PrefixExpression, got %T", expr)
	}

	if prefix.Operator != "!" {
		t.Errorf("expected operator '!', got %q", prefix.Operator)
	}

	testBooleanLiteral(t, prefix.Right, true)
}

func TestInfixExpression(t *testing.T) {
	tests := []struct {
		input      string
		leftValue  int64
		operator   string
		rightValue int64
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"5 >= 5;", 5, ">=", 5},
		{"5 <= 5;", 5, "<=", 5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			requireStatementCount(t, program, 1)

			expr := requireExpressionStatement(t, program.Statements[0])
			infix, ok := expr.(*ast.InfixExpression)
			if !ok {
				t.Fatalf("expected *ast.InfixExpression, got %T", expr)
			}

			testIntegerLiteral(t, infix.Left, tt.leftValue)

			if infix.Operator != tt.operator {
				t.Errorf("expected operator %q, got %q", tt.operator, infix.Operator)
			}

			testIntegerLiteral(t, infix.Right, tt.rightValue)
		})
	}
}

func TestInfixExpressionWithBooleans(t *testing.T) {
	tests := []struct {
		input      string
		leftValue  bool
		operator   string
		rightValue bool
	}{
		{"true == true;", true, "==", true},
		{"true != false;", true, "!=", false},
		{"false == false;", false, "==", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			requireStatementCount(t, program, 1)

			expr := requireExpressionStatement(t, program.Statements[0])
			infix, ok := expr.(*ast.InfixExpression)
			if !ok {
				t.Fatalf("expected *ast.InfixExpression, got %T", expr)
			}

			testBooleanLiteral(t, infix.Left, tt.leftValue)

			if infix.Operator != tt.operator {
				t.Errorf("expected operator %q, got %q", tt.operator, infix.Operator)
			}

			testBooleanLiteral(t, infix.Right, tt.rightValue)
		})
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b;", "((-a) * b)"},
		{"!-a;", "(!(-a))"},
		{"a + b + c;", "((a + b) + c)"},
		{"a + b - c;", "((a + b) - c)"},
		{"a * b * c;", "((a * b) * c)"},
		{"a * b / c;", "((a * b) / c)"},
		{"a + b / c;", "(a + (b / c))"},
		{"a + b * c + d / e - f;", "(((a + (b * c)) + (d / e)) - f)"},
		{"5 > 4 == 3 < 4;", "((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4;", "((5 < 4) != (3 > 4))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5;", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"1 + (2 + 3) + 4;", "((1 + ((2 + 3))) + 4)"},
		{"(5 + 5) * 2;", "(((5 + 5)) * 2)"},
		{"2 / (5 + 5);", "(2 / ((5 + 5)))"},
		{"-(5 + 5);", "(-((5 + 5)))"},
		{"!(true == true);", "(!((true == true)))"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			requireStatementCount(t, program, 1)

			expr := requireExpressionStatement(t, program.Statements[0])
			actual := expr.String()

			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestGroupedExpression(t *testing.T) {
	program := parseProgram(t, "(1 + 2);")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	grouped, ok := expr.(*ast.GroupedExpression)
	if !ok {
		t.Fatalf("expected *ast.GroupedExpression, got %T", expr)
	}

	infix, ok := grouped.Expression.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected *ast.InfixExpression, got %T", grouped.Expression)
	}

	testIntegerLiteral(t, infix.Left, 1)
	if infix.Operator != "+" {
		t.Errorf("expected operator '+', got %q", infix.Operator)
	}
	testIntegerLiteral(t, infix.Right, 2)
}

func TestListLiteralEmpty(t *testing.T) {
	program := parseProgram(t, "[];")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	list, ok := expr.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("expected *ast.ListLiteral, got %T", expr)
	}

	if len(list.Elements) != 0 {
		t.Errorf("expected 0 elements, got %d", len(list.Elements))
	}
}

func TestListLiteral(t *testing.T) {
	program := parseProgram(t, "[1, 2, 3];")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	list, ok := expr.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("expected *ast.ListLiteral, got %T", expr)
	}

	if len(list.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list.Elements))
	}

	testIntegerLiteral(t, list.Elements[0], 1)
	testIntegerLiteral(t, list.Elements[1], 2)
	testIntegerLiteral(t, list.Elements[2], 3)
}

func TestListLiteralMixedTypes(t *testing.T) {
	program := parseProgram(t, `[1, "two", true];`)
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	list, ok := expr.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("expected *ast.ListLiteral, got %T", expr)
	}

	if len(list.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list.Elements))
	}

	testIntegerLiteral(t, list.Elements[0], 1)
	testStringLiteral(t, list.Elements[1], "two")
	testBooleanLiteral(t, list.Elements[2], true)
}

func TestObjectLiteralEmpty(t *testing.T) {
	program := parseProgram(t, "{};")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	obj, ok := expr.(*ast.ObjectLiteral)
	if !ok {
		t.Fatalf("expected *ast.ObjectLiteral, got %T", expr)
	}

	if len(obj.Pairs) != 0 {
		t.Errorf("expected 0 pairs, got %d", len(obj.Pairs))
	}
}

func TestObjectLiteral(t *testing.T) {
	program := parseProgram(t, `{name: "test", count: 5};`)
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	obj, ok := expr.(*ast.ObjectLiteral)
	if !ok {
		t.Fatalf("expected *ast.ObjectLiteral, got %T", expr)
	}

	if len(obj.Pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(obj.Pairs))
	}

	// Check first pair
	if obj.Pairs[0].Key.Value != "name" {
		t.Errorf("expected key 'name', got %q", obj.Pairs[0].Key.Value)
	}
	testStringLiteral(t, obj.Pairs[0].Value, "test")

	// Check second pair
	if obj.Pairs[1].Key.Value != "count" {
		t.Errorf("expected key 'count', got %q", obj.Pairs[1].Key.Value)
	}
	testIntegerLiteral(t, obj.Pairs[1].Value, 5)
}

func TestObjectLiteralNested(t *testing.T) {
	program := parseProgram(t, `{outer: {inner: 42}};`)
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	obj, ok := expr.(*ast.ObjectLiteral)
	if !ok {
		t.Fatalf("expected *ast.ObjectLiteral, got %T", expr)
	}

	if len(obj.Pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(obj.Pairs))
	}

	innerObj, ok := obj.Pairs[0].Value.(*ast.ObjectLiteral)
	if !ok {
		t.Fatalf("expected nested *ast.ObjectLiteral, got %T", obj.Pairs[0].Value)
	}

	if len(innerObj.Pairs) != 1 {
		t.Fatalf("expected 1 inner pair, got %d", len(innerObj.Pairs))
	}

	testIntegerLiteral(t, innerObj.Pairs[0].Value, 42)
}

func TestIndexExpression(t *testing.T) {
	program := parseProgram(t, "items[0];")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	indexExpr, ok := expr.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expected *ast.IndexExpression, got %T", expr)
	}

	testIdentifier(t, indexExpr.Left, "items")
	testIntegerLiteral(t, indexExpr.Index, 0)
}

func TestIndexExpressionWithExpression(t *testing.T) {
	program := parseProgram(t, "items[i + 1];")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	indexExpr, ok := expr.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expected *ast.IndexExpression, got %T", expr)
	}

	testIdentifier(t, indexExpr.Left, "items")

	infix, ok := indexExpr.Index.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected *ast.InfixExpression, got %T", indexExpr.Index)
	}

	testIdentifier(t, infix.Left, "i")
	testIntegerLiteral(t, infix.Right, 1)
}

func TestMemberExpression(t *testing.T) {
	program := parseProgram(t, "user.name;")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	memberExpr, ok := expr.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected *ast.MemberExpression, got %T", expr)
	}

	testIdentifier(t, memberExpr.Object, "user")
	if memberExpr.Member.Value != "name" {
		t.Errorf("expected member 'name', got %q", memberExpr.Member.Value)
	}
}

func TestMemberExpressionChained(t *testing.T) {
	program := parseProgram(t, "config.lambda.memory;")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	memberExpr, ok := expr.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected *ast.MemberExpression, got %T", expr)
	}

	// Outer: (config.lambda).memory
	if memberExpr.Member.Value != "memory" {
		t.Errorf("expected member 'memory', got %q", memberExpr.Member.Value)
	}

	// Inner: config.lambda
	innerMember, ok := memberExpr.Object.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected inner *ast.MemberExpression, got %T", memberExpr.Object)
	}

	testIdentifier(t, innerMember.Object, "config")
	if innerMember.Member.Value != "lambda" {
		t.Errorf("expected member 'lambda', got %q", innerMember.Member.Value)
	}
}

func TestCallExpressionNoArgs(t *testing.T) {
	program := parseProgram(t, "foo();")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	callExpr, ok := expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected *ast.CallExpression, got %T", expr)
	}

	testIdentifier(t, callExpr.Function, "foo")
	if len(callExpr.Arguments) != 0 {
		t.Errorf("expected 0 arguments, got %d", len(callExpr.Arguments))
	}
}

func TestCallExpressionWithArgs(t *testing.T) {
	program := parseProgram(t, `print("hello", 42);`)
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	callExpr, ok := expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected *ast.CallExpression, got %T", expr)
	}

	testIdentifier(t, callExpr.Function, "print")
	if len(callExpr.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(callExpr.Arguments))
	}

	// Both should be positional
	if callExpr.Arguments[0].Name != nil {
		t.Error("expected first argument to be positional")
	}
	testStringLiteral(t, callExpr.Arguments[0].Value, "hello")

	if callExpr.Arguments[1].Name != nil {
		t.Error("expected second argument to be positional")
	}
	testIntegerLiteral(t, callExpr.Arguments[1].Value, 42)
}

func TestCallExpressionNamedArgs(t *testing.T) {
	program := parseProgram(t, `lambda.list(runtime: "python3.12");`)
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	callExpr, ok := expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected *ast.CallExpression, got %T", expr)
	}

	// Function should be a member expression: lambda.list
	memberExpr, ok := callExpr.Function.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected *ast.MemberExpression, got %T", callExpr.Function)
	}
	testIdentifier(t, memberExpr.Object, "lambda")
	if memberExpr.Member.Value != "list" {
		t.Errorf("expected member 'list', got %q", memberExpr.Member.Value)
	}

	if len(callExpr.Arguments) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(callExpr.Arguments))
	}

	// Should be named argument
	if callExpr.Arguments[0].Name == nil {
		t.Fatal("expected named argument")
	}
	if callExpr.Arguments[0].Name.Value != "runtime" {
		t.Errorf("expected argument name 'runtime', got %q", callExpr.Arguments[0].Name.Value)
	}
	testStringLiteral(t, callExpr.Arguments[0].Value, "python3.12")
}

func TestCallExpressionMixedArgs(t *testing.T) {
	program := parseProgram(t, `invoke("func", payload: data);`)
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	callExpr, ok := expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected *ast.CallExpression, got %T", expr)
	}

	if len(callExpr.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(callExpr.Arguments))
	}

	// First: positional
	if callExpr.Arguments[0].Name != nil {
		t.Error("expected first argument to be positional")
	}
	testStringLiteral(t, callExpr.Arguments[0].Value, "func")

	// Second: named
	if callExpr.Arguments[1].Name == nil {
		t.Fatal("expected second argument to be named")
	}
	if callExpr.Arguments[1].Name.Value != "payload" {
		t.Errorf("expected argument name 'payload', got %q", callExpr.Arguments[1].Name.Value)
	}
	testIdentifier(t, callExpr.Arguments[1].Value, "data")
}

func TestCallExpressionChained(t *testing.T) {
	program := parseProgram(t, "foo()();")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	outerCall, ok := expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected *ast.CallExpression, got %T", expr)
	}

	innerCall, ok := outerCall.Function.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected inner *ast.CallExpression, got %T", outerCall.Function)
	}

	testIdentifier(t, innerCall.Function, "foo")
}

func TestPipeExpressionCSV(t *testing.T) {
	program := parseProgram(t, "items | format csv;")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	pipeExpr, ok := expr.(*ast.PipeExpression)
	if !ok {
		t.Fatalf("expected *ast.PipeExpression, got %T", expr)
	}

	testIdentifier(t, pipeExpr.Left, "items")
	if pipeExpr.Format != "csv" {
		t.Errorf("expected format 'csv', got %q", pipeExpr.Format)
	}
}

func TestPipeExpressionTable(t *testing.T) {
	program := parseProgram(t, "data | format table;")
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	pipeExpr, ok := expr.(*ast.PipeExpression)
	if !ok {
		t.Fatalf("expected *ast.PipeExpression, got %T", expr)
	}

	testIdentifier(t, pipeExpr.Left, "data")
	if pipeExpr.Format != "table" {
		t.Errorf("expected format 'table', got %q", pipeExpr.Format)
	}
}

func TestAssignmentStatement(t *testing.T) {
	tests := []struct {
		input         string
		expectedName  string
		expectedValue int64
	}{
		{"x = 5;", "x", 5},
		{"count = 42;", "count", 42},
		{"_value = 0;", "_value", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			requireStatementCount(t, program, 1)

			stmt, ok := program.Statements[0].(*ast.AssignmentStatement)
			if !ok {
				t.Fatalf("expected *ast.AssignmentStatement, got %T", program.Statements[0])
			}

			if stmt.Name.Value != tt.expectedName {
				t.Errorf("expected name %q, got %q", tt.expectedName, stmt.Name.Value)
			}

			testIntegerLiteral(t, stmt.Value, tt.expectedValue)
		})
	}
}

func TestAssignmentStatementWithExpression(t *testing.T) {
	program := parseProgram(t, "result = a + b;")
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("expected *ast.AssignmentStatement, got %T", program.Statements[0])
	}

	if stmt.Name.Value != "result" {
		t.Errorf("expected name 'result', got %q", stmt.Name.Value)
	}

	infix, ok := stmt.Value.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected *ast.InfixExpression, got %T", stmt.Value)
	}

	testIdentifier(t, infix.Left, "a")
	testIdentifier(t, infix.Right, "b")
}

func TestContextStatementProfile(t *testing.T) {
	program := parseProgram(t, `profile "production";`)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ContextStatement)
	if !ok {
		t.Fatalf("expected *ast.ContextStatement, got %T", program.Statements[0])
	}

	if stmt.Token.Literal != "profile" {
		t.Errorf("expected token literal 'profile', got %q", stmt.Token.Literal)
	}

	if stmt.Value != "production" {
		t.Errorf("expected value 'production', got %q", stmt.Value)
	}
}

func TestContextStatementRegion(t *testing.T) {
	program := parseProgram(t, `region "us-west-2";`)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ContextStatement)
	if !ok {
		t.Fatalf("expected *ast.ContextStatement, got %T", program.Statements[0])
	}

	if stmt.Token.Literal != "region" {
		t.Errorf("expected token literal 'region', got %q", stmt.Token.Literal)
	}

	if stmt.Value != "us-west-2" {
		t.Errorf("expected value 'us-west-2', got %q", stmt.Value)
	}
}

func TestIfStatement(t *testing.T) {
	program := parseProgram(t, `if (x > 5) { y; }`)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected *ast.IfStatement, got %T", program.Statements[0])
	}

	// Check condition
	infix, ok := stmt.Condition.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected *ast.InfixExpression, got %T", stmt.Condition)
	}
	testIdentifier(t, infix.Left, "x")
	if infix.Operator != ">" {
		t.Errorf("expected operator '>', got %q", infix.Operator)
	}
	testIntegerLiteral(t, infix.Right, 5)

	// Check consequence
	if len(stmt.Consequence.Statements) != 1 {
		t.Fatalf("expected 1 consequence statement, got %d", len(stmt.Consequence.Statements))
	}

	// No alternative
	if stmt.Alternative != nil {
		t.Error("expected no alternative")
	}
}

func TestIfElseStatement(t *testing.T) {
	program := parseProgram(t, `if (x) { a; } else { b; }`)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected *ast.IfStatement, got %T", program.Statements[0])
	}

	// Check condition
	testIdentifier(t, stmt.Condition, "x")

	// Check consequence
	if len(stmt.Consequence.Statements) != 1 {
		t.Fatalf("expected 1 consequence statement, got %d", len(stmt.Consequence.Statements))
	}

	// Check alternative
	if stmt.Alternative == nil {
		t.Fatal("expected alternative")
	}
	if len(stmt.Alternative.Statements) != 1 {
		t.Fatalf("expected 1 alternative statement, got %d", len(stmt.Alternative.Statements))
	}
}

func TestForStatement(t *testing.T) {
	program := parseProgram(t, `for (item in items) { print(item); }`)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("expected *ast.ForStatement, got %T", program.Statements[0])
	}

	if stmt.Iterator.Value != "item" {
		t.Errorf("expected iterator 'item', got %q", stmt.Iterator.Value)
	}

	testIdentifier(t, stmt.Iterable, "items")

	if len(stmt.Body.Statements) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(stmt.Body.Statements))
	}
}

func TestForStatementWithListLiteral(t *testing.T) {
	program := parseProgram(t, `for (i in [1, 2, 3]) { x; }`)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("expected *ast.ForStatement, got %T", program.Statements[0])
	}

	if stmt.Iterator.Value != "i" {
		t.Errorf("expected iterator 'i', got %q", stmt.Iterator.Value)
	}

	list, ok := stmt.Iterable.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("expected *ast.ListLiteral, got %T", stmt.Iterable)
	}

	if len(list.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(list.Elements))
	}
}

func TestReturnStatement(t *testing.T) {
	program := parseProgram(t, `return 42;`)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected *ast.ReturnStatement, got %T", program.Statements[0])
	}

	testIntegerLiteral(t, stmt.Value, 42)
}

func TestReturnStatementBare(t *testing.T) {
	program := parseProgram(t, `return;`)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected *ast.ReturnStatement, got %T", program.Statements[0])
	}

	if stmt.Value != nil {
		t.Error("expected nil value for bare return")
	}
}

func TestReturnStatementWithExpression(t *testing.T) {
	program := parseProgram(t, `return a + b;`)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected *ast.ReturnStatement, got %T", program.Statements[0])
	}

	infix, ok := stmt.Value.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected *ast.InfixExpression, got %T", stmt.Value)
	}

	testIdentifier(t, infix.Left, "a")
	testIdentifier(t, infix.Right, "b")
}

func TestFunctionDeclarationNoParams(t *testing.T) {
	program := parseProgram(t, `fn greet() { print("hello"); }`)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.FunctionDeclaration)
	if !ok {
		t.Fatalf("expected *ast.FunctionDeclaration, got %T", program.Statements[0])
	}

	if stmt.Name.Value != "greet" {
		t.Errorf("expected name 'greet', got %q", stmt.Name.Value)
	}

	if len(stmt.Parameters) != 0 {
		t.Errorf("expected 0 parameters, got %d", len(stmt.Parameters))
	}

	if len(stmt.Body.Statements) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(stmt.Body.Statements))
	}
}

func TestFunctionDeclarationWithParams(t *testing.T) {
	program := parseProgram(t, `fn add(a, b) { return a + b; }`)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.FunctionDeclaration)
	if !ok {
		t.Fatalf("expected *ast.FunctionDeclaration, got %T", program.Statements[0])
	}

	if stmt.Name.Value != "add" {
		t.Errorf("expected name 'add', got %q", stmt.Name.Value)
	}

	if len(stmt.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(stmt.Parameters))
	}

	if stmt.Parameters[0].Value != "a" {
		t.Errorf("expected first param 'a', got %q", stmt.Parameters[0].Value)
	}
	if stmt.Parameters[1].Value != "b" {
		t.Errorf("expected second param 'b', got %q", stmt.Parameters[1].Value)
	}
}

func TestFunctionDeclarationSingleParam(t *testing.T) {
	program := parseProgram(t, `fn double(x) { return x * 2; }`)
	requireStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.FunctionDeclaration)
	if !ok {
		t.Fatalf("expected *ast.FunctionDeclaration, got %T", program.Statements[0])
	}

	if len(stmt.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(stmt.Parameters))
	}

	if stmt.Parameters[0].Value != "x" {
		t.Errorf("expected param 'x', got %q", stmt.Parameters[0].Value)
	}
}

func TestComplexMemberCallChain(t *testing.T) {
	program := parseProgram(t, `users_table.query(pk: "ORG#acme");`)
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])
	callExpr, ok := expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected *ast.CallExpression, got %T", expr)
	}

	memberExpr, ok := callExpr.Function.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected *ast.MemberExpression, got %T", callExpr.Function)
	}

	testIdentifier(t, memberExpr.Object, "users_table")
	if memberExpr.Member.Value != "query" {
		t.Errorf("expected member 'query', got %q", memberExpr.Member.Value)
	}

	if len(callExpr.Arguments) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(callExpr.Arguments))
	}

	if callExpr.Arguments[0].Name.Value != "pk" {
		t.Errorf("expected argument name 'pk', got %q", callExpr.Arguments[0].Name.Value)
	}
}

func TestComplexPostfixChain(t *testing.T) {
	program := parseProgram(t, `obj.list()[0].name;`)
	requireStatementCount(t, program, 1)

	expr := requireExpressionStatement(t, program.Statements[0])

	// Should be: ((((obj).list)())[0]).name
	outerMember, ok := expr.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected *ast.MemberExpression, got %T", expr)
	}
	if outerMember.Member.Value != "name" {
		t.Errorf("expected member 'name', got %q", outerMember.Member.Value)
	}

	indexExpr, ok := outerMember.Object.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expected *ast.IndexExpression, got %T", outerMember.Object)
	}
	testIntegerLiteral(t, indexExpr.Index, 0)

	callExpr, ok := indexExpr.Left.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected *ast.CallExpression, got %T", indexExpr.Left)
	}

	innerMember, ok := callExpr.Function.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected *ast.MemberExpression, got %T", callExpr.Function)
	}

	testIdentifier(t, innerMember.Object, "obj")
	if innerMember.Member.Value != "list" {
		t.Errorf("expected member 'list', got %q", innerMember.Member.Value)
	}
}

func TestMultipleStatements(t *testing.T) {
	input := `
		profile "production";
		region "us-west-2";
		x = 42;
	`
	program := parseProgram(t, input)
	requireStatementCount(t, program, 3)

	// Check types
	_, ok := program.Statements[0].(*ast.ContextStatement)
	if !ok {
		t.Errorf("expected *ast.ContextStatement, got %T", program.Statements[0])
	}

	_, ok = program.Statements[1].(*ast.ContextStatement)
	if !ok {
		t.Errorf("expected *ast.ContextStatement, got %T", program.Statements[1])
	}

	_, ok = program.Statements[2].(*ast.AssignmentStatement)
	if !ok {
		t.Errorf("expected *ast.AssignmentStatement, got %T", program.Statements[2])
	}
}

func TestASTString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x = 5;", "x = 5;"},
		{`profile "prod";`, `profile "prod";`},
		{`region "us-west-2";`, `region "us-west-2";`},
		{"return;", "return;"},
		{"return 42;", "return 42;"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			actual := program.String()
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestPositionTracking(t *testing.T) {
	input := `x = 5;
y = 10;`
	program := parseProgram(t, input)
	requireStatementCount(t, program, 2)

	// First statement at line 1
	pos := program.Statements[0].Pos()
	if pos.Line != 1 {
		t.Errorf("expected line 1, got %d", pos.Line)
	}
	if pos.Column != 1 {
		t.Errorf("expected column 1, got %d", pos.Column)
	}

	// Second statement at line 2
	pos = program.Statements[1].Pos()
	if pos.Line != 2 {
		t.Errorf("expected line 2, got %d", pos.Line)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCount int
		errorContains string
	}{
		{
			name:          "missing semicolon",
			input:         "x = 5",
			expectedCount: 1,
			errorContains: "expected ;",
		},
		{
			name:          "missing closing paren",
			input:         "foo(;",
			expectedCount: 1,
			errorContains: "expected )",
		},
		{
			name:          "missing closing bracket",
			input:         "[1, 2;",
			expectedCount: 1,
			errorContains: "expected ]",
		},
		{
			name:          "missing closing brace",
			input:         "{x: 1;",
			expectedCount: 1,
			errorContains: "expected }",
		},
		{
			name:          "invalid pipe format",
			input:         "x | format json;",
			expectedCount: 1,
			errorContains: "expected 'csv' or 'table'",
		},
		{
			name:          "missing format keyword",
			input:         "x | csv;",
			expectedCount: 1,
			errorContains: "expected 'format'",
		},
		{
			name:          "if missing paren",
			input:         "if x { y; }",
			expectedCount: 1,
			errorContains: "expected (",
		},
		{
			name:          "for missing in",
			input:         "for (x items) { y; }",
			expectedCount: 1,
			errorContains: "expected IN",
		},
		{
			name:          "function missing name",
			input:         "fn () { x; }",
			expectedCount: 1,
			errorContains: "expected IDENT",
		},
		{
			name:          "context missing string",
			input:         "profile production;",
			expectedCount: 1,
			errorContains: "expected STRING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := parseProgramWithErrors(t, tt.input)

			if len(errors) < tt.expectedCount {
				t.Fatalf("expected at least %d errors, got %d", tt.expectedCount, len(errors))
			}

			found := false
			for _, err := range errors {
				if contains(err.Message, tt.errorContains) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected error containing %q, got errors: %v", tt.errorContains, errors)
			}
		})
	}
}

func TestErrorRecovery(t *testing.T) {
	// Parser should recover and parse multiple statements even with errors
	input := `
		x = ;
		y = 5;
		z = ;
		w = 10;
	`
	program, errors := parseProgramWithErrors(t, input)

	// Should have some errors
	if len(errors) == 0 {
		t.Error("expected errors")
	}

	// Should still parse valid statements
	if len(program.Statements) < 2 {
		t.Errorf("expected at least 2 statements after recovery, got %d", len(program.Statements))
	}
}

func TestMaxErrors(t *testing.T) {
	// Generate input that would produce many errors
	input := ""
	for i := 0; i < 30; i++ {
		input += "@ "
	}

	_, errors := parseProgramWithErrors(t, input)

	if len(errors) > MaxErrors {
		t.Errorf("expected at most %d errors, got %d", MaxErrors, len(errors))
	}
}

// contains checks if s contains substr (simple helper to avoid importing strings)
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestEmptyProgram(t *testing.T) {
	program := parseProgram(t, "")
	if len(program.Statements) != 0 {
		t.Errorf("expected 0 statements, got %d", len(program.Statements))
	}
}

func TestWhitespaceOnlyProgram(t *testing.T) {
	program := parseProgram(t, "   \n\t\n   ")
	if len(program.Statements) != 0 {
		t.Errorf("expected 0 statements, got %d", len(program.Statements))
	}
}

func TestCommentsIgnored(t *testing.T) {
	input := `
		// This is a comment
		x = 5; // inline comment
		// Another comment
		y = 10;
	`
	program := parseProgram(t, input)
	requireStatementCount(t, program, 2)
}

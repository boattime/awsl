package eval

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/boattime/awsl/internal/lexer"
	"github.com/boattime/awsl/internal/parser"
)

// testEvalWithBuiltins creates an environment with builtins registered.
func testEvalWithBuiltins(input string, stdout io.Writer) Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := NewEnvironment(stdout)
	RegisterBuiltins(env)
	return Eval(program, env)
}

func TestBuiltinPrintReturnsNull(t *testing.T) {
	var stdout bytes.Buffer
	result := testEvalWithBuiltins(`print("hello");`, &stdout)
	testStdout(t, stdout, "hello\n")
	testNullObject(t, result)
}

func TestBuiltinPrintMultipleArgs(t *testing.T) {
	var stdout bytes.Buffer
	result := testEvalWithBuiltins(`print("hello", "world", 42);`, &stdout)
	testStdout(t, stdout, "hello world 42\n")
	testNullObject(t, result)
}

func TestBuiltinPrintNoArgs(t *testing.T) {
	var stdout bytes.Buffer
	result := testEvalWithBuiltins(`print();`, &stdout)
	testStdout(t, stdout, "\n")
	testNullObject(t, result)
}

func TestBuiltinPrintWithVariable(t *testing.T) {
	var stdout bytes.Buffer
	result := testEvalWithBuiltins(`x = 42; print(x);`, &stdout)
	testStdout(t, stdout, "42\n")
	testNullObject(t, result)
}

func TestBuiltinPrintWithExpression(t *testing.T) {
	var stdout bytes.Buffer
	result := testEvalWithBuiltins(`print(5 + 3);`, &stdout)
	testStdout(t, stdout, "8\n")
	testNullObject(t, result)
}

func TestRegisterBuiltins(t *testing.T) {
	env := NewEnvironment(os.Stdout)
	RegisterBuiltins(env)

	// Check print is registered
	val, ok := env.Get("print")
	if !ok {
		t.Fatal("expected 'print' to be registered")
	}

	builtin, ok := val.(*Builtin)
	if !ok {
		t.Fatalf("expected *Builtin, got %T", val)
	}

	if builtin.Name != "print" {
		t.Errorf("expected name 'print', got %q", builtin.Name)
	}
}

// testStdout checks that bytes is the expected value.
func testStdout(t *testing.T, stdout bytes.Buffer, expected string) bool {
	t.Helper()

	result := stdout.String()
	if result == "<nil>" {
		t.Errorf("Buffer is nill")
		return false
	}

	if result != expected {
		t.Errorf("stdout has wrong value. got=%q, want=%q", result, expected)
		return false
	}

	return true
}

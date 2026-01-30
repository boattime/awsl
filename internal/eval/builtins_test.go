package eval

import (
	"bytes"
	"os"
	"testing"

	"github.com/boattime/awsl/internal/lexer"
	"github.com/boattime/awsl/internal/parser"
)

// testEvalWithBuiltins creates an environment with builtins registered.
func testEvalWithBuiltins(input string) Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	var stdout bytes.Buffer
	env := NewEnvironment(&stdout)
	RegisterBuiltins(env)
	return Eval(program, env)
}

func TestBuiltinPrintReturnsNull(t *testing.T) {
	result := testEvalWithBuiltins(`print("hello");`)
	testNullObject(t, result)
}

func TestBuiltinPrintMultipleArgs(t *testing.T) {
	result := testEvalWithBuiltins(`print("hello", "world", 42);`)
	testNullObject(t, result)
}

func TestBuiltinPrintNoArgs(t *testing.T) {
	result := testEvalWithBuiltins(`print();`)
	testNullObject(t, result)
}

func TestBuiltinPrintWithVariable(t *testing.T) {
	result := testEvalWithBuiltins(`x = 42; print(x);`)
	testNullObject(t, result)
}

func TestBuiltinPrintWithExpression(t *testing.T) {
	result := testEvalWithBuiltins(`print(5 + 3);`)
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

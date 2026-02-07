// Package object defines the runtime value types for the AWSL interpreter.
// These objects are produced by the evaluator when executing AWSL programs.
package eval

import (
	"fmt"
	"strings"

	"github.com/boattime/awsl/internal/ast"
)

// ObjectType represents the type of a runtime object as a string.
type ObjectType string

// Object types.
const (
	INTEGER_OBJ      = "INTEGER"
	FLOAT_OBJ        = "FLOAT"
	STRING_OBJ       = "STRING"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	ERROR_OBJ        = "ERROR"
	BUILTIN_OBJ      = "BUILTIN"
	LIST_OBJ         = "LIST"
	FUNCTION_OBJ     = "FUNCTION"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	HASH_OBJ         = "HASH"
)

// Object is the interface that all runtime values implement.
type Object interface {
	// Type returns the type of the object.
	Type() ObjectType

	// Inspect returns a string representation of the object for debugging.
	Inspect() string
}

// Singleton objects for boolean and null values.
// These are reused rather than allocating new objects each time.
var (
	TRUE  = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}
	NULL  = &Null{}
)

// Integer represents an integer value at runtime.
type Integer struct {
	Value int64
}

// Type returns INTEGER_OBJ.
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

// Inspect returns the integer as a string.
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

// Float represents a floating-point value at runtime.
type Float struct {
	Value float64
}

// Type returns FLOAT_OBJ.
func (f *Float) Type() ObjectType { return FLOAT_OBJ }

// Inspect returns the float as a string.
func (f *Float) Inspect() string { return fmt.Sprintf("%g", f.Value) }

// String represents a string value at runtime.
type String struct {
	Value string
}

// Type returns STRING_OBJ.
func (s *String) Type() ObjectType { return STRING_OBJ }

// Inspect returns the string value.
func (s *String) Inspect() string { return s.Value }

// Boolean represents a boolean value at runtime.
// Use the TRUE and FALSE singletons rather than creating new instances.
type Boolean struct {
	Value bool
}

// Type returns BOOLEAN_OBJ.
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

// Inspect returns "true" or "false".
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// Null represents the absence of a value.
// Use the NULL singleton rather than creating new instances.
type Null struct{}

// Type returns NULL_OBJ.
func (n *Null) Type() ObjectType { return NULL_OBJ }

// Inspect returns "null".
func (n *Null) Inspect() string { return "null" }

// Error represents a runtime error with position information.
type Error struct {
	Message string
	Line    int
	Column  int
}

// Type returns ERROR_OBJ.
func (e *Error) Type() ObjectType { return ERROR_OBJ }

// Inspect returns the formatted error message with position.
func (e *Error) Inspect() string {
	return fmt.Sprintf("error at line %d, column %d: %s", e.Line, e.Column, e.Message)
}

// BuiltinFunction is the signature for built-in functions.
type BuiltinFunction func(env *Environment, args ...Object) Object

// Builtin wraps a Go function as an AWSL callable object.
type Builtin struct {
	Name string
	Fn   BuiltinFunction
}

// Type returns BUILTIN_OBJ.
func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }

// Inspect returns the builtin function name.
func (b *Builtin) Inspect() string { return "builtin:" + b.Name }

// List represents a list/array value at runtime.
type List struct {
	Elements []Object
}

// Type returns LIST_OBJ.
func (l *List) Type() ObjectType { return LIST_OBJ }

// Inspect returns the list as a string.
func (l *List) Inspect() string {
	var out strings.Builder
	out.WriteString("[")
	for i, elem := range l.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(elem.Inspect())
	}
	out.WriteString("]")
	return out.String()
}

// Function represents a user-defined function.
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

// Type returns FUNCTION_OBJ.
func (f *Function) Type() ObjectType { return FUNCTION_OBJ }

// Inspect returns a string representation of the function.
func (f *Function) Inspect() string {
	var out strings.Builder
	params := make([]string, len(f.Parameters))
	for i, p := range f.Parameters {
		params[i] = p.Value
	}
	out.WriteString("fn(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {...}")
	return out.String()
}

// ReturnValue wraps a value being returned from a function.
type ReturnValue struct {
	Value Object
}

// Type returns RETURN_VALUE_OBJ.
func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

// Inspect returns the wrapped value's representation.
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }

// HashPair represents a key-value pair in a hash.
type HashPair struct {
	Key   string
	Value Object
}

// Hash represents an object/map with string keys.
type Hash struct {
	Pairs map[string]Object
}

// Type returns HASH_OBJ.
func (h *Hash) Type() ObjectType { return HASH_OBJ }

// Inspect returns the hash as a string.
func (h *Hash) Inspect() string {
	var out strings.Builder
	out.WriteString("{")
	i := 0
	for k, v := range h.Pairs {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(k)
		out.WriteString(": ")
		out.WriteString(v.Inspect())
		i++
	}
	out.WriteString("}")
	return out.String()
}

// Get retrieves a value from the hash by key.
func (h *Hash) Get(key string) (Object, bool) {
	val, ok := h.Pairs[key]
	return val, ok
}

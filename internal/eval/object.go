// Package object defines the runtime value types for the AWSL interpreter.
// These objects are produced by the evaluator when executing AWSL programs.
package eval

import "fmt"

// ObjectType represents the type of a runtime object as a string.
type ObjectType string

// Object types.
const (
	INTEGER_OBJ = "INTEGER"
	FLOAT_OBJ   = "FLOAT"
	STRING_OBJ  = "STRING"
	BOOLEAN_OBJ = "BOOLEAN"
	NULL_OBJ    = "NULL"
	ERROR_OBJ   = "ERROR"
	BUILTIN_OBJ = "BUILTIN"
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
type BuiltinFunction func(args ...Object) Object

// Builtin wraps a Go function as an AWSL callable object.
type Builtin struct {
	Name string
	Fn   BuiltinFunction
}

// Type returns BUILTIN_OBJ.
func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }

// Inspect returns the builtin function name.
func (b *Builtin) Inspect() string { return "builtin:" + b.Name }

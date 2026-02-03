// Package eval implements the tree-walking interpreter for AWSL.
package eval

import (
	"io"
)

// Environment stores variable bindings for the current scope.
// It supports nested scopes through an optional outer environment,
// enabling lexical scoping for functions.
type Environment struct {
	store  map[string]Object
	outer  *Environment
	stdout io.Writer
}

// NewEnvironment creates a new empty environment.
// Use this to create the global/top-level environment.
func NewEnvironment(stdout io.Writer) *Environment {
	return &Environment{
		store:  make(map[string]Object),
		outer:  nil,
		stdout: stdout,
	}
}

// NewEnclosedEnvironment creates a new environment with an outer scope.
// This is used for function calls where variables from outer scopes
// should be readable but assignments create local bindings.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	return &Environment{
		store:  make(map[string]Object),
		outer:  outer,
		stdout: outer.stdout,
	}
}

// Get retrieves a value from the environment by name.
// It searches the current scope first, then walks up the scope chain
// until the variable is found or all scopes are exhausted.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return obj, ok
}

// Set stores a value in the outer scope first then
// falls back to current scope.
func (e *Environment) Set(name string, val Object) Object {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return val
	}

	if e.outer != nil && e.outer.Has(name) {
		return e.outer.Set(name, val)
	}

	e.store[name] = val
	return val
}

// SetLocal always creates or updates a binding in the current scope only,
// shadowing any variable with the same name in outer scopes.
func (e *Environment) SetLocal(name string, val Object) Object {
	e.store[name] = val
	return val
}

// Has checks if a variable exists in this scope or any outer scope recursively.
func (e *Environment) Has(name string) bool {
	if _, ok := e.store[name]; ok {
		return true
	}
	if e.outer != nil {
		return e.outer.Has(name)
	}
	return false
}

// Stdout returns the stdout writer.
func (e *Environment) Stdout() io.Writer {
	if e.stdout != nil {
		return e.stdout
	}
	return nil
}

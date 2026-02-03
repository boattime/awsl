// Package eval implements the tree-walking interpreter for AWSL.
package eval

import (
	"fmt"
	"time"
)

// Builtins contains all built-in functions available in AWSL.
var Builtins = map[string]*Builtin{
	"print": {
		Name: "print",
		Fn:   builtinPrint,
	},
	"clock": {
		Name: "clock",
		Fn:   builtinClock,
	},
}

// RegisterBuiltins adds all built-in functions to the environment.
func RegisterBuiltins(env *Environment) {
	for name, builtin := range Builtins {
		env.Set(name, builtin)
	}
}

// builtinPrint prints values to stdout separated by spaces.
// Returns NULL.
func builtinPrint(env *Environment, args ...Object) Object {
	values := make([]any, len(args))
	for i, arg := range args {
		values[i] = arg.Inspect()
	}
	fmt.Fprintln(env.Stdout(), values...)
	return NULL
}

// builtinClock get the current time in unix seconds.
// Returns Integer in unix seconds.
func builtinClock(env *Environment, args ...Object) Object {
	return &Integer{Value: time.Now().Unix()}
}

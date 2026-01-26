// Package eval implements the tree-walking interpreter for AWSL.
package eval

import "fmt"

// Builtins contains all built-in functions available in AWSL.
var Builtins = map[string]*Builtin{
	"print": {
		Name: "print",
		Fn:   builtinPrint,
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
func builtinPrint(args ...Object) Object {
	values := make([]any, len(args))
	for i, arg := range args {
		values[i] = arg.Inspect()
	}
	// TODO: pass writer through environment for better testing
	// main.go passes os.Stdout, tests pass bytes.Buffer
	fmt.Println(values...)
	return NULL
}

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/boattime/awsl/internal/eval"
	"github.com/boattime/awsl/internal/lexer"
	"github.com/boattime/awsl/internal/parser"
)

// Version information (set via ldflags during build).
var (
	Version   = "dev"
	GitCommit = "unknown"
)

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

// run executes the AWSL interpreter with the given arguments and writers.
// It returns an exit code (0 for success, non-zero for errors).
// This function is separated from main() to enable testing.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: awsl <script.awsl>")
		return 1
	}

	// Handle version flag
	if args[1] == "--version" || args[1] == "-v" {
		fmt.Fprintf(stdout, "awsl version %s (commit: %s)\n", Version, GitCommit)
		return 0
	}

	filename := args[1]
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(stderr, "error reading file: %v\n", err)
		return 1
	}

	// Lex and parse the source file
	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parse errors
	if p.HasErrors() {
		for _, parseErr := range p.Errors() {
			fmt.Fprintln(stderr, parseErr)
		}
		return 1
	}

	env := eval.NewEnvironment(stdout)
	eval.RegisterBuiltins(env)
	result := eval.Eval(program, env)

	if result.Type() == eval.ERROR_OBJ {
		fmt.Fprintln(stderr, result.Inspect())
		return 1
	}

	if result != eval.NULL {
		fmt.Fprintln(stdout, result.Inspect())
	}

	return 0
}

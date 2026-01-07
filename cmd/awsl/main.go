// Package main provides the entry point for the AWSL interpreter.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/boattime/awsl/internal/lexer"
	"github.com/boattime/awsl/internal/token"
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

	// Lex the source file
	if err := lexSource(string(source), stdout); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	return 0
}

// lexSource tokenizes the input source and writes the tokens to the writer.
// Each token is output on its own line with position information.
func lexSource(source string, w io.Writer) error {
	l := lexer.New(source)

	for {
		tok := l.NextToken()

		// Format: LINE:COLUMN\tTYPE\tLITERAL
		fmt.Fprintf(w, "%d:%d\t%s\t%s\n", tok.Line, tok.Column, tok.Type, formatLiteral(tok))

		if tok.Type == token.EOF {
			break
		}

		if tok.Type == token.ILLEGAL {
			return fmt.Errorf("illegal token %q at line %d, column %d", tok.Literal, tok.Line, tok.Column)
		}
	}

	return nil
}

// formatLiteral returns a display-friendly version of the token literal.
// String literals are shown with quotes, empty literals show as <empty>.
func formatLiteral(tok token.Token) string {
	if tok.Type == token.STRING {
		return fmt.Sprintf("%q", tok.Literal)
	}
	if tok.Type == token.EOF {
		return "<eof>"
	}
	if tok.Literal == "" {
		return "<empty>"
	}
	return tok.Literal
}

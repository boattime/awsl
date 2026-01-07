package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: awsl <script.awsl>")
		os.Exit(1)
	}

	filename := os.Args[1]
	_, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file %v\n", err)
		os.Exit(1)
	}
}

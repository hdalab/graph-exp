package main

import (
	"encoding/json"
	"fmt"
	"os"

	ga "github.com/hdalab/ga"
)

// Example of parsing graph specs from either Gexp or JSON formats.
// The program detects the format automatically using json.Valid.
//
// Usage:
//
//	go run ./examples examples/x.gexp
//	go run ./examples examples/x.json
func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run ./examples <spec-file>")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
		os.Exit(1)
	}

	var spec *ga.Spec
	if json.Valid(data) {
		spec, err = ga.ParseJson(data)
	} else {
		spec, err = ga.ParseGexp(data)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal error:", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}

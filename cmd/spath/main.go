package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hdalab/ga"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "matrix":
		fs := flag.NewFlagSet("matrix", flag.ExitOnError)
		in := fs.String("in", "", "input .gexp file")
		_ = fs.Parse(os.Args[2:])
		must(*in != "", "-in required")
		data, err := os.ReadFile(*in)
		mustErr(err)
		spec, err := ga.ParseGexp(data)
		mustErr(err)
		ms := ga.StructuralMatrix(&spec.G)
		printMatrix(os.Stdout, ms)
	case "mdnf":
		fs := flag.NewFlagSet("mdnf", flag.ExitOnError)
		in := fs.String("in", "", "input .gexp file")
		_ = fs.Parse(os.Args[2:])
		must(*in != "", "-in required")
		data, err := os.ReadFile(*in)
		mustErr(err)
		spec, err := ga.ParseGexp(data)
		mustErr(err)
		var paths []ga.Path
		start := time.Now()
		_, _ = ga.EnumerateMDNF(context.Background(), &spec.G, spec.S, spec.T, ga.EnumOptions{}, func(p ga.Path) bool {
			paths = append(paths, p)
			return true
		})
		for _, p := range paths {
			for i, id := range p.EdgeIDs {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Print(id)
			}
			fmt.Println()
		}
		fmt.Fprintf(os.Stderr, "#paths=%d, elapsed=%s\n", len(paths), time.Since(start))
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n  spath matrix -in examples/x.gexp\n  spath mdnf -in examples/x.gexp\n")
	os.Exit(2)
}
func must(cond bool, msg string) {
	if !cond {
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}
}
func mustErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func printMatrix(w io.Writer, ms ga.Structural) {
	n := len(ms)

	// ширина номера строки (0..n-1)
	idxw := len(strconv.Itoa(n - 1))

	// ширина ячейки = макс(длина меток/0)
	cellw := 1
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			s := ms[i][j]
			if s == "" {
				s = "0"
			}
			if l := len(s); l > cellw {
				cellw = l
			}
		}
	}

	// шапка
	fmt.Fprint(w, strings.Repeat(" ", idxw+3)) // место под "i [ "
	for j := 0; j < n; j++ {
		if j > 0 {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprintf(w, "%*d", cellw, j)
	}
	fmt.Fprintln(w)

	// строки
	for i := 0; i < n; i++ {
		fmt.Fprintf(w, "%*d [ ", idxw, i)
		for j := 0; j < n; j++ {
			if j > 0 {
				fmt.Fprint(w, "  ")
			}
			s := ms[i][j]
			if s == "" {
				s = "0"
			}
			fmt.Fprintf(w, "%*s", cellw, s)
		}
		fmt.Fprintln(w, " ]")
	}
}

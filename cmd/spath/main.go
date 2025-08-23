package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hdalab/ga"
)

func humanDur(ns int64) string {
	if ns < 1_000 { // < 1µs
		return fmt.Sprintf("%dns", ns)
	}
	if ns < 1_000_000 { // < 1ms
		return fmt.Sprintf("%dµs", ns/1_000)
	}
	if ns < 1_000_000_000 { // < 1s
		return fmt.Sprintf("%dms", ns/1_000_000)
	}
	// секунды с десятыми
	s := float64(ns) / 1_000_000_000.0
	return fmt.Sprintf("%.1fs", s)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "matrix":
		fs := flag.NewFlagSet("matrix", flag.ExitOnError)
		in := fs.String("in", "", "input .gexp or .json file")
		_ = fs.Parse(os.Args[2:])
		must(*in != "", "-in required")
		spec, err := readSpec(*in)
		mustErr(err)
		ms := ga.StructuralMatrix(&spec.G)
		printMatrix(os.Stdout, ms)
	case "mdnf":
		fs := flag.NewFlagSet("mdnf", flag.ExitOnError)
		in := fs.String("in", "", "input .gexp or .json file")
		statsFlag := fs.Bool("stats", false, "print stats to stderr")
		statsJSON := fs.String("stats-json", "", "write stats to JSON file")
		_ = fs.Parse(os.Args[2:])
		must(*in != "", "-in required")
		spec, err := readSpec(*in)
		mustErr(err)
		var paths []ga.Path
		start := time.Now()
		ctx := context.Background()
		stats, err := ga.EnumerateMDNF(ctx, &spec.G, spec.S, spec.T, ga.EnumOptions{}, func(p ga.Path) bool {
			paths = append(paths, p)
			return true
		})
		mustErr(err)
		// дозаполняем время при необходимости
		measuredNS := time.Since(start).Nanoseconds()
		if stats.ElapsedNS == 0 {
			stats.ElapsedNS = measuredNS
			if stats.NumPaths > 0 {
				stats.NsPerPath = float64(stats.ElapsedNS) / float64(stats.NumPaths)
			}
		}

		fmt.Fprintln(os.Stdout, ga.MDNF(paths))

		if *statsFlag || *statsJSON != "" {
			file := *in
			n := spec.G.N
			m := len(spec.G.Edges)
			if *statsFlag {
				perPath := ""
				if stats.NumPaths > 0 && stats.NsPerPath > 0 {
					perPath = fmt.Sprintf(" (%.1fµs/path)", stats.NsPerPath/1_000.0)
				}
				fmt.Fprintf(os.Stderr,
					"stats: file=%s n=%d m=%d s=%d t=%d paths=%d expanded=%d pruned=%d elapsed=%s%s\n",
					file, n, m, spec.S, spec.T, stats.NumPaths, stats.NodesExpanded, stats.Pruned, humanDur(stats.ElapsedNS), perPath)
			}
			if *statsJSON != "" {
				b, err := json.MarshalIndent(stats, "", "  ")
				if err != nil {
					fmt.Fprintf(os.Stderr, "write --stats-json: %v\n", err)
					os.Exit(1)
				}
				if err := os.WriteFile(*statsJSON, b, 0644); err != nil {
					fmt.Fprintf(os.Stderr, "write --stats-json: %v\n", err)
					os.Exit(1)
				}
			}
		}
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n  spath matrix -in examples/x.gexp\n  spath matrix -in examples/x.json\n  spath mdnf -in examples/x.gexp\n  spath mdnf -in examples/x.json\n")
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

func readSpec(path string) (*ga.Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if json.Valid(data) {
		return ga.ParseJson(data)
	}
	return ga.ParseGexp(data)
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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hdalab/ga"
	util "github.com/hdalab/graph-exp/internal/util"
)

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
		spec, err := util.ReadSpec(*in)
		mustErr(err)
		ms := ga.StructuralMatrix(&spec.G)
		util.PrintMatrix(os.Stdout, ms)
	case "mdnf":
		fs := flag.NewFlagSet("mdnf", flag.ExitOnError)
		in := fs.String("in", "", "input .gexp or .json file")
		statsFlag := fs.Bool("stats", false, "print stats to stderr")
		statsJSON := fs.String("stats-json", "", "write stats to JSON file")
		_ = fs.Parse(os.Args[2:])
		must(*in != "", "-in required")
		spec, err := util.ReadSpec(*in)
		mustErr(err)
		var paths []ga.Path
		start := time.Now()
		ctx := context.Background()
		stats, err := ga.EnumerateMDNF(ctx, &spec.G, spec.S, spec.T, ga.EnumOptions{}, func(p ga.Path) bool {
			paths = append(paths, p)
			return true
		})
		mustErr(err)
		// fill stats if elapsed not set
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
					file, n, m, spec.S, spec.T, stats.NumPaths, stats.NodesExpanded, stats.Pruned, util.HumanDur(stats.ElapsedNS), perPath)
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

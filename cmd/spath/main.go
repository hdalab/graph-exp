// Пакет main предоставляет CLI утилиту для работы с графовыми спецификациями.
package main

import (
	"context"
	"encoding/json"
	"errors"
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
		spec, _ := parseSpec("matrix", os.Args[2:], nil)
		ms := ga.StructuralMatrix(&spec.G)
		util.PrintMatrix(os.Stdout, ms)
	case "mdnf":
		var statsFlag *bool
		var statsJSON *string
		var countOnly *bool
		var maxPaths *int
		var timeout *time.Duration
		var outFile *string
		var quiet bool
		spec, in := parseSpec("mdnf", os.Args[2:], func(fs *flag.FlagSet) {
			statsFlag = fs.Bool("stats", false, "print stats to stderr")
			statsJSON = fs.String("stats-json", "", "write stats to JSON file")
			countOnly = fs.Bool("count", false, "only print number of terms/paths")
			maxPaths = fs.Int("max-paths", 0, "stop after enumerating N paths (0 for no limit)")
			timeout = fs.Duration("timeout", 0, "stop after duration (e.g. 2s, 500ms)")
			outFile = fs.String("o", "", "write MDNF result to file")
			fs.BoolVar(&quiet, "q", false, "suppress regular output")
			fs.BoolVar(&quiet, "quiet", false, "suppress regular output")
		})

		start := time.Now()
		ctx := context.Background()
		if *timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, *timeout)
			defer cancel()
		}

		var paths []ga.Path
		var numPaths int
		var limitReached bool
		var timedOut bool
		stats, err := ga.EnumerateMDNF(ctx, &spec.G, spec.S, spec.T, ga.EnumOptions{}, func(p ga.Path) bool {
			if ctx.Err() != nil {
				timedOut = true
				return false
			}
			numPaths++
			if !*countOnly {
				paths = append(paths, p)
			}
			if *maxPaths > 0 && numPaths >= *maxPaths {
				limitReached = true
				return false
			}
			return true
		})
		if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			mustErr(err)
		}
		if ctx.Err() != nil {
			timedOut = true
		}

		finish := time.Now()
		measuredNS := finish.Sub(start).Nanoseconds()
		if stats.ElapsedNS == 0 {
			stats.ElapsedNS = measuredNS
			if stats.NumPaths > 0 {
				stats.NsPerPath = float64(stats.ElapsedNS) / float64(stats.NumPaths)
			}
		}

		if !quiet {
			if *countOnly {
				if *outFile != "" {
					if err := os.WriteFile(*outFile, []byte(fmt.Sprintf("%d\n", numPaths)), 0644); err != nil {
						fmt.Fprintf(os.Stderr, "write -o: %v\n", err)
						os.Exit(1)
					}
				}
				fmt.Fprintln(os.Stdout, numPaths)
			} else {
				mdnf := ga.MDNF(paths)
				fmt.Fprintln(os.Stdout, mdnf)
				if *outFile != "" {
					if err := os.WriteFile(*outFile, []byte(mdnf+"\n"), 0644); err != nil {
						fmt.Fprintf(os.Stderr, "write -o: %v\n", err)
						os.Exit(1)
					}
				}
			}
		} else if *outFile != "" {
			if *countOnly {
				if err := os.WriteFile(*outFile, []byte(fmt.Sprintf("%d\n", numPaths)), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "write -o: %v\n", err)
					os.Exit(1)
				}
			} else {
				mdnf := ga.MDNF(paths)
				if err := os.WriteFile(*outFile, []byte(mdnf+"\n"), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "write -o: %v\n", err)
					os.Exit(1)
				}
			}
		}

		if *statsFlag || *statsJSON != "" {
			file := in
			n := spec.G.N
			m := len(spec.G.Edges)
			if *statsFlag {
				perPath := ""
				if stats.NumPaths > 0 && stats.NsPerPath > 0 {
					perPath = fmt.Sprintf(" (%.1fµs/path)", stats.NsPerPath/1_000.0)
				}
				msg := fmt.Sprintf("stats: file=%s n=%d m=%d s=%d t=%d paths=%d expanded=%d pruned=%d elapsed=%s%s",
					file, n, m, spec.S, spec.T, stats.NumPaths, stats.NodesExpanded, stats.Pruned, util.HumanDur(stats.ElapsedNS), perPath)
				if limitReached {
					msg += fmt.Sprintf(" truncated at %d", *maxPaths)
				}
				if timedOut {
					msg += " timed out"
				}
				fmt.Fprintln(os.Stderr, msg)
			}
			if *statsJSON != "" {
				type statsOut struct {
					ga.EnumStats
					StartedAt    string `json:"startedAt"`
					FinishedAt   string `json:"finishedAt"`
					TimedOut     bool   `json:"timedOut"`
					LimitReached bool   `json:"limitReached"`
				}
				sj := statsOut{
					EnumStats:    stats,
					StartedAt:    start.Format(time.RFC3339),
					FinishedAt:   finish.Format(time.RFC3339),
					TimedOut:     timedOut,
					LimitReached: limitReached,
				}
				b, err := json.MarshalIndent(sj, "", "  ")
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

// usage выводит краткую информацию о доступных командах.
func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n  spath matrix -in examples/x.gexp\n  spath matrix -in examples/x.json\n  spath mdnf -in examples/x.gexp\n  spath mdnf -in examples/x.json\n")
	os.Exit(2)
}

// must завершает программу, если условие ложно, выводя сообщение.
func must(cond bool, msg string) {
	if !cond {
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}
}

// mustErr завершает выполнение при возникновении ошибки.
func mustErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// parseSpec разбирает общие флаги и читает спецификацию графа.
func parseSpec(name string, args []string, register func(fs *flag.FlagSet)) (*ga.Spec, string) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	in := fs.String("in", "", "input .gexp or .json file")
	if register != nil {
		register(fs)
	}
	_ = fs.Parse(args)
	must(*in != "", "-in required")
	spec, err := util.ReadSpec(*in)
	mustErr(err)
	return spec, *in
}

package main

import (
    "context"
    "flag"
    "fmt"
    "io"
    "os"
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
        data, err := os.ReadFile(*in); mustErr(err)
        spec, err := ga.ParseGexp(data); mustErr(err)
        ms := ga.StructuralMatrix(&spec.G)
        printMatrix(os.Stdout, ms)
    case "mdnf":
        fs := flag.NewFlagSet("mdnf", flag.ExitOnError)
        in := fs.String("in", "", "input .gexp file")
        _ = fs.Parse(os.Args[2:])
        must(*in != "", "-in required")
        data, err := os.ReadFile(*in); mustErr(err)
        spec, err := ga.ParseGexp(data); mustErr(err)
        var paths []ga.Path
        start := time.Now()
        _, _ = ga.EnumerateMDNF(context.Background(), &spec.G, spec.S, spec.T, ga.EnumOptions{}, func(p ga.Path) bool {
            paths = append(paths, p); return true
        })
        for _, p := range paths {
            for i, id := range p.EdgeIDs {
                if i > 0 { fmt.Print(" ") }
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
    if !cond { fmt.Fprintln(os.Stderr, msg); os.Exit(1) }
}
func mustErr(err error) {
    if err != nil { fmt.Fprintln(os.Stderr, "error:", err); os.Exit(1) }
}

func printMatrix(w io.Writer, ms ga.Structural) {
    // header
    fmt.Fprint(w, "      ")
    for j := 0; j < len(ms); j++ {
        if j > 0 { fmt.Fprint(w, "  ") }
        fmt.Fprintf(w, "%d", j)
    }
    fmt.Fprintln(w)
    for i := 0; i < len(ms); i++ {
        fmt.Fprintf(w, "%d [ ", i)
        for j := 0; j < len(ms); j++ {
            cell := ms[i][j]
            if cell == "" { cell = "0" }
            if j > 0 { fmt.Fprint(w, "  ") }
            fmt.Fprint(w, cell)
        }
        fmt.Fprintln(w, " ]")
    }
}

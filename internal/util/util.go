package util

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/hdalab/ga"
)

// HumanDur converts nanoseconds to a short human-readable string.
func HumanDur(ns int64) string {
	if ns < 1_000 { // < 1µs
		return fmt.Sprintf("%dns", ns)
	}
	if ns < 1_000_000 { // < 1ms
		return fmt.Sprintf("%dµs", ns/1_000)
	}
	if ns < 1_000_000_000 { // < 1s
		return fmt.Sprintf("%dms", ns/1_000_000)
	}
	// seconds with one decimal place
	s := float64(ns) / 1_000_000_000.0
	return fmt.Sprintf("%.1fs", s)
}

// ReadSpec loads a graph specification from either Gexp or JSON format.
func ReadSpec(path string) (*ga.Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if json.Valid(data) {
		return ga.ParseJson(data)
	}
	return ga.ParseGexp(data)
}

// PrintMatrix renders a structural matrix to the given writer.
func PrintMatrix(w io.Writer, ms ga.Structural) {
	n := len(ms)

	// width of row number (0..n-1)
	idxw := len(strconv.Itoa(n - 1))

	// width of cell = max(len(label),1)
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

	// header
	fmt.Fprint(w, strings.Repeat(" ", idxw+3)) // space for "i [ "
	for j := 0; j < n; j++ {
		if j > 0 {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprintf(w, "%*d", cellw, j)
	}
	fmt.Fprintln(w)

	// rows
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

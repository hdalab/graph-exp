package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/hdalab/ga"
	util "github.com/hdalab/graph-exp/internal/util"
)

// runCmd запускает утилиту spath с заданными аргументами.
func runCmd(t *testing.T, args ...string) (string, string) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
	var out bytes.Buffer
	var err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	if e := cmd.Run(); e != nil {
		t.Fatalf("runCmd: %v stderr=%s", e, err.String())
	}
	return out.String(), err.String()
}

// expectedMatrix возвращает ожидаемую структурную матрицу.
func expectedMatrix(t *testing.T) string {
	spec, err := util.ReadSpec("../../examples/x.gexp")
	if err != nil {
		t.Fatalf("ReadSpec: %v", err)
	}
	ms := ga.StructuralMatrix(&spec.G)
	var buf bytes.Buffer
	util.PrintMatrix(&buf, ms)
	return buf.String()
}

// expectedMDNF вычисляет эталонное выражение MDNF.
func expectedMDNF(t *testing.T) string {
	spec, err := util.ReadSpec("../../examples/x.json")
	if err != nil {
		t.Fatalf("ReadSpec: %v", err)
	}
	var paths []ga.Path
	_, err = ga.EnumerateMDNF(context.Background(), &spec.G, spec.S, spec.T, ga.EnumOptions{}, func(p ga.Path) bool {
		paths = append(paths, p)
		return true
	})
	if err != nil {
		t.Fatalf("EnumerateMDNF: %v", err)
	}
	return ga.MDNF(paths) + "\n"
}

// TestMatrix проверяет вывод подкоманды matrix.
func TestMatrix(t *testing.T) {
	out, err := runCmd(t, "matrix", "-in", "../../examples/x.json")
	if out != expectedMatrix(t) {
		t.Fatalf("matrix output:\n%s\nwant:\n%s", out, expectedMatrix(t))
	}
	if err != "" {
		t.Fatalf("unexpected stderr: %s", err)
	}
}

// TestMDNF проверяет вывод подкоманды mdnf.
func TestMDNF(t *testing.T) {
	out, err := runCmd(t, "mdnf", "-in", "../../examples/x.json")
	if out != expectedMDNF(t) {
		t.Fatalf("mdnf output:\n%s\nwant:\n%s", out, expectedMDNF(t))
	}
	if err != "" {
		t.Fatalf("unexpected stderr: %s", err)
	}
}

// TestMDNFStats проверяет вывод флагов статистики.
func TestMDNFStats(t *testing.T) {
	out, err := runCmd(t, "mdnf", "-in", "../../examples/x.json", "--stats")
	if out != expectedMDNF(t) {
		t.Fatalf("mdnf output:\n%s\nwant:\n%s", out, expectedMDNF(t))
	}
	if !strings.HasPrefix(err, "stats:") {
		t.Fatalf("missing stats: %s", err)
	}
}

// TestMDNFCount проверяет вывод флага -count.
func TestMDNFCount(t *testing.T) {
	out, err := runCmd(t, "mdnf", "-in", "../../examples/x.json", "-count")
	expected := strings.Count(expectedMDNF(t), "|") + 1
	if out != fmt.Sprintf("%d\n", expected) {
		t.Fatalf("count output=%s want=%d\n", out, expected)
	}
	if err != "" {
		t.Fatalf("unexpected stderr: %s", err)
	}
}

// TestMDNFMaxPaths проверяет усечение по флагу -max-paths.
func TestMDNFMaxPaths(t *testing.T) {
	out, err := runCmd(t, "mdnf", "-in", "../../examples/x.json", "-max-paths", "1")
	first := strings.TrimSpace(strings.Split(expectedMDNF(t), "|")[0]) + "\n"
	if out != first {
		t.Fatalf("max-paths output=%s want=%s", out, first)
	}
	if err != "" {
		t.Fatalf("unexpected stderr: %s", err)
	}
}

// TestMDNFQuiet проверяет подавление вывода флагом -q.
func TestMDNFQuiet(t *testing.T) {
	out, err := runCmd(t, "mdnf", "-in", "../../examples/x.json", "-q")
	if out != "" {
		t.Fatalf("quiet stdout=%s want empty", out)
	}
	if err != "" {
		t.Fatalf("unexpected stderr: %s", err)
	}
}

// TestMDNFOutFile проверяет сохранение результата в файл флагом -o.
func TestMDNFOutFile(t *testing.T) {
	tmp := t.TempDir()
	file := tmp + "/out.mdnf"
	out, err := runCmd(t, "mdnf", "-in", "../../examples/x.json", "-o", file, "-q")
	if out != "" {
		t.Fatalf("stdout=%s want empty", out)
	}
	data, readErr := os.ReadFile(file)
	if readErr != nil {
		t.Fatalf("read file: %v", readErr)
	}
	if string(data) != expectedMDNF(t) {
		t.Fatalf("file output=%s want=%s", string(data), expectedMDNF(t))
	}
	if err != "" {
		t.Fatalf("unexpected stderr: %s", err)
	}
}

// TestMDNFTimeout проверяет остановку по таймауту.
func TestMDNFTimeout(t *testing.T) {
	out, err := runCmd(t, "mdnf", "-in", "../../examples/x.json", "-count", "-timeout", "1ns")
	if out != "0\n" {
		t.Fatalf("timeout stdout=%s want 0", out)
	}
	if err != "" {
		t.Fatalf("unexpected stderr: %s", err)
	}
}

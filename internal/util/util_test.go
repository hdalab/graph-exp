package util

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/hdalab/ga"
)

func TestHumanDur(t *testing.T) {
	cases := []struct {
		ns   int64
		want string
	}{
		{500, "500ns"},
		{15_000, "15µs"},
		{2_500_000, "2ms"},
		{1_500_000_000, "1.5s"},
	}
	for _, c := range cases {
		if got := HumanDur(c.ns); got != c.want {
			t.Errorf("HumanDur(%d)=%s want %s", c.ns, got, c.want)
		}
	}
}

func TestReadSpec(t *testing.T) {
	gexp, err := ReadSpec("../../examples/x.gexp")
	if err != nil {
		t.Fatalf("ReadSpec gexp: %v", err)
	}
	jsonSpec, err := ReadSpec("../../examples/x.json")
	if err != nil {
		t.Fatalf("ReadSpec json: %v", err)
	}
	if !reflect.DeepEqual(gexp, jsonSpec) {
		t.Fatalf("specs differ: %#v vs %#v", gexp, jsonSpec)
	}
}

func TestPrintMatrix(t *testing.T) {
	spec, err := ReadSpec("../../examples/x.gexp")
	if err != nil {
		t.Fatalf("ReadSpec: %v", err)
	}
	ms := ga.StructuralMatrix(&spec.G)
	var buf bytes.Buffer
	PrintMatrix(&buf, ms)
	want := "    0  1  2  3  4  5\n" +
		"0 [ 0  a  b  0  0  0 ]\n" +
		"1 [ 0  0  c  d  0  0 ]\n" +
		"2 [ 0  0  0  e  f  i ]\n" +
		"3 [ 0  0  0  0  g  0 ]\n" +
		"4 [ 0  0  0  0  0  h ]\n" +
		"5 [ 0  0  0  0  0  0 ]\n"
	if got := buf.String(); got != want {
		t.Fatalf("PrintMatrix:\n%s\nwant:\n%s", got, want)
	}
}

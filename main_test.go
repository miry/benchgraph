package main

import (
	"testing"

	"golang.org/x/tools/benchmark/parse"
)

var bTests = []struct {
	line    string // input
	name    string // expected result
	arg     string
	nsperop float64
}{
	{"BenchmarkF2_F0000000-4		50000000	        29.4 ns/op", "F2", "F0000000", 29.4},
	{"BenchmarkF0_FF-2		10000000	        37.4 ns/op", "F0", "FF", 37.4},
	{"BenchmarkF_0-2		40000000	        11.2 ns/op", "F", "0", 11.2},
	{"BenchmarkF3/quicksort_100-4		40000000	        11.2 ns/op", "F3/quicksort", "100", 11.2},
}

func TestParser(t *testing.T) {
	for _, tt := range bTests {
		b, _ := parse.ParseLine(tt.line)
		functionSig, _ := parseFunctionSignature(defaultFunctionSignaturePattern, b.Name)
		if functionSig.name != tt.name {
			t.Errorf(
				"parseFunctionSignature(%s): expected %s, actual %s",
				b.Name,
				tt.name,
				functionSig.name,
			)
		}
		if functionSig.arg != tt.arg {
			t.Errorf(
				"parseFunctionSignature(%s): expected %s, actual %s",
				b.Name,
				tt.arg,
				functionSig.arg,
			)
		}
		if b.NsPerOp != tt.nsperop {
			t.Errorf("parseFunctionSignature(%s): expected %f, actual %f", b.Name, tt.nsperop, b.NsPerOp)
		}
	}
}

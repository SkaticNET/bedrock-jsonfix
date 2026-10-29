package bedrockjsonfix

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFixLeadingZeroNumbers(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"version", `{"version":[01,026,052]}`, `{"version":[1,26,52]}`},
		{"decimal integers", `[08,009,-0012,000,-000]`, `[8,9,-12,0,-0]`},
		{"fractions", `[000.125,-001.2500,00.000]`, `[0.125,-1.2500,0.000]`},
		{"exponents", `[001e+002,-0002E-003,00e000]`, `[1e+002,-2E-003,0e000]`},
		{"precision", `[0009007199254740993,000.12345678901234567890123456789]`, `[9007199254740993,0.12345678901234567890123456789]`},
		{"quoted bytes", `{"01":"000","text":"\\\"001e02","01":002}`, `{"01":"000","text":"\\\"001e02","01":2}`},
		{"other repairs", "prefix {/* 002 */\"value\":001,\"list\":[002,],} suffix", `{ "value":1,"list":[2]}`},
		{"valid numbers", `[0,-0,0.001,1e+003,-1.02E-004,9007199254740993]`, `[0,-0,0.001,1e+003,-1.02E-004,9007199254740993]`},
		{"valid numbers with comments", `/* repair */[0.001,1e+003,-1.02E-004]`, `[0.001,1e+003,-1.02E-004]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := []byte(tt.input)
			before := bytes.Clone(input)
			got, err := Fix(input)
			if err != nil {
				t.Fatalf("Fix: %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %s; want %s", got, tt.want)
			}
			if !bytes.Equal(input, before) {
				t.Fatal("Fix changed its input")
			}
			again, err := Fix(got)
			if err != nil || !bytes.Equal(again, got) {
				t.Fatalf("second Fix = %s, %v", again, err)
			}
		})
	}
}

func TestFixLeadingZerosRejectsMalformedNumbers(t *testing.T) {
	for _, number := range []string{"+01", "01e", "01e+", "01.", "0x01", "01 02", "01..2", "--01", "01e01e01"} {
		t.Run(number, func(t *testing.T) {
			if got, err := Fix([]byte("[" + number + "]")); err == nil {
				t.Fatalf("accepted %q as %s", number, got)
			}
		})
	}
}

func FuzzFixNumbers(f *testing.F) {
	for _, input := range []string{`[001,-00.125,01e02]`, `{"01":"\\\"002","n":003,}`, `/*x*/[1e+003]`, `[01..2]`, `{"a":1}/*`, `/*lead*/[1,]/*`, `null {"later":1}`, `/*lead*/true/*`, `"first" {}`, `001 []`, `1e+ {}`} {
		f.Add([]byte(input))
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		before := bytes.Clone(input)
		got, err := Fix(input)
		if !bytes.Equal(input, before) {
			t.Fatal("Fix changed its input")
		}
		if err != nil {
			return
		}
		if !json.Valid(got) {
			t.Fatalf("invalid result: %q", got)
		}
		again, err := Fix(got)
		if err != nil || !bytes.Equal(got, again) {
			t.Fatalf("second Fix = %q, %v", again, err)
		}
	})
}

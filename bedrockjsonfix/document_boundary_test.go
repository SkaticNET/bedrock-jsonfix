package bedrockjsonfix

import (
	"bytes"
	"testing"
)

func TestFixStopsAfterFirstDocument(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"object", `{"cat":{"event_name":"old"}}/*`, `{"cat":{"event_name":"old"}}`},
		{"array", `[1,{"nested":[2]}] /*`, `[1,{"nested":[2]}]`},
		{"existing repairs", `/* lead */{"a":001,"a":002,"items":[1,],} /*`, `{"a":1,"a":2,"items":[1]}`},
		{"comment braces", `{"a":/* } ] */[1]}/*`, `{"a": [1]}`},
		{"quoted delimiters", `{"a":"} ] /* \\\"","b":[{}]} /*`, `{"a":"} ] /* \\\"","b":[{}]}`},
		{"quoted first value", `"skip { }" {"a":1}/*`, `"skip { }"`},
		{"prefix closers", `} ] prefix {"a":1}/*`, `{"a":1}`},
		{"second document", `{"a":1}/* {"a":2}`, `{"a":1}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := []byte(tc.input)
			before := append([]byte(nil), input...)
			got, err := Fix(input)
			if err != nil || string(got) != tc.want {
				t.Fatalf("Fix = %q, %v; want %q", got, err, tc.want)
			}
			if !bytes.Equal(input, before) {
				t.Fatal("Fix changed its input")
			}
			if again, err := Fix(got); err != nil || !bytes.Equal(again, got) {
				t.Fatalf("second Fix = %q, %v; want %q", again, err, got)
			}
		})
	}
}

func TestFixRejectsIncompleteSelectedDocument(t *testing.T) {
	for _, input := range []string{`/* {"a":1}`, `{"a":/* }`, `{"a":[1}/*`, `{"a":1 /* }`, `{"a":} {"a":1}/*`} {
		if got, err := Fix([]byte(input)); err == nil {
			t.Errorf("Fix(%q) = %q without error", input, got)
		}
	}
}

func TestFixFirstScalarBoundary(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{`1abc {}`, `1`}, {`1.234567890123456789e+12345 /*`, `1.234567890123456789e+12345`},
		{`null {"cat":1}`, `null`}, {`nullx {}`, `null`}, {`truegarbage /*`, `true`}, {`true [1]`, `true`}, {`false {}`, `false`},
		{`"text" {}`, `"text"`}, {`-1.25e+2 []`, `-1.25e+2`}, {`001 {}`, `1`},
		{`/*lead*/ null /*`, `null`}, {`true /*`, `true`}, {`"a\"b" /*`, `"a\"b"`},
		{"\"a\x01b\" /*", `"a\u0001b"`}, {`-001.5 /*`, `-1.5`},
	} {
		t.Run(tc.input, func(t *testing.T) {
			in := []byte(tc.input)
			before := bytes.Clone(in)
			got, err := Fix(in)
			if err != nil || string(got) != tc.want {
				t.Fatalf("Fix = %q, %v; want %q", got, err, tc.want)
			}
			if !bytes.Equal(in, before) {
				t.Fatal("input mutated")
			}
		})
	}
}

func TestFixRejectsMalformedFirstScalar(t *testing.T) {
	for _, in := range []string{`text "garbage" {"a":1}/*`, `nul {}`, `nux {}`, `tru {}`, `1e+ {}`, `- {} `, `"unfinished {}`} {
		if got, err := Fix([]byte(in)); err == nil {
			t.Errorf("Fix(%q) = %q without error", in, got)
		}
	}
}

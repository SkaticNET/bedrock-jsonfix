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
		{"quoted prefix", `"skip { }" {"a":1}/*`, `{"a":1}`},
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

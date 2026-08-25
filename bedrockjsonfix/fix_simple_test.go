package bedrockjsonfix

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFixMakesBedrockAuthoringSyntaxValidJSON(t *testing.T) {
	input := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`pack header {
  // retained data must stay intact
  "name": "Example",
  "url": "https://example.invalid/a//b",
  "value": "low",
  "value": "high",
}`)...)

	fixed, err := Fix(input)
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	if !json.Valid(fixed) {
		t.Fatalf("Fix returned invalid JSON: %q", fixed)
	}
	var value map[string]string
	if err := json.Unmarshal(fixed, &value); err != nil {
		t.Fatalf("decode fixed JSON: %v", err)
	}
	if value["name"] != "Example" || value["url"] != "https://example.invalid/a//b" || value["value"] != "high" {
		t.Fatalf("fixed data mismatch: %#v", value)
	}
}

func TestFixDropsOnlyDataOutsideTheFirstDocument(t *testing.T) {
	fixed, err := Fix([]byte(`prefix {"name":"Example","items":[1,2,],} suffix`))
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	want := []byte(`{"name":"Example","items":[1,2]}`)
	if !bytes.Equal(fixed, want) {
		t.Fatalf("Fix output = %q, want %q", fixed, want)
	}
}

func TestFixSkipsBracesInsideLeadingGarbageStrings(t *testing.T) {
	fixed, err := Fix([]byte(`prefix "not a document: { }" {"name":"Example"}`))
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	want := []byte(`{"name":"Example"}`)
	if !bytes.Equal(fixed, want) {
		t.Fatalf("Fix output = %q, want %q", fixed, want)
	}
}

func TestFixRejectsIncompleteJSON(t *testing.T) {
	if _, err := Fix([]byte(`{"name":`)); err == nil {
		t.Fatal("Fix accepted incomplete JSON")
	}
}

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunRepairsFileAndWritesRequestedReport(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	inputPath := filepath.Join(directory, "hello.json")
	outputPath := filepath.Join(directory, "hello.cleaned.json")
	reportPath := filepath.Join(directory, "hello.report.json")
	input := []byte("discarded prefix /* Resource-pack authoring comment */\n{\n  \"name\": \"hello\",\n  \"items\": [\"first\",],\n}\nignored trailer")
	if err := os.WriteFile(inputPath, input, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{"--report", reportPath, inputPath, outputPath}); err != nil {
		t.Fatal(err)
	}

	fixed, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(fixed), "{\n  \"name\": \"hello\",\n  \"items\": [\"first\"]\n}"; got != want {
		t.Errorf("fixed output = %q, want %q", got, want)
	}

	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	var report map[string]any
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		t.Fatalf("report is not JSON: %v", err)
	}
	if got, want := report["input_path"], inputPath; got != want {
		t.Errorf("input_path = %#v, want %#v", got, want)
	}
	if got, want := report["output_path"], outputPath; got != want {
		t.Errorf("output_path = %#v, want %#v", got, want)
	}
	if got, want := report["input_bytes"], float64(len(input)); got != want {
		t.Errorf("input_bytes = %#v, want %#v", got, want)
	}
	if got, want := report["output_bytes"], float64(len(fixed)); got != want {
		t.Errorf("output_bytes = %#v, want %#v", got, want)
	}
	if got := report["changed"]; got != true {
		t.Errorf("changed = %#v, want true", got)
	}
	if got := report["valid_json"]; got != true {
		t.Errorf("valid_json = %#v, want true", got)
	}
}

func TestRunRepairsFileWithoutReport(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	inputPath := filepath.Join(directory, "hello.json")
	outputPath := filepath.Join(directory, "hello.cleaned.json")
	if err := os.WriteFile(inputPath, []byte(`{"name":"hello",}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{inputPath, outputPath}); err != nil {
		t.Fatal(err)
	}

	fixed, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(fixed), `{"name":"hello"}`; got != want {
		t.Errorf("fixed output = %q, want %q", got, want)
	}
}

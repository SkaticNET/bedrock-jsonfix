package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/SkaticNET/bedrock-jsonfix/v2/bedrockjsonfix"
)

type report struct {
	InputPath   string `json:"input_path"`
	OutputPath  string `json:"output_path"`
	InputBytes  int    `json:"input_bytes"`
	OutputBytes int    `json:"output_bytes"`
	Changed     bool   `json:"changed"`
	ValidJSON   bool   `json:"valid_json"`
}

type usageError struct{}

func (usageError) Error() string {
	return "usage: fixfile [--report <report.json>] <input.json> <output.json>"
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if _, ok := err.(usageError); ok {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	reportPath := ""
	if len(args) > 0 && args[0] == "--report" {
		if len(args) < 2 || args[1] == "" {
			return usageError{}
		}
		reportPath = args[1]
		args = args[2:]
	}
	if len(args) != 2 {
		return usageError{}
	}

	inputPath := args[0]
	outputPath := args[1]
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read %q: %w", inputPath, err)
	}
	fixed, err := bedrockjsonfix.Fix(input)
	if err != nil {
		return fmt.Errorf("fix %q: %w", inputPath, err)
	}
	if err := os.WriteFile(outputPath, fixed, 0o644); err != nil {
		return fmt.Errorf("write %q: %w", outputPath, err)
	}
	if reportPath == "" {
		return nil
	}

	data, err := json.MarshalIndent(report{
		InputPath:   inputPath,
		OutputPath:  outputPath,
		InputBytes:  len(input),
		OutputBytes: len(fixed),
		Changed:     !bytes.Equal(input, fixed),
		ValidJSON:   json.Valid(fixed),
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if err := os.WriteFile(reportPath, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write report %q: %w", reportPath, err)
	}
	return nil
}

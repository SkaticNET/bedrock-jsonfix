package main

import (
	"fmt"
	"os"

	"github.com/SkaticNET/bedrock-jsonfix/bedrockjsonfix"
)

func run() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: fixfile <input.json> <output.json>")
	}
	in, err := os.ReadFile(os.Args[1])
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	opt := bedrockjsonfix.DefaultOptions()
	res, err := bedrockjsonfix.FixBytes(in, opt)
	if err != nil {
		return fmt.Errorf("fix input: %w", err)
	}
	if err := os.WriteFile(os.Args[2], res.Output, 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	fmt.Printf("fixed root=%v bytes=%d\n", res.Root, len(res.Output))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

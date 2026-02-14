package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/SkaticNET/bedrock-jsonfix/bedrockjsonfix"
)

func run() error {
	pretty := flag.Bool("pretty", true, "pretty print output")
	flag.Parse()
	if flag.NArg() < 1 {
		return fmt.Errorf("usage: jsonfix <file>")
	}
	in, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	opt := bedrockjsonfix.DefaultOptions()
	opt.Pretty = *pretty
	res, err := bedrockjsonfix.FixBytes(in, opt)
	if err != nil {
		return fmt.Errorf("fix input: %w", err)
	}
	if _, err := os.Stdout.Write(res.Output); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

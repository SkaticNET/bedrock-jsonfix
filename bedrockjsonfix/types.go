package bedrockjsonfix

import (
	"strings"
)

// Mode controls parser tolerance behavior.
type Mode int

const (
	// ModeStrict accepts strict JSON only.
	ModeStrict Mode = iota
	// ModeBedrock accepts tolerant Bedrock JSON-ish input.
	ModeBedrock
	// ModeBedrockSafe aliases ModeBedrock and is recommended default.
	ModeBedrockSafe = ModeBedrock
)

// RootKind indicates detected top-level JSON root type.
type RootKind int

const (
	// RootObject indicates a JSON object root.
	RootObject RootKind = iota + 1
	// RootArray indicates a JSON array root.
	RootArray
	// RootUnknown indicates no known root was detected.
	RootUnknown
)

// Options configure normalization and safety limits.
type Options struct {
	Mode Mode

	Pretty          bool
	Indent          string
	Prefix          string
	PreserveIfValid bool

	MaxInputBytes  int64
	MaxOutputBytes int64

	AllowCP1252Fallback    bool
	AggressiveWhitespace   bool
	DropJunkOutsideStrings bool
	TrimToFirstRoot        bool
	TrimAfterFirstRoot     bool
	RootScanAttempts       int

	EscapeStringControls bool
}

// Warning represents a non-fatal observation.
type Warning struct {
	Code    string
	Message string
}

// Report describes applied fixes and parsing decisions.
type Report struct {
	InputWasInvalidUTF8 bool
	UsedCP1252Fallback  bool

	RemovedBOM                  int
	ReplacedNBSP                int
	RemovedZeroWidth            int
	RemovedASCIIControls        int
	NormalizedCRLF              int
	EscapedStringControls       int
	NormalizedNewlinesInStrings int

	StrippedLineComments  int
	StrippedBlockComments int
	RemovedTrailingCommas int

	DroppedJunkOutsideStrings int
	TrimmedLeadingJunkBytes   int
	TrimmedTrailingJunkBytes  int
	RootScanUsed              bool
	RootScanAttemptsUsed      int

	ValidJSON bool
}

// Result is the output of a normalization run.
type Result struct {
	Output   []byte
	Root     RootKind
	Report   Report
	Warnings []Warning
}

// DefaultOptions returns safe defaults for public services.
func DefaultOptions() Options {
	return Options{
		Mode:                   ModeBedrock,
		Pretty:                 true,
		Indent:                 "  ",
		Prefix:                 "",
		PreserveIfValid:        true,
		MaxInputBytes:          2 << 20,
		MaxOutputBytes:         7 << 20,
		AllowCP1252Fallback:    true,
		AggressiveWhitespace:   false,
		DropJunkOutsideStrings: true,
		TrimToFirstRoot:        true,
		TrimAfterFirstRoot:     true,
		RootScanAttempts:       5,
		EscapeStringControls:   true,
	}
}

// Validate checks option consistency.
func (o *Options) Validate() error {
	if o.MaxInputBytes <= 0 || o.MaxOutputBytes <= 0 {
		return &FixError{Code: "invalid_options", Message: "size limits must be > 0", Cause: ErrOptionsInvalid}
	}
	if o.RootScanAttempts < 0 {
		return &FixError{Code: "invalid_options", Message: "root scan attempts cannot be negative", Cause: ErrOptionsInvalid}
	}
	if strings.ContainsAny(o.Indent, "\r\n") || strings.ContainsAny(o.Prefix, "\r\n") {
		return &FixError{Code: "invalid_options", Message: "indent/prefix cannot contain newlines", Cause: ErrOptionsInvalid}
	}
	if o.Mode != ModeStrict && o.Mode != ModeBedrock && o.Mode != ModeBedrockSafe {
		return &FixError{Code: "invalid_options", Message: "unknown mode", Cause: ErrOptionsInvalid}
	}
	return nil
}

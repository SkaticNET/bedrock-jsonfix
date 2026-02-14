package bedrockjsonfix

import "errors"

var (
	ErrInputTooLarge   = errors.New("input too large")
	ErrOutputTooLarge  = errors.New("output too large")
	ErrNoRootFound     = errors.New("no json root found")
	ErrInvalidJSON     = errors.New("invalid json")
	ErrOptionsInvalid  = errors.New("invalid options")
	ErrContextCanceled = errors.New("context canceled")
)

// FixError provides stable error coding and wrapped causes.
type FixError struct {
	Code    string
	Message string
	Cause   error
}

func (e *FixError) Error() string { return e.Code + ": " + e.Message }
func (e *FixError) Unwrap() error { return e.Cause }

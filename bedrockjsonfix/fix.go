package bedrockjsonfix

import (
	"context"
	"errors"
	"fmt"
	"io"
)

var errRootRejected = errors.New("root rejected by validator")

func parseCandidate(raw []byte, opt Options) ([]byte, RootKind, error) {
	out, kind, parsed, err := parseAndMarshalWithParsed(raw, opt)
	if err != nil {
		return nil, RootUnknown, err
	}
	if !acceptRootCandidate(opt, kind, raw, parsed) {
		return nil, RootUnknown, errRootRejected
	}
	return out, kind, nil
}

// FixString normalizes string input.
func FixString(input string, opt Options) (Result, error) { return FixBytes([]byte(input), opt) }

// FixReader normalizes from an io.Reader with context and bounded memory.
func FixReader(ctx context.Context, r io.Reader, opt Options) (Result, error) {
	if err := opt.Validate(); err != nil {
		return Result{}, err
	}
	select {
	case <-ctx.Done():
		return Result{}, &FixError{Code: "context_canceled", Message: "operation canceled", Cause: ErrContextCanceled}
	default:
	}
	limited := io.LimitReader(r, opt.MaxInputBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		if ctx.Err() != nil {
			return Result{}, &FixError{Code: "context_canceled", Message: "operation canceled", Cause: ErrContextCanceled}
		}
		return Result{}, err
	}
	if int64(len(data)) > opt.MaxInputBytes {
		return Result{}, &FixError{Code: "input_too_large", Message: fmt.Sprintf("input exceeds limit (%d > %d)", len(data), opt.MaxInputBytes), Cause: ErrInputTooLarge}
	}
	return FixBytes(data, opt)
}

// FixBytes normalizes tolerant JSON-ish bytes into strict JSON.
func FixBytes(input []byte, opt Options) (Result, error) {
	if err := opt.Validate(); err != nil {
		return Result{}, err
	}
	if int64(len(input)) > opt.MaxInputBytes {
		return Result{}, &FixError{Code: "input_too_large", Message: fmt.Sprintf("input exceeds limit (%d > %d)", len(input), opt.MaxInputBytes), Cause: ErrInputTooLarge}
	}

	if opt.PreserveIfValid {
		if ok, root := strictJSONSingleDocument(input); ok {
			if int64(len(input)) > opt.MaxOutputBytes {
				return Result{}, &FixError{Code: "output_too_large", Message: fmt.Sprintf("output exceeds limit (%d > %d)", len(input), opt.MaxOutputBytes), Cause: ErrOutputTooLarge}
			}
			res := Result{Output: append([]byte(nil), input...), Root: root}
			res.Report.ValidJSON = true
			return res, nil
		}
	}

	var rep Report
	decoded, err := decodeInput(input, opt, &rep)
	if err != nil {
		return Result{}, err
	}

	rootKind := RootUnknown
	candidate := decoded
	if opt.Mode == ModeStrict {
		out, kind, parseErr := parseAndMarshal(candidate, opt)
		if parseErr != nil {
			return Result{}, &FixError{Code: "invalid_json", Message: "strict mode requires a valid single JSON document", Cause: errors.Join(ErrInvalidJSON, parseErr)}
		}
		if int64(len(out)) > opt.MaxOutputBytes {
			return Result{}, &FixError{Code: "output_too_large", Message: fmt.Sprintf("output exceeds limit (%d > %d)", len(out), opt.MaxOutputBytes), Cause: ErrOutputTooLarge}
		}
		rep.ValidJSON = true
		return Result{Output: out, Root: kind, Report: rep}, nil
	}

	clean := sanitize(decoded, opt, &rep)
	if opt.EscapeStringControls {
		clean = escapeStringControls(clean, &rep)
	}
	clean = normalizeLiteralNewlinesInStrings(clean, &rep)
	clean = stripComments(clean, &rep)
	clean = removeTrailingCommas(clean, &rep)
	candidate = clean
	if opt.Mode == ModeBedrock || opt.Mode == ModeBedrockSafe {
		if opt.TrimToFirstRoot {
			i := firstRootStartOutsideStrings(candidate, 0)
			if i < 0 {
				return Result{}, &FixError{Code: "no_root", Message: "no JSON root object/array found", Cause: ErrNoRootFound}
			}
			rep.TrimmedLeadingJunkBytes += i
			candidate = candidate[i:]
		}
		if opt.DropJunkOutsideStrings {
			candidate = dropUnknownOutsideStrings(candidate, &rep)
		}
		if opt.TrimAfterFirstRoot {
			oldLen := len(candidate)
			if end, ok, _ := trimAfterFirstValueUsingDecoder(candidate); ok && end <= oldLen {
				clampedEnd := minInt(end, oldLen)
				candidate = candidate[:clampedEnd]
				rep.TrimmedTrailingJunkBytes += oldLen - clampedEnd
			} else {
				start, end, kind, er, extractErr := ExtractFirstJSONValue(candidate, opt)
				mergeReport(&rep, er)
				if extractErr == nil {
					if start > 0 {
						rep.TrimmedLeadingJunkBytes += start
					}
					if end < len(candidate) {
						rep.TrimmedTrailingJunkBytes += len(candidate) - end
					}
					candidate = candidate[start:end]
					rootKind = kind
				}
			}
		}
	}

	var (
		out      []byte
		kind     RootKind
		parseErr error
	)
	if opt.Mode == ModeBedrock || opt.Mode == ModeBedrockSafe {
		out, kind, parseErr = parseCandidate(candidate, opt)
	} else {
		out, kind, parseErr = parseAndMarshal(candidate, opt)
	}
	if parseErr != nil && (opt.Mode == ModeBedrock || opt.Mode == ModeBedrockSafe) && shouldScanAfterFailure(parseErr, opt, candidate) {
		rep.RootScanUsed = true
		maxCandidates := opt.effectiveRootScanMaxCandidates()
		scanFrom := 0
		baseLeading := rep.TrimmedLeadingJunkBytes
		for attempt := 1; attempt <= maxCandidates; attempt++ {
			next := nextRootCandidate(candidate, scanFrom+1)
			if next < 0 {
				break
			}
			rep.RootScanAttemptsUsed = attempt
			scanFrom = next
			rep.TrimmedLeadingJunkBytes = baseLeading + next
			trimmed := candidate[next:]
			if opt.TrimAfterFirstRoot {
				if end, ok, _ := trimAfterFirstValueUsingDecoder(trimmed); ok && end <= len(trimmed) {
					trimmed = trimmed[:minInt(end, len(trimmed))]
				} else {
					s, e, _, er, extractErr := ExtractFirstJSONValue(trimmed, opt)
					mergeReport(&rep, er)
					if extractErr != nil {
						continue
					}
					trimmed = trimmed[s:e]
				}
			}
			out, kind, parseErr = parseCandidate(trimmed, opt)
			if parseErr == nil {
				break
			}
			if opt.RootPolicy == RootPolicyScanLeadingJunk && !errors.Is(parseErr, errRootRejected) && !isLikelyWrongRootStart(parseErr, opt.WrongStartMaxOffset) {
				break
			}
		}
	}
	if parseErr != nil {
		return Result{}, &FixError{Code: "invalid_json", Message: "unable to normalize into a valid single JSON document", Cause: errors.Join(ErrInvalidJSON, parseErr)}
	}
	if int64(len(out)) > opt.MaxOutputBytes {
		return Result{}, &FixError{Code: "output_too_large", Message: fmt.Sprintf("output exceeds limit (%d > %d)", len(out), opt.MaxOutputBytes), Cause: ErrOutputTooLarge}
	}
	rep.ValidJSON = true
	if rootKind == RootUnknown {
		rootKind = kind
	}
	return Result{Output: out, Root: rootKind, Report: rep}, nil
}

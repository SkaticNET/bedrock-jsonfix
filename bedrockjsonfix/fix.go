package bedrockjsonfix

import (
	"context"
	"errors"
	"fmt"
	"io"
)

const readerChunkSize = 32 << 10

var errRootRejected = errors.New("root rejected by validator")

func contextCanceledError() *FixError {
	return &FixError{Code: "context_canceled", Message: "operation canceled", Cause: ErrContextCanceled}
}

func inputTooLargeError(size, limit int64) *FixError {
	return &FixError{Code: "input_too_large", Message: fmt.Sprintf("input exceeds limit (%d > %d)", size, limit), Cause: ErrInputTooLarge}
}

func readAllWithContext(ctx context.Context, r io.Reader, maxInputBytes int64) ([]byte, error) {
	capHint := readerChunkSize
	if maxInputBytes < int64(capHint) {
		capHint = int(maxInputBytes)
	}
	data := make([]byte, 0, capHint)
	buf := make([]byte, readerChunkSize)
	for {
		select {
		case <-ctx.Done():
			return nil, contextCanceledError()
		default:
		}

		n, err := r.Read(buf)
		if n > 0 {
			size := int64(len(data)) + int64(n)
			if size > maxInputBytes {
				return nil, inputTooLargeError(size, maxInputBytes)
			}
			data = append(data, buf[:n]...)
			if ctx.Err() != nil {
				return nil, contextCanceledError()
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return data, nil
			}
			if ctx.Err() != nil {
				return nil, contextCanceledError()
			}
			return nil, err
		}
	}
}

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

func trimAfterFirstRootCandidate(candidate []byte, opt Options) ([]byte, RootKind, Report, bool) {
	var rep Report
	oldLen := len(candidate)
	start, end, kind, er, extractErr := ExtractFirstJSONValue(candidate, opt)
	if extractErr == nil {
		rep = er
		rep.TrimmedLeadingJunkBytes += start
		rep.TrimmedTrailingJunkBytes += oldLen - end
		return candidate[start:end], kind, rep, true
	}

	if end, ok, _ := trimAfterFirstValueUsingDecoder(candidate); ok && end <= oldLen {
		clampedEnd := minInt(end, oldLen)
		rep.TrimmedTrailingJunkBytes = oldLen - clampedEnd
		return candidate[:clampedEnd], RootUnknown, rep, true
	}

	return candidate, RootUnknown, rep, false
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
		return Result{}, contextCanceledError()
	default:
	}
	data, err := readAllWithContext(ctx, r, opt.MaxInputBytes)
	if err != nil {
		return Result{}, err
	}
	return FixBytes(data, opt)
}

// FixBytes normalizes tolerant JSON-ish bytes into strict JSON.
func FixBytes(input []byte, opt Options) (Result, error) {
	if err := opt.Validate(); err != nil {
		return Result{}, err
	}
	if int64(len(input)) > opt.MaxInputBytes {
		return Result{}, inputTooLargeError(int64(len(input)), opt.MaxInputBytes)
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
	decodeOpt := opt
	if opt.Mode == ModeStrict {
		decodeOpt.AllowCP1252Fallback = false
	}
	decoded, err := decodeInput(input, decodeOpt, &rep)
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
	isBedrockMode := opt.Mode == ModeBedrock || opt.Mode == ModeBedrockSafe
	scanCandidate := candidate
	scanRep := rep
	if isBedrockMode {
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
		scanCandidate = candidate
		scanRep = rep
		if opt.TrimAfterFirstRoot {
			trimmed, kind, er, ok := trimAfterFirstRootCandidate(candidate, opt)
			mergeReport(&rep, er)
			if ok {
				candidate = trimmed
				if kind != RootUnknown {
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
	if isBedrockMode {
		out, kind, parseErr = parseCandidate(candidate, opt)
	} else {
		out, kind, parseErr = parseAndMarshal(candidate, opt)
	}
	if parseErr != nil && isBedrockMode && shouldScanAfterFailure(parseErr, opt, scanCandidate) {
		rep = scanRep
		rootKind = RootUnknown
		rep.RootScanUsed = true
		maxCandidates := opt.effectiveRootScanMaxCandidates()
		scanFrom := 0
		baseLeading := rep.TrimmedLeadingJunkBytes
		for attempt := 1; attempt <= maxCandidates; attempt++ {
			next := nextRootCandidate(scanCandidate, scanFrom+1)
			if next < 0 {
				break
			}
			rep.RootScanAttemptsUsed = attempt
			scanFrom = next
			rep.TrimmedLeadingJunkBytes = baseLeading + next
			trimmed := scanCandidate[next:]
			var trimRep Report
			if opt.TrimAfterFirstRoot {
				var ok bool
				trimmed, _, trimRep, ok = trimAfterFirstRootCandidate(trimmed, opt)
				if !ok {
					continue
				}
			}
			out, kind, parseErr = parseCandidate(trimmed, opt)
			if parseErr == nil {
				mergeReport(&rep, trimRep)
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

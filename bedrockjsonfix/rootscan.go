package bedrockjsonfix

import (
	"encoding/json"
	"errors"
	"strings"
)

func firstRootStartOutsideStrings(input []byte, from int) int {
	if from < 0 {
		from = 0
	}
	inStr := false
	esc := false
	for i := 0; i < len(input); i++ {
		c := input[i]
		if inStr {
			if c == '\\' && !esc {
				esc = true
				continue
			}
			if c == '"' && !esc {
				inStr = false
			}
			esc = false
			continue
		}
		if c == '"' {
			inStr = true
			continue
		}
		if (c == '{' || c == '[') && i >= from {
			return i
		}
	}
	return -1
}

func nextRootCandidate(input []byte, from int) int {
	for {
		idx := firstRootStartOutsideStrings(input, from)
		if idx < 0 {
			return -1
		}
		j := idx - 1
		for j >= 0 && isSpace(input[j]) {
			j--
		}
		if j >= 0 {
			prev := input[j]
			if prev == ':' || prev == ',' || prev == '{' || prev == '[' {
				from = idx + 1
				continue
			}
		}
		return idx
	}
}

func shouldScanAfterFailure(parseErr error, opt Options, candidate []byte) bool {
	switch opt.RootPolicy {
	case RootPolicyFirst:
		return false
	case RootPolicyScanBestEffort:
		return opt.effectiveRootScanMaxCandidates() > 0
	case RootPolicyScanLeadingJunk:
		if errors.Is(parseErr, errRootRejected) {
			return opt.effectiveRootScanMaxCandidates() > 0
		}
		if isMultipleDocsWrongStart(parseErr, candidate) {
			return opt.effectiveRootScanMaxCandidates() > 0
		}
		if !isLikelyWrongRootStart(parseErr, opt.WrongStartMaxOffset) {
			return false
		}
		return opt.effectiveRootScanMaxCandidates() > 0
	default:
		return false
	}
}

func isMultipleDocsWrongStart(parseErr error, candidate []byte) bool {
	if parseErr == nil || !strings.Contains(parseErr.Error(), "multiple documents") {
		return false
	}
	i := 0
	for i < len(candidate) && isSpace(candidate[i]) {
		i++
	}
	if i >= len(candidate) {
		return false
	}
	return candidate[i] != '{' && candidate[i] != '['
}

func isLikelyWrongRootStart(parseErr error, maxOffset int64) bool {
	if parseErr == nil {
		return false
	}
	var syn *json.SyntaxError
	if errors.As(parseErr, &syn) && syn.Offset > 0 && syn.Offset <= maxOffset {
		return true
	}
	msg := parseErr.Error()
	return strings.Contains(msg, "looking for beginning of value")
}

func acceptRootCandidate(opt Options, kind RootKind, raw []byte, parsed any) bool {
	if opt.RootValidator == nil {
		return true
	}
	return opt.RootValidator(kind, raw, parsed)
}

func mergeReport(dst *Report, src Report) {
	dst.TrimmedLeadingJunkBytes += src.TrimmedLeadingJunkBytes
	dst.TrimmedTrailingJunkBytes += src.TrimmedTrailingJunkBytes
}

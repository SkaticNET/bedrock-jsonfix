package bedrockjsonfix

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestPreserveIfValid(t *testing.T) {
	opt := DefaultOptions()
	input := []byte("{\"a\":1}\n")
	res, err := FixBytes(input, opt)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(input, res.Output) {
		t.Fatalf("expected identical output")
	}
}

func TestCommentsStripping(t *testing.T) {
	opt := DefaultOptions()
	res, err := FixBytes([]byte("{//x\n\"a\":1/*y*/}\n"), opt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.Output), "\"a\": 1") {
		t.Fatalf("unexpected output: %s", res.Output)
	}
	if res.Report.StrippedLineComments == 0 || res.Report.StrippedBlockComments == 0 {
		t.Fatalf("expected comment counters")
	}
}

func TestTrailingCommaRemoval(t *testing.T) {
	opt := DefaultOptions()
	res, err := FixBytes([]byte("{\"a\":1,}\n"), opt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.Output), "\"a\": 1") {
		t.Fatalf("unexpected output: %s", res.Output)
	}
	if res.Report.RemovedTrailingCommas == 0 {
		t.Fatalf("expected trailing comma removal")
	}
}

func TestIllegalNewlineInsideString(t *testing.T) {
	opt := DefaultOptions()
	in := []byte("{\"a\":\"hello\nworld\"}")
	res, err := FixBytes(in, opt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.Output), `hello\nworld`) {
		t.Fatalf("newline was not normalized: %s", res.Output)
	}
}

func TestControlCharInsideString(t *testing.T) {
	opt := DefaultOptions()
	in := []byte{'{', '"', 'a', '"', ':', '"', 0x01, '"', '}'}
	res, err := FixBytes(in, opt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.Output), `\u0001`) {
		t.Fatalf("control char not escaped: %s", res.Output)
	}
}

func TestGarbageBeforeRoot(t *testing.T) {
	opt := DefaultOptions()
	res, err := FixBytes([]byte("garbage...{\"a\":1}"), opt)
	if err != nil {
		t.Fatal(err)
	}
	if res.Report.TrimmedLeadingJunkBytes == 0 {
		t.Fatalf("expected trimmed leading junk")
	}
}

func TestGarbageAfterRoot(t *testing.T) {
	opt := DefaultOptions()
	res, err := FixBytes([]byte("{\"a\":1} trailing"), opt)
	if err != nil {
		t.Fatal(err)
	}
	if res.Report.TrimmedTrailingJunkBytes == 0 {
		t.Fatalf("expected trailing junk trim")
	}
}

func TestRootScanFallback(t *testing.T) {
	opt := DefaultOptions()
	opt.TrimToFirstRoot = false
	res, err := FixBytes([]byte("{{oops} {\"ok\":true}"), opt)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Report.RootScanUsed {
		t.Fatalf("expected root scan fallback")
	}
	if !strings.Contains(string(res.Output), `"ok": true`) {
		t.Fatalf("unexpected output: %s", res.Output)
	}
}

func TestMaxInputBytes(t *testing.T) {
	opt := DefaultOptions()
	opt.MaxInputBytes = 5
	_, err := FixBytes([]byte("{\"a\":1}"), opt)
	if !errors.Is(err, ErrInputTooLarge) {
		t.Fatalf("expected ErrInputTooLarge, got %v", err)
	}
}

func TestMaxOutputBytes(t *testing.T) {
	opt := DefaultOptions()
	opt.MaxOutputBytes = 5
	opt.PreserveIfValid = false
	_, err := FixBytes([]byte("{\"a\":1}"), opt)
	if !errors.Is(err, ErrOutputTooLarge) {
		t.Fatalf("expected ErrOutputTooLarge, got %v", err)
	}
}

func TestPreserveIfValidHonorsMaxOutputBytes(t *testing.T) {
	opt := DefaultOptions()
	opt.MaxOutputBytes = 5
	_, err := FixBytes([]byte("{\"a\":1}\n"), opt)
	if !errors.Is(err, ErrOutputTooLarge) {
		t.Fatalf("expected ErrOutputTooLarge, got %v", err)
	}
}

func TestModeStrictRejectsBOMPrefixedJSON(t *testing.T) {
	opt := DefaultOptions()
	opt.Mode = ModeStrict
	opt.PreserveIfValid = false
	in := append([]byte{0xEF, 0xBB, 0xBF}, []byte("{\"a\":1}")...)
	_, err := FixBytes(in, opt)
	if !errors.Is(err, ErrInvalidJSON) {
		t.Fatalf("expected ErrInvalidJSON, got %v", err)
	}
}

func TestTrimToFirstRootIgnoresBracesInsideStringsAtStart(t *testing.T) {
	opt := DefaultOptions()
	res, err := FixBytes([]byte(`"{not root}"  {"a":1}`), opt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.Output), `"a": 1`) {
		t.Fatalf("unexpected output: %s", res.Output)
	}
	if res.Report.TrimmedLeadingJunkBytes == 0 {
		t.Fatalf("expected trimmed leading junk")
	}
}

func TestTrimToFirstRootIgnoresBracesInsideStrings(t *testing.T) {
	opt := DefaultOptions()
	res, err := FixBytes([]byte(`"x{y"  {"a":1}`), opt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.Output), `"a": 1`) {
		t.Fatalf("unexpected output: %s", res.Output)
	}
}

func TestRootScanSkipsNestedRoots(t *testing.T) {
	opt := DefaultOptions()
	opt.TrimToFirstRoot = false
	opt.TrimAfterFirstRoot = false
	in := []byte(`{"a":{"b":1},"c":oops} garbage {"ok":true}`)
	res, err := FixBytes(in, opt)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Report.RootScanUsed {
		t.Fatalf("expected root scan fallback")
	}
	if !strings.Contains(string(res.Output), `"ok": true`) {
		t.Fatalf("expected to land on second root, got: %s", res.Output)
	}
}

func TestDropUnknownOutsideStringsRemovesUnderscorePlusBackslash(t *testing.T) {
	opt := DefaultOptions()
	opt.TrimToFirstRoot = false
	opt.TrimAfterFirstRoot = false
	in := []byte(`{_ "a":+ \ }{"ok":true}`)
	res, err := FixBytes(in, opt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.Output), `"ok": true`) {
		t.Fatalf("unexpected output: %s", res.Output)
	}
	if res.Report.DroppedJunkOutsideStrings == 0 {
		t.Fatalf("expected junk to be dropped")
	}
}

func TestTrimToFirstRootHandlesLeadingJunkQuote(t *testing.T) {
	opt := DefaultOptions()
	_, err := FixBytes([]byte(`" {"a":1}`), opt)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNoRootFound) {
		t.Fatalf("expected ErrNoRootFound, got %v", err)
	}
}

func TestDropUnknownOutsideStringsLiteralBoundary(t *testing.T) {
	opt := DefaultOptions()
	opt.TrimToFirstRoot = false
	opt.TrimAfterFirstRoot = false
	in := []byte(`{"a":truely} {"ok":true}`)
	res, err := FixBytes(in, opt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.Output), `"ok": true`) {
		t.Fatalf("expected to recover later root, got: %s", res.Output)
	}
}

func TestTrailingStrayQuoteAfterRootIsTrimmed(t *testing.T) {
	opt := DefaultOptions()
	res, err := FixBytes([]byte(`{"a":1}"`), opt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.Output), `"a": 1`) {
		t.Fatalf("unexpected output: %s", res.Output)
	}
	if res.Report.TrimmedTrailingJunkBytes == 0 {
		t.Fatalf("expected trailing junk trim")
	}
}

func TestTrailingGarbageAfterRootIsTrimmedEvenIfNotValidJSONToken(t *testing.T) {
	opt := DefaultOptions()
	res, err := FixBytes([]byte(`{"a":1} $$$`), opt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.Output), `"a": 1`) {
		t.Fatalf("unexpected output: %s", res.Output)
	}
	if res.Report.TrimmedTrailingJunkBytes == 0 {
		t.Fatalf("expected trailing junk trim")
	}
}

func TestDecoderTrimDoesNotBreakValidMultiDocDetectionWhenTrimAfterFirstRootIsFalse(t *testing.T) {
	opt := DefaultOptions()
	opt.TrimAfterFirstRoot = false
	opt.DropJunkOutsideStrings = false
	opt.TrimToFirstRoot = false
	opt.RootScanAttempts = 0
	_, err := FixBytes([]byte(`{"a":1} {"b":2}`), opt)
	if err == nil {
		t.Fatal("expected invalid json error")
	}
	if !errors.Is(err, ErrInvalidJSON) {
		t.Fatalf("expected ErrInvalidJSON, got %v", err)
	}
}

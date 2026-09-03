// Package bedrockjsonfix turns Bedrock-style JSON into valid JSON bytes.
package bedrockjsonfix

import (
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"
)

const maxInputBytes = 128 << 20

var (
	errInputTooLarge = errors.New("input too large")
	errInvalidJSON   = errors.New("invalid JSON")
)

// Fix returns one valid JSON object or array without changing JSON member names.
func Fix(input []byte) ([]byte, error) {
	if len(input) > maxInputBytes {
		return nil, fmt.Errorf("%w: %d bytes exceeds %d bytes", errInputTooLarge, len(input), maxInputBytes)
	}
	input = stripBOM(input)
	if !utf8.Valid(input) {
		input = repairInvalidUTF8AsWindows1252(input)
	}
	if json.Valid(input) {
		return append([]byte(nil), input...), nil
	}

	cleaned := cleanOutsideStrings(input)
	var err error
	cleaned, err = stripComments(cleaned)
	if err != nil {
		return nil, err
	}
	cleaned = escapeStringControls(cleaned)
	cleaned = removeTrailingCommas(cleaned)

	start, end, err := firstDocument(cleaned)
	if err != nil {
		return nil, err
	}
	out := append([]byte(nil), cleaned[start:end]...)
	if !json.Valid(out) {
		return nil, errInvalidJSON
	}
	return out, nil
}

var windows1252C1 = [...]rune{
	'\u20AC', '\u0081', '\u201A', '\u0192',
	'\u201E', '\u2026', '\u2020', '\u2021',
	'\u02C6', '\u2030', '\u0160', '\u2039',
	'\u0152', '\u008D', '\u017D', '\u008F',
	'\u0090', '\u2018', '\u2019', '\u201C',
	'\u201D', '\u2022', '\u2013', '\u2014',
	'\u02DC', '\u2122', '\u0161', '\u203A',
	'\u0153', '\u009D', '\u017E', '\u0178',
}

func repairInvalidUTF8AsWindows1252(input []byte) []byte {
	out := make([]byte, 0, len(input)+16)
	for len(input) > 0 {
		r, size := utf8.DecodeRune(input)
		if r != utf8.RuneError || size > 1 {
			out = append(out, input[:size]...)
			input = input[size:]
			continue
		}

		b := input[0]
		input = input[1:]
		if b >= 0x80 && b <= 0x9F {
			r = windows1252C1[int(b-0x80)]
		} else {
			r = rune(b)
		}
		out = utf8.AppendRune(out, r)
	}
	return out
}

func stripBOM(input []byte) []byte {
	if len(input) >= 3 && input[0] == 0xef && input[1] == 0xbb && input[2] == 0xbf {
		return input[3:]
	}
	return input
}

func cleanOutsideStrings(input []byte) []byte {
	out := make([]byte, 0, len(input))
	inString := false
	escaped := false
	for i := 0; i < len(input); {
		if input[i] < utf8.RuneSelf {
			c := input[i]
			i++
			if !inString && c < 0x20 && c != '\t' && c != '\n' && c != '\r' {
				continue
			}
			out = append(out, c)
			inString, escaped = nextStringState(inString, escaped, c)
			continue
		}

		r, size := utf8.DecodeRune(input[i:])
		if !inString {
			switch r {
			case '\u00a0':
				out = append(out, ' ')
				i += size
				continue
			case '\u200b', '\u200c', '\u200d', '\u2060':
				i += size
				continue
			}
		}
		out = append(out, input[i:i+size]...)
		i += size
	}
	return out
}

func stripComments(input []byte) ([]byte, error) {
	out := make([]byte, 0, len(input))
	inString := false
	escaped := false
	for i := 0; i < len(input); {
		c := input[i]
		if !inString && c == '/' && i+1 < len(input) {
			switch input[i+1] {
			case '/':
				i += 2
				for i < len(input) && input[i] != '\n' && input[i] != '\r' {
					i++
				}
				if i < len(input) {
					out = append(out, '\n')
					if input[i] == '\r' && i+1 < len(input) && input[i+1] == '\n' {
						i++
					}
					i++
				}
				continue
			case '*':
				i += 2
				for i+1 < len(input) && (input[i] != '*' || input[i+1] != '/') {
					i++
				}
				if i+1 >= len(input) {
					return nil, fmt.Errorf("%w: unterminated block comment", errInvalidJSON)
				}
				out = append(out, ' ')
				i += 2
				continue
			}
		}
		out = append(out, c)
		inString, escaped = nextStringState(inString, escaped, c)
		i++
	}
	return out, nil
}

func escapeStringControls(input []byte) []byte {
	out := make([]byte, 0, len(input))
	inString := false
	escaped := false
	for _, c := range input {
		if inString && !escaped && c < 0x20 {
			switch c {
			case '\b':
				out = append(out, '\\', 'b')
			case '\f':
				out = append(out, '\\', 'f')
			case '\n':
				out = append(out, '\\', 'n')
			case '\r':
				out = append(out, '\\', 'r')
			case '\t':
				out = append(out, '\\', 't')
			default:
				out = append(out, '\\', 'u', '0', '0', "0123456789ABCDEF"[c>>4], "0123456789ABCDEF"[c&0x0f])
			}
			continue
		}
		out = append(out, c)
		inString, escaped = nextStringState(inString, escaped, c)
	}
	return out
}

func removeTrailingCommas(input []byte) []byte {
	out := make([]byte, 0, len(input))
	inString := false
	escaped := false
	for i, c := range input {
		if !inString && c == ',' {
			j := i + 1
			for j < len(input) && isJSONSpace(input[j]) {
				j++
			}
			if j < len(input) && (input[j] == '}' || input[j] == ']') {
				continue
			}
		}
		out = append(out, c)
		inString, escaped = nextStringState(inString, escaped, c)
	}
	return out
}

func firstDocument(input []byte) (int, int, error) {
	inString := false
	escaped := false
	for start, c := range input {
		inString, escaped = nextStringState(inString, escaped, c)
		if inString || c != '{' && c != '[' {
			continue
		}
		stack := []byte{c}
		candidateInString := false
		candidateEscaped := false
		for i := start + 1; i < len(input); i++ {
			current := input[i]
			if candidateInString {
				candidateInString, candidateEscaped = nextStringState(candidateInString, candidateEscaped, current)
				continue
			}
			switch current {
			case '"':
				candidateInString = true
			case '{', '[':
				stack = append(stack, current)
			case '}':
				if stack[len(stack)-1] != '{' {
					return 0, 0, errInvalidJSON
				}
				stack = stack[:len(stack)-1]
			case ']':
				if stack[len(stack)-1] != '[' {
					return 0, 0, errInvalidJSON
				}
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return start, i + 1, nil
			}
		}
		return 0, 0, errInvalidJSON
	}
	return 0, 0, errInvalidJSON
}

func nextStringState(inString, escaped bool, c byte) (bool, bool) {
	if c == '"' && !escaped {
		return !inString, false
	}
	if c == '\\' && !escaped {
		return inString, true
	}
	return inString, false
}

func isJSONSpace(c byte) bool {
	return c == ' ' || c == '\n' || c == '\r' || c == '\t'
}

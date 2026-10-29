// Package bedrockjsonfix turns Bedrock-style JSON into valid JSON bytes.
package bedrockjsonfix

import (
	"bytes"
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

// Fix returns one valid JSON value without changing JSON member names.
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
	if start, scalar := firstScalar(cleaned); scalar {
		value := normalizeLeadingZeros(escapeStringControls(cleaned[start:]))
		var first json.RawMessage
		if err := json.NewDecoder(bytes.NewReader(value)).Decode(&first); err != nil {
			return nil, errInvalidJSON
		}
		return first, nil
	}
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
	out := normalizeLeadingZeros(cleaned[start:end])
	if !json.Valid(out) {
		return nil, errInvalidJSON
	}
	return out, nil
}

func normalizeLeadingZeros(input []byte) []byte {
	out := make([]byte, 0, len(input))
	inString, escaped := false, false
	for i := 0; i < len(input); {
		c := input[i]
		if !inString && (c == '-' || (c >= '0' && c <= '9')) {
			start := i
			for i < len(input) {
				c = input[i]
				if (c < '0' || c > '9') && c != '-' && c != '+' && c != '.' && c != 'e' && c != 'E' {
					break
				}
				i++
			}
			digits := start
			if input[digits] == '-' {
				out = append(out, '-')
				digits++
			}
			// Keep numeric text intact apart from redundant integer zeros.
			for digits+1 < i && input[digits] == '0' && input[digits+1] >= '0' && input[digits+1] <= '9' {
				digits++
			}
			out = append(out, input[digits:i]...)
			continue
		}
		out = append(out, c)
		inString, escaped = nextStringState(inString, escaped, c)
		i++
	}
	return out
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
	depth := 0
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
		if !inString {
			switch c {
			case '{', '[':
				depth++
			case '}', ']':
				if depth > 0 {
					depth--
					if depth == 0 {
						// Trailing input cannot invalidate the first completed document.
						return out, nil
					}
				}
			}
		}
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

// firstScalar only recognizes a scalar at the first non-comment token. Other
// prefixes remain the responsibility of the existing compound recovery path.
func firstScalar(input []byte) (int, bool) {
	i := 0
	for i < len(input) {
		if isJSONSpace(input[i]) {
			i++
			continue
		}
		if input[i] == '/' && i+1 < len(input) {
			if input[i+1] == '/' {
				i += 2
				for i < len(input) && input[i] != '\n' && input[i] != '\r' {
					i++
				}
				continue
			}
			if input[i+1] == '*' {
				i += 2
				for i+1 < len(input) && (input[i] != '*' || input[i+1] != '/') {
					i++
				}
				if i+1 >= len(input) {
					return 0, false
				}
				i += 2
				continue
			}
		}
		break
	}
	if i == len(input) {
		return 0, false
	}
	c := input[i]
	return i, c == '"' || c == '-' || c >= '0' && c <= '9' || c == 'n' || c == 't' || c == 'f'
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

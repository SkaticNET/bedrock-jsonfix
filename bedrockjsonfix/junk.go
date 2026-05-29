package bedrockjsonfix

import "bytes"

func dropUnknownOutsideStrings(input []byte, rep *Report) []byte {
	var b bytes.Buffer
	changed := false
	inStr := false
	esc := false
	for i := 0; i < len(input); i++ {
		c := input[i]
		if !inStr {
			if isJSONTokenOutsideString(c) {
				if changed {
					b.WriteByte(c)
				}
			} else if isAlpha(c) {
				switch {
				case hasLiteralToken(input, i, "true"):
					if changed {
						b.WriteString("true")
					}
					i += 3
				case hasLiteralToken(input, i, "false"):
					if changed {
						b.WriteString("false")
					}
					i += 4
				case hasLiteralToken(input, i, "null"):
					if changed {
						b.WriteString("null")
					}
					i += 3
				default:
					if !changed {
						b.Grow(len(input))
						b.Write(input[:i])
						changed = true
					}
					j := i
					for j < len(input) && isAlpha(input[j]) {
						rep.DroppedJunkOutsideStrings++
						j++
					}
					if b.Len() == 0 || b.Bytes()[b.Len()-1] != ' ' {
						b.WriteByte(' ')
					}
					i = j - 1
				}
			} else {
				if !changed {
					b.Grow(len(input))
					b.Write(input[:i])
					changed = true
				}
				rep.DroppedJunkOutsideStrings++
				if b.Len() == 0 || b.Bytes()[b.Len()-1] != ' ' {
					b.WriteByte(' ')
				}
			}
		} else {
			if changed {
				b.WriteByte(c)
			}
		}
		if c == '"' && !esc {
			inStr = !inStr
		}
		if c == '\\' && !esc {
			esc = true
		} else {
			esc = false
		}
	}
	if !changed {
		return input
	}
	return b.Bytes()
}

func hasLiteralToken(input []byte, start int, lit string) bool {
	end := start + len(lit)
	if end > len(input) {
		return false
	}
	for i := 0; i < len(lit); i++ {
		if input[start+i] != lit[i] {
			return false
		}
	}
	if end < len(input) && isAlpha(input[end]) {
		return false
	}
	return true
}

func isJSONTokenOutsideString(c byte) bool {
	if isSpace(c) {
		return true
	}
	if (c >= '0' && c <= '9') || c == '-' || c == '.' || c == 'e' || c == 'E' {
		return true
	}
	switch c {
	case '{', '}', '[', ']', ':', ',', '"':
		return true
	default:
		return false
	}
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isSpace(c byte) bool { return c == ' ' || c == '\n' || c == '\r' || c == '\t' }

package bedrockjsonfix

import "bytes"

func dropUnknownOutsideStrings(input []byte, rep *Report) []byte {
	var b bytes.Buffer
	inStr := false
	esc := false
	for i := 0; i < len(input); i++ {
		c := input[i]
		if !inStr {
			if isJSONTokenOutsideString(c) {
				b.WriteByte(c)
			} else if isAlpha(c) {
				switch {
				case hasLiteralToken(input, i, "true"):
					b.WriteString("true")
					i += 3
				case hasLiteralToken(input, i, "false"):
					b.WriteString("false")
					i += 4
				case hasLiteralToken(input, i, "null"):
					b.WriteString("null")
					i += 3
				default:
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
				rep.DroppedJunkOutsideStrings++
				if b.Len() == 0 || b.Bytes()[b.Len()-1] != ' ' {
					b.WriteByte(' ')
				}
			}
		} else {
			b.WriteByte(c)
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
	return b.Bytes()
}

func hasLiteralToken(input []byte, start int, lit string) bool {
	if !bytes.HasPrefix(input[start:], []byte(lit)) {
		return false
	}
	end := start + len(lit)
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

package bedrockjsonfix

import (
	"unicode/utf8"
)

func sanitize(input []byte, opt Options, rep *Report) []byte {
	start := 0
	if len(input) >= 3 && input[0] == 0xEF && input[1] == 0xBB && input[2] == 0xBF {
		start = 3
		rep.RemovedBOM++
	}
	var out []byte
	inStr := false
	esc := false
	ensureOut := func(i int) {
		if out == nil {
			out = make([]byte, 0, len(input)-start)
			out = append(out, input[start:i]...)
		}
	}
	for i := start; i < len(input); {
		r, sz := utf8.DecodeRune(input[i:])
		if r == utf8.RuneError && sz == 1 {
			r = rune(input[i])
			ensureOut(i)
			out = utf8.AppendRune(out, r)
			i += sz
			continue
		}
		outside := !inStr || opt.AggressiveWhitespace
		if outside {
			if r == '\u00A0' {
				ensureOut(i)
				out = append(out, ' ')
				rep.ReplacedNBSP++
				i += sz
				continue
			}
			if r == '\u200B' || r == '\u200C' || r == '\u200D' || r == '\u2060' {
				ensureOut(i)
				rep.RemovedZeroWidth++
				i += sz
				continue
			}
		}
		if !inStr {
			if r == '\r' {
				ensureOut(i)
				rep.NormalizedCRLF++
				next := i + sz
				if next < len(input) && input[next] == '\n' {
					next++
				}
				out = append(out, '\n')
				i = next
				continue
			}
			if r < 0x20 && r != '\n' && r != '\t' {
				ensureOut(i)
				rep.RemovedASCIIControls++
				i += sz
				continue
			}
		}
		if r == '"' && !esc {
			inStr = !inStr
		}
		if inStr && r == '\\' && !esc {
			esc = true
		} else {
			esc = false
		}
		if out != nil {
			out = append(out, input[i:i+sz]...)
		}
		i += sz
	}
	if out != nil {
		return out
	}
	return input[start:]
}

package bedrockjsonfix

import (
	"bytes"
	"unicode/utf8"
)

func sanitize(input []byte, opt Options, rep *Report) []byte {
	if bytes.HasPrefix(input, []byte{0xEF, 0xBB, 0xBF}) {
		input = input[3:]
		rep.RemovedBOM++
	}
	var b bytes.Buffer
	b.Grow(len(input))
	inStr := false
	esc := false
	for i := 0; i < len(input); {
		r, sz := utf8.DecodeRune(input[i:])
		if r == utf8.RuneError && sz == 1 {
			r = rune(input[i])
		}
		outside := !inStr || opt.AggressiveWhitespace
		if outside {
			if r == '\u00A0' {
				b.WriteByte(' ')
				rep.ReplacedNBSP++
				i += sz
				continue
			}
			if r == '\u200B' || r == '\u200C' || r == '\u200D' || r == '\u2060' {
				rep.RemovedZeroWidth++
				i += sz
				continue
			}
		}
		if !inStr {
			if r == '\r' {
				rep.NormalizedCRLF++
				if i+1 < len(input) && input[i+1] == '\n' {
					i++
				}
				b.WriteByte('\n')
				i += sz
				continue
			}
			if r < 0x20 && r != '\n' && r != '\t' {
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
		b.WriteRune(r)
		i += sz
	}
	return b.Bytes()
}

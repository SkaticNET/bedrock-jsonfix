package bedrockjsonfix

import (
	"bytes"
	"fmt"
)

func escapeStringControls(input []byte, rep *Report) []byte {
	var b bytes.Buffer
	inStr := false
	esc := false
	for i := 0; i < len(input); i++ {
		c := input[i]
		if inStr && !esc && c < 0x20 {
			rep.EscapedStringControls++
			switch c {
			case '\n':
				b.WriteString(`\n`)
			case '\r':
				b.WriteString(`\r`)
			case '\t':
				b.WriteString(`\t`)
			default:
				b.WriteString(fmt.Sprintf(`\u%04X`, c))
			}
			continue
		}
		b.WriteByte(c)
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

func normalizeLiteralNewlinesInStrings(input []byte, rep *Report) []byte {
	var b bytes.Buffer
	inStr := false
	esc := false
	for i := 0; i < len(input); i++ {
		c := input[i]
		if inStr && !esc && c == '\n' {
			rep.NormalizedNewlinesInStrings++
			b.WriteString(`\n`)
			continue
		}
		b.WriteByte(c)
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

package bedrockjsonfix

import (
	"bytes"
)

const hexUpper = "0123456789ABCDEF"

func writeEscapedControl(b *bytes.Buffer, c byte) {
	var esc [6]byte
	esc[0] = '\\'
	esc[1] = 'u'
	esc[2] = '0'
	esc[3] = '0'
	esc[4] = hexUpper[c>>4]
	esc[5] = hexUpper[c&0x0F]
	b.Write(esc[:])
}

func escapeStringControls(input []byte, rep *Report) []byte {
	var b bytes.Buffer
	b.Grow(len(input))
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
				writeEscapedControl(&b, c)
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
	b.Grow(len(input))
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

package bedrockjsonfix

func stripComments(input []byte, rep *Report) []byte {
	var out []byte
	inStr := false
	esc := false
	for i := 0; i < len(input); i++ {
		c := input[i]
		if !inStr && c == '/' && i+1 < len(input) {
			n := input[i+1]
			if n == '/' {
				rep.StrippedLineComments++
				if out == nil {
					out = make([]byte, 0, len(input))
					out = append(out, input[:i]...)
				}
				i += 2
				for i < len(input) && input[i] != '\n' {
					i++
				}
				if i < len(input) {
					out = append(out, '\n')
				}
				continue
			}
			if n == '*' {
				rep.StrippedBlockComments++
				if out == nil {
					out = make([]byte, 0, len(input))
					out = append(out, input[:i]...)
				}
				i += 2
				for i+1 < len(input) && (input[i] != '*' || input[i+1] != '/') {
					i++
				}
				i++
				continue
			}
		}
		if out != nil {
			out = append(out, c)
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
	if out == nil {
		return input
	}
	return out
}

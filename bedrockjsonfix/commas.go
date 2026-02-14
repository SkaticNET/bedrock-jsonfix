package bedrockjsonfix

func removeTrailingCommas(input []byte, rep *Report) []byte {
	var out []byte
	inStr := false
	esc := false
	for i := 0; i < len(input); i++ {
		c := input[i]
		if !inStr && c == ',' {
			j := i + 1
			for j < len(input) && isSpace(input[j]) {
				j++
			}
			if j < len(input) && (input[j] == '}' || input[j] == ']') {
				rep.RemovedTrailingCommas++
				if out == nil {
					out = make([]byte, 0, len(input))
					out = append(out, input[:i]...)
				}
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

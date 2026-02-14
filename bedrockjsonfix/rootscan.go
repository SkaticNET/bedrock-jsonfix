package bedrockjsonfix

func firstRootStartOutsideStrings(input []byte, from int) int {
	inStr := false
	esc := false
	for i := from; i < len(input); i++ {
		c := input[i]
		if inStr {
			if c == '\\' && !esc {
				esc = true
				continue
			}
			if c == '"' && !esc {
				inStr = false
			}
			esc = false
			continue
		}
		if c == '"' {
			inStr = true
			continue
		}
		if c == '{' || c == '[' {
			return i
		}
	}
	return -1
}

func nextRootCandidate(input []byte, from int) int {
	for {
		idx := firstRootStartOutsideStrings(input, from)
		if idx < 0 {
			return -1
		}
		j := idx - 1
		for j >= 0 && isSpace(input[j]) {
			j--
		}
		if j >= 0 {
			prev := input[j]
			if prev == ':' || prev == ',' || prev == '{' || prev == '[' {
				from = idx + 1
				continue
			}
		}
		return idx
	}
}

func mergeReport(dst *Report, src Report) {
	dst.TrimmedLeadingJunkBytes += src.TrimmedLeadingJunkBytes
	dst.TrimmedTrailingJunkBytes += src.TrimmedTrailingJunkBytes
}

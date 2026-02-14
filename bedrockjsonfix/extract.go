package bedrockjsonfix

// ExtractFirstJSONValue returns the [start,end) region containing first complete root JSON object/array.
func ExtractFirstJSONValue(input []byte, _ Options) (start, end int, kind RootKind, rep Report, err error) {
	s := firstRootStartOutsideStrings(input, 0)
	if s < 0 {
		return 0, 0, RootUnknown, rep, &FixError{Code: "no_root", Message: "no JSON root object/array found", Cause: ErrNoRootFound}
	}
	stack := []byte{input[s]}
	if input[s] == '{' {
		kind = RootObject
	} else {
		kind = RootArray
	}
	inStr := false
	esc := false
	for i := s + 1; i < len(input); i++ {
		c := input[i]
		if inStr {
			if c == '"' && !esc {
				inStr = false
			}
			if c == '\\' && !esc {
				esc = true
			} else {
				esc = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{', '[':
			stack = append(stack, c)
		case '}':
			if len(stack) == 0 || stack[len(stack)-1] != '{' {
				return 0, 0, RootUnknown, rep, &FixError{Code: "invalid_json", Message: "mismatched object delimiters", Cause: ErrInvalidJSON}
			}
			stack = stack[:len(stack)-1]
		case ']':
			if len(stack) == 0 || stack[len(stack)-1] != '[' {
				return 0, 0, RootUnknown, rep, &FixError{Code: "invalid_json", Message: "mismatched array delimiters", Cause: ErrInvalidJSON}
			}
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			return s, i + 1, kind, rep, nil
		}
	}
	return 0, 0, RootUnknown, rep, &FixError{Code: "no_root", Message: "incomplete JSON root", Cause: ErrNoRootFound}
}

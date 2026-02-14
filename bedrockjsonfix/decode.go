package bedrockjsonfix

import (
	"bytes"
	"unicode/utf8"
)

var cp1252SpecialRunes = [32]rune{
	'€', '�', '‚', 'ƒ', '„', '…', '†', '‡',
	'ˆ', '‰', 'Š', '‹', 'Œ', '�', 'Ž', '�',
	'�', '‘', '’', '“', '”', '•', '–', '—',
	'˜', '™', 'š', '›', 'œ', '�', 'ž', 'Ÿ',
}

func decodeInput(input []byte, opt Options, rep *Report) ([]byte, error) {
	if utf8.Valid(input) {
		return append([]byte(nil), input...), nil
	}
	rep.InputWasInvalidUTF8 = true
	if !opt.AllowCP1252Fallback {
		return nil, &FixError{Code: "invalid_encoding", Message: "input must be valid UTF-8", Cause: ErrInvalidJSON}
	}
	out := decodeCP1252(input)
	rep.UsedCP1252Fallback = true
	return out, nil
}

func decodeCP1252(input []byte) []byte {
	var b bytes.Buffer
	b.Grow(len(input))
	for _, c := range input {
		if c < 0x80 {
			b.WriteByte(c)
			continue
		}
		r := cp1252Rune(c)
		b.WriteRune(r)
	}
	return b.Bytes()
}

func cp1252Rune(c byte) rune {
	if c >= 0x80 && c <= 0x9F {
		return cp1252SpecialRunes[c-0x80]
	}
	return rune(c)
}

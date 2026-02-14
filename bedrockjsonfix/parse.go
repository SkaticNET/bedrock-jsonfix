package bedrockjsonfix

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func parseAndMarshal(input []byte, opt Options) ([]byte, RootKind, error) {
	out, kind, _, err := parseAndMarshalWithParsed(input, opt)
	return out, kind, err
}

func parseAndMarshalWithParsed(input []byte, opt Options) ([]byte, RootKind, any, error) {
	dec := json.NewDecoder(bytes.NewReader(input))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, RootUnknown, nil, err
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, RootUnknown, nil, errors.New("multiple documents")
		}
		return nil, RootUnknown, nil, err
	}
	kind := RootUnknown
	switch v.(type) {
	case map[string]any:
		kind = RootObject
	case []any:
		kind = RootArray
	}
	var out []byte
	var err error
	if opt.Pretty {
		out, err = json.MarshalIndent(v, opt.Prefix, opt.Indent)
	} else {
		out, err = json.Marshal(v)
	}
	if err != nil {
		return nil, RootUnknown, nil, err
	}
	out = append(out, '\n')
	return out, kind, v, nil
}

func strictJSONSingleDocument(input []byte) (bool, RootKind) {
	dec := json.NewDecoder(bytes.NewReader(input))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return false, RootUnknown
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return false, RootUnknown
	}
	switch v.(type) {
	case map[string]any:
		return true, RootObject
	case []any:
		return true, RootArray
	default:
		return true, RootUnknown
	}
}

func trimAfterFirstValueUsingDecoder(input []byte) (end int, ok bool, err error) {
	dec := json.NewDecoder(bytes.NewReader(input))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return 0, false, err
	}
	off := int(dec.InputOffset())
	if off <= 0 {
		return 0, false, nil
	}
	return off, true, nil
}

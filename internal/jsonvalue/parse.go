package jsonvalue

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Parse decodes a JSON document into a Value, preserving object key insertion
// order (unlike a map-based decode). Numbers that are whole and fit in an int64
// become KindInt64; all others become KindNumber.
func Parse(s string) (*Value, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	v, err := parseValue(dec, tok)
	if err != nil {
		return nil, err
	}

	// Reject trailing content so malformed input like "1 2" is an error.
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("invalid JSON: unexpected trailing content")
		}
		return nil, err
	}
	return v, nil
}

func parseValue(dec *json.Decoder, tok json.Token) (*Value, error) {
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return parseObject(dec)
		case '[':
			return parseArray(dec)
		default:
			return nil, fmt.Errorf("invalid JSON: unexpected %q", t)
		}
	case string:
		return NewString(t), nil
	case json.Number:
		return numberFromJSON(t), nil
	case bool:
		return NewBoolean(t), nil
	case nil:
		return NewNull(), nil
	default:
		return nil, fmt.Errorf("invalid JSON: unexpected token %v", tok)
	}
}

func parseObject(dec *json.Decoder) (*Value, error) {
	obj := NewObject()
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("invalid JSON: object key is not a string")
		}
		valTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		child, err := parseValue(dec, valTok)
		if err != nil {
			return nil, err
		}
		obj.ObjectSet(key, child)
	}
	// Consume the closing '}'.
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return obj, nil
}

func parseArray(dec *json.Decoder) (*Value, error) {
	arr := NewArray()
	for dec.More() {
		valTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		child, err := parseValue(dec, valTok)
		if err != nil {
			return nil, err
		}
		arr.ArrayAppend(child)
	}
	// Consume the closing ']'.
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return arr, nil
}

func numberFromJSON(n json.Number) *Value {
	if i, err := n.Int64(); err == nil {
		return NewInt64(i)
	}
	if f, err := n.Float64(); err == nil {
		return NewNumber(f)
	}
	// Should be unreachable for tokens produced by encoding/json.
	return NewString(n.String())
}

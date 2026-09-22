package evaluator

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

// parseJSONText retains duplicate names in custom Stringify results. Unlike
// JSON.Parse, this snapshot is only written as text and has no lookup semantics.
func parseJSONText(text string) (*jsonTextValue, error) {
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	value, err := readJSONText(decoder)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("unexpected trailing JSON content")
	}
	return value, nil
}

func readJSONText(decoder *json.Decoder) (*jsonTextValue, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch value := token.(type) {
	case json.Delim:
		return readJSONTextContainer(decoder, value)
	case string:
		return jsonTextFromValue(jsonvalue.NewString(value)), nil
	case bool:
		return jsonTextFromValue(jsonvalue.NewBoolean(value)), nil
	case json.Number:
		parsed, err := jsonvalue.Parse(string(value))
		return jsonTextFromValue(parsed), err
	case nil:
		return jsonTextFromValue(jsonvalue.NewNull()), nil
	default:
		return nil, fmt.Errorf("unexpected JSON token %v", token)
	}
}

func readJSONTextContainer(decoder *json.Decoder, opening json.Delim) (*jsonTextValue, error) {
	result := newJSONTextArray()
	if opening == '{' {
		result = newJSONTextObject()
	} else if opening != '[' {
		return nil, fmt.Errorf("unexpected JSON delimiter %q", opening)
	}
	for decoder.More() {
		var name string
		if opening == '{' {
			key, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			var ok bool
			name, ok = key.(string)
			if !ok {
				return nil, fmt.Errorf("expected JSON member name")
			}
		}
		child, err := readJSONText(decoder)
		if err != nil {
			return nil, err
		}
		if opening == '{' {
			result.appendMember(name, child)
		} else {
			result.elements = append(result.elements, child)
		}
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return result, nil
}

// jsonTextValue is a serialization snapshot, not a live JSON object. Members
// are ordered pairs because distinct associative keys can have identical text.
// It never adopts source nodes or changes their ownership.
type jsonTextValue struct {
	scalar   *jsonvalue.Value
	members  []jsonMember
	elements []*jsonTextValue
	kind     jsonvalue.Kind
}

func newJSONTextObject() *jsonTextValue { return &jsonTextValue{kind: jsonvalue.KindObject} }
func newJSONTextArray() *jsonTextValue  { return &jsonTextValue{kind: jsonvalue.KindArray} }

func jsonTextFromValue(value *jsonvalue.Value) *jsonTextValue {
	result := &jsonTextValue{scalar: value, kind: value.Kind()}
	switch value.Kind() {
	case jsonvalue.KindObject:
		for _, key := range value.ObjectKeys() {
			result.members = append(result.members, jsonMember{name: key, jv: jsonTextFromValue(value.ObjectGet(key))})
		}
	case jsonvalue.KindArray:
		for _, child := range value.ArrayElements() {
			result.elements = append(result.elements, jsonTextFromValue(child))
		}
	}
	return result
}

func (v *jsonTextValue) appendMember(name string, child *jsonTextValue) {
	v.members = append(v.members, jsonMember{name: name, jv: child})
}

func (v *jsonTextValue) stringify(indent string, pretty bool) string {
	var out strings.Builder
	v.write(&out, indent, pretty, 0)
	return out.String()
}

func (v *jsonTextValue) write(out *strings.Builder, indent string, pretty bool, depth int) {
	if v.kind != jsonvalue.KindObject && v.kind != jsonvalue.KindArray {
		out.WriteString(jsonvalue.Stringify(v.scalar))
		return
	}
	opening, closing := byte('['), byte(']')
	count := len(v.elements)
	if v.kind == jsonvalue.KindObject {
		opening, closing, count = '{', '}', len(v.members)
	}
	out.WriteByte(opening)
	if count == 0 {
		if pretty {
			out.WriteByte(' ')
		}
		out.WriteByte(closing)
		return
	}
	for i := 0; i < count; i++ {
		if i > 0 {
			out.WriteByte(',')
		}
		if pretty {
			out.WriteString("\r\n")
			out.WriteString(strings.Repeat(indent, depth+1))
		}
		var child *jsonTextValue
		if v.kind == jsonvalue.KindObject {
			member := v.members[i]
			jsonvalue.WriteJSONString(out, member.name)
			if pretty {
				out.WriteString(" : ")
			} else {
				out.WriteByte(':')
			}
			child = member.jv
		} else {
			child = v.elements[i]
		}
		child.write(out, indent, pretty, depth+1)
	}
	if pretty {
		out.WriteString("\r\n")
		out.WriteString(strings.Repeat(indent, depth))
	}
	out.WriteByte(closing)
}

// Package jsonvalue provides an internal representation of DWScript JSON values.
package jsonvalue

import "encoding/json"

// Kind represents the type of a JSON value. It mirrors the DWScript TdwsJSONValueType.
type Kind uint8

const (
	KindUndefined Kind = iota
	KindNull
	KindObject
	KindArray
	KindString
	KindNumber
	KindInt64
	KindBoolean
)

// String returns a human-readable form of the kind.
func (k Kind) String() string {
	switch k {
	case KindUndefined:
		return "Undefined"
	case KindNull:
		return "Null"
	case KindObject:
		return "Object"
	case KindArray:
		return "Array"
	case KindString:
		return "String"
	case KindNumber:
		return "Number"
	case KindInt64:
		return "Int64"
	case KindBoolean:
		return "Boolean"
	default:
		return "Unknown"
	}
}

// Value represents a JSON value in memory. It intentionally avoids using interface{}
// to make downstream use in the interpreter simpler and more type-safe.
type Value struct {
	kind Kind

	// Object fields
	objEntries map[string]*Value
	objKeys    []string // preserves insertion order

	// Array elements
	arrElems []*Value

	// Primitive payloads
	str  string
	num  float64
	i64  int64
	bool bool
}

// Kind returns the kind of the value.
func (v *Value) Kind() Kind {
	if v == nil {
		return KindUndefined
	}
	return v.kind
}

// NewUndefined returns a value flagged as undefined.
func NewUndefined() *Value {
	return &Value{kind: KindUndefined}
}

// NewNull returns a JSON null value.
func NewNull() *Value {
	return &Value{kind: KindNull}
}

// NewBoolean returns a JSON boolean value.
func NewBoolean(b bool) *Value {
	return &Value{kind: KindBoolean, bool: b}
}

// NewNumber returns a JSON number value.
func NewNumber(n float64) *Value {
	return &Value{kind: KindNumber, num: n}
}

// NewInt64 returns a JSON int64 value.
func NewInt64(n int64) *Value {
	return &Value{kind: KindInt64, i64: n}
}

// NewString returns a JSON string value.
func NewString(s string) *Value {
	return &Value{kind: KindString, str: s}
}

// NewArray returns an empty JSON array value.
func NewArray() *Value {
	return &Value{
		kind:     KindArray,
		arrElems: make([]*Value, 0),
	}
}

// NewObject returns an empty JSON object value.
func NewObject() *Value {
	return &Value{
		kind:       KindObject,
		objEntries: make(map[string]*Value),
		objKeys:    make([]string, 0),
	}
}

// ObjectGet returns the value associated with the provided key. Nil is returned
// if the receiver is not an object or the key does not exist.
func (v *Value) ObjectGet(key string) *Value {
	if v == nil || v.kind != KindObject {
		return nil
	}
	return v.objEntries[key]
}

// ObjectSet associates key with child within the object. The method preserves
// insertion order, appending new keys to objKeys. If the key already exists its
// value is replaced in place.
func (v *Value) ObjectSet(key string, child *Value) {
	if v == nil || v.kind != KindObject {
		return
	}
	if _, exists := v.objEntries[key]; !exists {
		v.objKeys = append(v.objKeys, key)
	}
	v.objEntries[key] = child
}

// ObjectDelete removes the entry if present. It returns true when a key was removed.
func (v *Value) ObjectDelete(key string) bool {
	if v == nil || v.kind != KindObject {
		return false
	}
	if _, exists := v.objEntries[key]; !exists {
		return false
	}
	delete(v.objEntries, key)
	for i, k := range v.objKeys {
		if k == key {
			v.objKeys = append(v.objKeys[:i], v.objKeys[i+1:]...)
			break
		}
	}
	return true
}

// ObjectKeys returns the keys of the object in insertion order.
func (v *Value) ObjectKeys() []string {
	if v == nil || v.kind != KindObject {
		return nil
	}
	keys := make([]string, len(v.objKeys))
	copy(keys, v.objKeys)
	return keys
}

// ArrayLen returns the number of elements in the array or zero otherwise.
func (v *Value) ArrayLen() int {
	if v == nil || v.kind != KindArray {
		return 0
	}
	return len(v.arrElems)
}

// ArrayGet returns the element at index or nil if out of bounds.
func (v *Value) ArrayGet(index int) *Value {
	if v == nil || v.kind != KindArray {
		return nil
	}
	if index < 0 || index >= len(v.arrElems) {
		return nil
	}
	return v.arrElems[index]
}

// ArraySet writes the element at index if the receiver is an array and the
// index is valid. It returns true when the assignment succeeded.
func (v *Value) ArraySet(index int, child *Value) bool {
	if v == nil || v.kind != KindArray {
		return false
	}
	if index < 0 || index >= len(v.arrElems) {
		return false
	}
	v.arrElems[index] = child
	return true
}

// ArrayAppend appends an element to the end of the array.
func (v *Value) ArrayAppend(child *Value) {
	if v == nil || v.kind != KindArray {
		return
	}
	v.arrElems = append(v.arrElems, child)
}

// ArrayDelete removes the element at index when valid. It returns true on success.
func (v *Value) ArrayDelete(index int) bool {
	if v == nil || v.kind != KindArray {
		return false
	}
	if index < 0 || index >= len(v.arrElems) {
		return false
	}
	copy(v.arrElems[index:], v.arrElems[index+1:])
	v.arrElems = v.arrElems[:len(v.arrElems)-1]
	return true
}

// ArrayElements returns a shallow copy of the array elements slice.
func (v *Value) ArrayElements() []*Value {
	if v == nil || v.kind != KindArray {
		return nil
	}
	elements := make([]*Value, len(v.arrElems))
	copy(elements, v.arrElems)
	return elements
}

// ============================================================================
// Primitive Value Getters
// ============================================================================

// BoolValue returns the boolean value if this is a KindBoolean, otherwise returns false.
func (v *Value) BoolValue() bool {
	if v == nil || v.kind != KindBoolean {
		return false
	}
	return v.bool
}

// StringValue returns the string value if this is a KindString, otherwise returns empty string.
func (v *Value) StringValue() string {
	if v == nil || v.kind != KindString {
		return ""
	}
	return v.str
}

// NumberValue returns the float64 value if this is a KindNumber, otherwise returns 0.0.
func (v *Value) NumberValue() float64 {
	if v == nil || v.kind != KindNumber {
		return 0.0
	}
	return v.num
}

// Int64Value returns the int64 value if this is a KindInt64, otherwise returns 0.
func (v *Value) Int64Value() int64 {
	if v == nil || v.kind != KindInt64 {
		return 0
	}
	return v.i64
}

// ============================================================================
// JSON Serialization
// MarshalJSON enables Go's encoding/json to serialize jsonvalue.Value
// ============================================================================

// MarshalJSON implements json.Marshaler interface for *Value.
// This allows jsonvalue.Value to be serialized directly using encoding/json.Marshal().
func (v *Value) MarshalJSON() ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	switch v.kind {
	case KindUndefined, KindNull:
		return []byte("null"), nil
	case KindBoolean:
		if v.bool {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	case KindInt64:
		// Convert to interface{} and let encoding/json handle the formatting
		return json.Marshal(v.i64)
	case KindNumber:
		return json.Marshal(v.num)
	case KindString:
		return json.Marshal(v.str)
	case KindArray:
		// Recursively marshal array elements
		return json.Marshal(v.arrElems)
	case KindObject:
		// Build a map preserving insertion order isn't directly supported by encoding/json,
		// but we can marshal a map. The order will be alphabetical in the output.
		// For formatted output with order preservation, we'd need custom serialization.
		// For now, use the map directly.
		return json.Marshal(v.objEntries)
	default:
		return []byte("null"), nil
	}
}

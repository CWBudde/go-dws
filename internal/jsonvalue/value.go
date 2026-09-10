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
	objEntries map[string]*Value
	owner      *Value
	str        string
	objKeys    []string
	arrElems   []*Value
	num        float64
	i64        int64
	kind       Kind
	bool       bool
}

// Owner returns the container (object or array) this value is currently stored
// in, or nil when the value is a root. It mirrors DWScript's TdwsJSONValue.Owner.
func (v *Value) Owner() *Value {
	if v == nil {
		return nil
	}
	return v.owner
}

// Detach removes the value from its owning container, mirroring DWScript's
// reparenting semantics: a node can only ever live in one place, so inserting it
// somewhere else first removes it from where it was. Detaching from an object
// drops the key; detaching from an array removes the slot (shrinking the array).
// Detaching a root value is a no-op.
func (v *Value) Detach() {
	if v == nil || v.owner == nil {
		return
	}
	owner := v.owner
	v.owner = nil
	switch owner.kind {
	case KindObject:
		for _, k := range owner.objKeys {
			if owner.objEntries[k] == v {
				owner.removeKey(k)
				return
			}
		}
	case KindArray:
		for i, elem := range owner.arrElems {
			if elem == v {
				owner.removeIndex(i)
				return
			}
		}
	}
}

// adopt makes child a member of v, detaching it from a previous owner first.
func (v *Value) adopt(child *Value) {
	if child == nil {
		return
	}
	child.Detach()
	child.owner = v
}

// disown clears the owner back-pointer of a value that is being evicted from v.
func disown(child *Value) {
	if child != nil {
		child.owner = nil
	}
}

// removeKey deletes an object entry without touching the child's owner pointer.
func (v *Value) removeKey(key string) {
	delete(v.objEntries, key)
	for i, k := range v.objKeys {
		if k == key {
			v.objKeys = append(v.objKeys[:i], v.objKeys[i+1:]...)
			break
		}
	}
}

// removeIndex deletes an array slot without touching the child's owner pointer.
func (v *Value) removeIndex(index int) {
	copy(v.arrElems[index:], v.arrElems[index+1:])
	v.arrElems[len(v.arrElems)-1] = nil
	v.arrElems = v.arrElems[:len(v.arrElems)-1]
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
	// Reparent first: the child may currently live in v itself, in which case
	// detaching drops the very key we are about to (re)create.
	v.adopt(child)
	if previous, exists := v.objEntries[key]; exists {
		if previous != child {
			disown(previous)
		}
	} else {
		v.objKeys = append(v.objKeys, key)
	}
	v.objEntries[key] = child
}

// ObjectDelete removes the entry if present. It returns true when a key was removed.
func (v *Value) ObjectDelete(key string) bool {
	if v == nil || v.kind != KindObject {
		return false
	}
	existing, exists := v.objEntries[key]
	if !exists {
		return false
	}
	disown(existing)
	v.removeKey(key)
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
	previous := v.arrElems[index]
	v.adopt(child)
	// adopt may have shrunk the array when child already lived in v.
	if index >= len(v.arrElems) {
		return false
	}
	if previous != child {
		disown(previous)
	}
	v.arrElems[index] = child
	return true
}

// ArraySwap exchanges two elements in place. Unlike ArraySet it does not
// reparent, so neither element is detached from the array.
func (v *Value) ArraySwap(i, j int) bool {
	if v == nil || v.kind != KindArray {
		return false
	}
	n := len(v.arrElems)
	if i < 0 || i >= n || j < 0 || j >= n {
		return false
	}
	v.arrElems[i], v.arrElems[j] = v.arrElems[j], v.arrElems[i]
	return true
}

// ArrayAppend appends an element to the end of the array.
func (v *Value) ArrayAppend(child *Value) {
	if v == nil || v.kind != KindArray {
		return
	}
	v.adopt(child)
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
	disown(v.arrElems[index])
	v.removeIndex(index)
	return true
}

// ClearArray removes all elements from the array.
func (v *Value) ClearArray() {
	if v == nil || v.kind != KindArray {
		return
	}
	for _, elem := range v.arrElems {
		disown(elem)
	}
	v.arrElems = v.arrElems[:0]
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

// IsFalsey reports whether the value is falsey per DWScript's JSON semantics:
// Undefined, Null, false, numeric zero, and the empty string are falsey; every
// object and array (even empty) and every non-empty string/non-zero number is
// truthy.
func (v *Value) IsFalsey() bool {
	if v == nil {
		return true
	}
	switch v.kind {
	case KindUndefined, KindNull:
		return true
	case KindBoolean:
		return !v.bool
	case KindInt64:
		return v.i64 == 0
	case KindNumber:
		return v.num == 0
	case KindString:
		return v.str == ""
	default:
		// Objects and arrays are always truthy.
		return false
	}
}

// Clone returns a deep copy of the value. Scalars are copied by value; objects
// and arrays are copied recursively, preserving key order.
func (v *Value) Clone() *Value {
	if v == nil {
		return nil
	}
	switch v.kind {
	case KindObject:
		obj := NewObject()
		for _, k := range v.objKeys {
			obj.ObjectSet(k, v.objEntries[k].Clone())
		}
		return obj
	case KindArray:
		arr := NewArray()
		for _, elem := range v.arrElems {
			arr.ArrayAppend(elem.Clone())
		}
		return arr
	default:
		copyVal := *v
		// A clone is always a fresh root; it is not a member of v's container.
		copyVal.owner = nil
		return &copyVal
	}
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

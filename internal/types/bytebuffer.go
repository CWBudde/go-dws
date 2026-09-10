package types

// ByteBufferType represents the DWScript built-in ByteBuffer type.
//
// A ByteBuffer is a mutable, resizable block of raw bytes with a read/write
// cursor (Position) and a set of little-endian typed accessors (Byte, Int8,
// Word, Int16, DWord, Int32, Int64, Single, Double, Extended) plus data-string,
// hex, base64 and JSON conversions.
//
// ByteBuffer has *reference* semantics: assigning one ByteBuffer variable to
// another makes both names refer to the same underlying storage. Unlike a class
// reference, however, a declared ByteBuffer variable is auto-instantiated, so
// `var b : ByteBuffer;` is immediately usable (it starts empty rather than nil).
// `Assign` is the operation that copies content between two buffers.
//
// Keeping this distinct from a ClassType is what allows the auto-instantiated,
// non-nillable behaviour the DWScript FunctionsByteBuffer corpus relies on.
type ByteBufferType struct{}

// String returns the DWScript spelling of the type.
func (t *ByteBufferType) String() string { return "ByteBuffer" }

// TypeKind returns the canonical kind tag for a ByteBuffer type.
func (t *ByteBufferType) TypeKind() string { return "BYTE_BUFFER" }

// Equals reports whether other is also the ByteBuffer type.
func (t *ByteBufferType) Equals(other Type) bool {
	if other == nil {
		return false
	}
	other = GetUnderlyingType(other)
	_, ok := other.(*ByteBufferType)
	return ok
}

// BYTE_BUFFER is the singleton ByteBuffer type instance.
var BYTE_BUFFER = &ByteBufferType{}

// IsByteBuffer reports whether t resolves to the built-in ByteBuffer type.
func IsByteBuffer(t Type) bool {
	if t == nil {
		return false
	}
	underlying := GetUnderlyingType(t)
	if underlying == nil {
		return false
	}
	return underlying.TypeKind() == "BYTE_BUFFER"
}

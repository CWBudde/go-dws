package runtime

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"unicode/utf16"
)

// ByteBufferError is a domain error raised by a ByteBuffer operation.
//
// The evaluator converts it into a catchable script Exception, appending the
// `[line: L, column: C]` suffix DWScript puts on runtime diagnostics. The
// message text itself is fixed by the DWScript corpus, so it is produced here
// rather than at the call site.
type ByteBufferError struct {
	// Message is the DWScript-compatible diagnostic without position suffix.
	Message string
}

// Error implements the error interface.
func (e *ByteBufferError) Error() string { return e.Message }

// newRangeError builds DWScript's "Out of range" diagnostic for an access of
// size bytes at index within a buffer of the given length.
func newRangeError(index, size, length int) *ByteBufferError {
	return &ByteBufferError{Message: fmt.Sprintf("Out of range (index %d, size %d for length %d)", index, size, length)}
}

// newOverflowError builds DWScript's "value out of <Type> range" diagnostic.
func newOverflowError(value int64, typeName string) *ByteBufferError {
	return &ByteBufferError{Message: fmt.Sprintf("value %d out of %s range", value, typeName)}
}

// MaxByteBufferLength is the largest buffer a script may request. The reference
// implementation has no explicit ceiling; it lets Delphi's SetLength raise
// EOutOfMemory, which surfaces as a catchable script exception. Go's make()
// panics instead and would take the host down, so a resize beyond this bound is
// turned into a ByteBuffer error. 2 GiB is far above any legitimate script use
// while keeping `b.SetLength(High(Integer))` a diagnosable error.
const MaxByteBufferLength = math.MaxInt32

// NewByteBufferLengthError builds the diagnostic for an unsatisfiable resize.
func NewByteBufferLengthError(length int64) *ByteBufferError {
	return &ByteBufferError{Message: fmt.Sprintf("Invalid length %d (maximum is %d)", length, MaxByteBufferLength)}
}

// ByteBufferValue is the runtime representation of a DWScript ByteBuffer.
//
// It is always handled through a pointer, which is what gives ByteBuffer its
// reference semantics: `b2 := b1` aliases the storage, while `b2.Assign(b1)`
// copies it.
type ByteBufferValue struct {
	data []byte
	pos  int
}

// Compile-time interface satisfaction check.
var _ Value = (*ByteBufferValue)(nil)

// NewByteBufferValue returns an empty ByteBuffer.
func NewByteBufferValue() *ByteBufferValue {
	return &ByteBufferValue{data: []byte{}}
}

// NewByteBufferValueFromBytes returns a ByteBuffer owning a copy of data.
func NewByteBufferValueFromBytes(data []byte) *ByteBufferValue {
	buf := make([]byte, len(data))
	copy(buf, data)
	return &ByteBufferValue{data: buf}
}

// Type returns "BYTEBUFFER".
func (b *ByteBufferValue) Type() string { return "BYTEBUFFER" }

// String returns the buffer rendered as a JSON byte array, matching ToJSON.
func (b *ByteBufferValue) String() string { return b.ToJSON() }

// ValueKind identifies the ByteBuffer runtime representation.
func (b *ByteBufferValue) ValueKind() ValueKind { return KindByteBuffer }

// Bytes returns the buffer's live backing storage. Callers must not retain it.
func (b *ByteBufferValue) Bytes() []byte { return b.data }

// Length returns the number of bytes held by the buffer.
func (b *ByteBufferValue) Length() int { return len(b.data) }

// Position returns the read/write cursor.
func (b *ByteBufferValue) Position() int { return b.pos }

// SetLength resizes the buffer. Growing appends zero bytes; shrinking discards
// the tail. The freshly exposed area is always zeroed, even when the buffer had
// previously been longer. A negative length is treated as zero.
func (b *ByteBufferValue) SetLength(n int) {
	if n < 0 {
		n = 0
	}
	resized := make([]byte, n)
	copy(resized, b.data)
	b.data = resized
	if b.pos > n {
		b.pos = n
	}
}

// SetPosition moves the cursor. Valid positions run from 0 to Length inclusive,
// so that a cursor left just past the final byte by a sequential read is legal.
func (b *ByteBufferValue) SetPosition(n int) error {
	if n < 0 || n > len(b.data) {
		return &ByteBufferError{Message: fmt.Sprintf("Position %d out of range (length %d)", n, len(b.data))}
	}
	b.pos = n
	return nil
}

// checkRange validates that size bytes are addressable at index.
//
// index and size both originate from script Integers, so the comparison must
// never form index+size: for an index near MaxInt64 that sum wraps negative and
// would let an out-of-range access through to a slice index panic. Comparing
// size against the remaining tail keeps every operand within range.
func (b *ByteBufferValue) checkRange(index, size int) error {
	if index < 0 || size < 0 || index > len(b.data) || size > len(b.data)-index {
		return newRangeError(index, size, len(b.data))
	}
	return nil
}

// ============================================================================
// Integer accessors
// ============================================================================

// readUint reads size bytes little-endian at index and returns them as a uint64.
func (b *ByteBufferValue) readUint(index, size int) (uint64, error) {
	if err := b.checkRange(index, size); err != nil {
		return 0, err
	}
	var raw uint64
	for i := size - 1; i >= 0; i-- {
		raw = raw<<8 | uint64(b.data[index+i])
	}
	return raw, nil
}

// writeUint stores the low size bytes of raw little-endian at index.
func (b *ByteBufferValue) writeUint(index, size int, raw uint64) error {
	if err := b.checkRange(index, size); err != nil {
		return err
	}
	for i := 0; i < size; i++ {
		b.data[index+i] = byte(raw >> (8 * uint(i)))
	}
	return nil
}

// signExtend interprets the low size bytes of raw as a signed integer.
func signExtend(raw uint64, size int) int64 {
	if size >= 8 {
		return int64(raw)
	}
	shift := uint(64 - 8*size)
	return int64(raw<<shift) >> shift
}

// GetIntAt reads a size-byte little-endian integer at index, interpreting it as
// signed or unsigned. It does not move the cursor.
func (b *ByteBufferValue) GetIntAt(index, size int, signed bool) (int64, error) {
	raw, err := b.readUint(index, size)
	if err != nil {
		return 0, err
	}
	if signed {
		return signExtend(raw, size), nil
	}
	return int64(raw), nil
}

// GetInt reads a size-byte little-endian integer at the cursor and advances it.
func (b *ByteBufferValue) GetInt(size int, signed bool) (int64, error) {
	value, err := b.GetIntAt(b.pos, size, signed)
	if err != nil {
		return 0, err
	}
	b.pos += size
	return value, nil
}

// ByteBufferIntSpec describes the width, signedness, diagnostic name and
// accepted value range of one of ByteBuffer's named integer accessors.
type ByteBufferIntSpec struct {
	// Name is the DWScript type name used in overflow diagnostics.
	Name string
	// Size is the accessor's width in bytes.
	Size int
	// Signed reports whether reads sign-extend.
	Signed bool
	// Low is the smallest value the accessor accepts.
	Low int64
	// High is the largest value the accessor accepts.
	High int64
}

// byteBufferIntegerSpecs maps the accessor suffix (byte, int8, word, int16,
// dword, int32, int64) to its specification.
var byteBufferIntegerSpecs = map[string]ByteBufferIntSpec{
	"byte":  {Name: "Byte", Size: 1, Signed: false, Low: 0, High: 255},
	"int8":  {Name: "Int8", Size: 1, Signed: true, Low: -128, High: 127},
	"word":  {Name: "Word", Size: 2, Signed: false, Low: 0, High: 65535},
	"int16": {Name: "Int16", Size: 2, Signed: true, Low: -32768, High: 32767},
	"dword": {Name: "DWord", Size: 4, Signed: false, Low: 0, High: 4294967295},
	"int32": {Name: "Int32", Size: 4, Signed: true, Low: -2147483648, High: 2147483647},
	"int64": {Name: "Int64", Size: 8, Signed: true, Low: math.MinInt64, High: math.MaxInt64},
}

// ByteBufferIntegerSpec reports the accessor specification for a normalized
// accessor suffix, and whether the suffix names an integer accessor at all.
func ByteBufferIntegerSpec(suffix string) (ByteBufferIntSpec, bool) {
	spec, ok := byteBufferIntegerSpecs[suffix]
	return spec, ok
}

// checkIntegerRange validates value against the accessor's accepted range.
// The overflow check precedes the buffer range check, matching DWScript.
func (spec ByteBufferIntSpec) checkIntegerRange(value int64) error {
	if spec.Size == 8 {
		return nil
	}
	if value < spec.Low || value > spec.High {
		return newOverflowError(value, spec.Name)
	}
	return nil
}

// SetIntAt writes value using the named accessor at index, without moving the
// cursor.
func (b *ByteBufferValue) SetIntAt(suffix string, index int, value int64) error {
	spec, ok := ByteBufferIntegerSpec(suffix)
	if !ok {
		return &ByteBufferError{Message: "unknown ByteBuffer integer accessor " + suffix}
	}
	if err := spec.checkIntegerRange(value); err != nil {
		return err
	}
	return b.writeUint(index, spec.Size, uint64(value))
}

// SetInt writes value using the named accessor at the cursor and advances it.
// A failed write leaves the cursor untouched.
func (b *ByteBufferValue) SetInt(suffix string, value int64) error {
	spec, ok := ByteBufferIntegerSpec(suffix)
	if !ok {
		return &ByteBufferError{Message: "unknown ByteBuffer integer accessor " + suffix}
	}
	if err := b.SetIntAt(suffix, b.pos, value); err != nil {
		return err
	}
	b.pos += spec.Size
	return nil
}

// GetIntegers reads count little-endian integers of elementSize bytes and
// returns them. It does not move the cursor.
//
// The index argument takes part in the bounds check but does not shift the read
// origin: the elements are always taken from the start of the buffer. That is
// what the reference implementation does, as pinned by
// testdata/fixtures/FunctionsByteBuffer/integers, where GetIntegers(1, 15, 2, ...)
// and GetIntegers(2, 7, 4, ...) both yield elements beginning at byte 0.
// (reference/dwscript-original/ is empty in this checkout, so the corpus is the
// only available specification.)
func (b *ByteBufferValue) GetIntegers(index, count, elementSize int, signed bool) ([]int64, error) {
	if elementSize <= 0 || elementSize > 8 {
		return nil, &ByteBufferError{Message: fmt.Sprintf("Invalid element size %d", elementSize)}
	}
	if count < 0 {
		count = 0
	}
	// count*elementSize would overflow for a hostile count, so reject anything
	// that cannot possibly fit before forming the product.
	if count > len(b.data) {
		return nil, newRangeError(index, count, len(b.data))
	}
	if err := b.checkRange(index, count*elementSize); err != nil {
		return nil, err
	}
	result := make([]int64, count)
	for i := 0; i < count; i++ {
		value, err := b.GetIntAt(i*elementSize, elementSize, signed)
		if err != nil {
			return nil, err
		}
		result[i] = value
	}
	return result, nil
}

// ============================================================================
// Floating point accessors
// ============================================================================

// GetSingleAt reads a 4-byte IEEE-754 float at index.
func (b *ByteBufferValue) GetSingleAt(index int) (float64, error) {
	raw, err := b.readUint(index, 4)
	if err != nil {
		return 0, err
	}
	return float64(math.Float32frombits(uint32(raw))), nil
}

// SetSingleAt writes value as a 4-byte IEEE-754 float at index.
func (b *ByteBufferValue) SetSingleAt(index int, value float64) error {
	return b.writeUint(index, 4, uint64(math.Float32bits(float32(value))))
}

// GetDoubleAt reads an 8-byte IEEE-754 double at index.
func (b *ByteBufferValue) GetDoubleAt(index int) (float64, error) {
	raw, err := b.readUint(index, 8)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(raw), nil
}

// SetDoubleAt writes value as an 8-byte IEEE-754 double at index.
func (b *ByteBufferValue) SetDoubleAt(index int, value float64) error {
	return b.writeUint(index, 8, math.Float64bits(value))
}

// extendedBias is the exponent bias of the x87 80-bit extended format.
const extendedBias = 16383

// GetExtendedAt reads a 10-byte x87 extended float at index and widens it to a
// Go float64. Precision beyond 53 mantissa bits is lost, as it is in DWScript
// on platforms whose Float is a double.
func (b *ByteBufferValue) GetExtendedAt(index int) (float64, error) {
	if err := b.checkRange(index, 10); err != nil {
		return 0, err
	}
	var mantissa uint64
	for i := 7; i >= 0; i-- {
		mantissa = mantissa<<8 | uint64(b.data[index+i])
	}
	signExp := uint16(b.data[index+9])<<8 | uint16(b.data[index+8])
	negative := signExp&0x8000 != 0
	exponent := int(signExp & 0x7FFF)

	var value float64
	switch {
	case exponent == 0x7FFF:
		if mantissa<<1 == 0 {
			value = math.Inf(1)
		} else {
			value = math.NaN()
		}
	case exponent == 0 && mantissa == 0:
		value = 0
	default:
		value = math.Ldexp(float64(mantissa), exponent-extendedBias-63)
	}
	if negative && !math.IsNaN(value) {
		value = -value
	}
	return value, nil
}

// SetExtendedAt writes value as a 10-byte x87 extended float at index.
func (b *ByteBufferValue) SetExtendedAt(index int, value float64) error {
	if err := b.checkRange(index, 10); err != nil {
		return err
	}
	var mantissa uint64
	var signExp uint16
	if math.Signbit(value) {
		signExp = 0x8000
	}
	abs := math.Abs(value)
	switch {
	case math.IsNaN(value):
		signExp = 0x7FFF
		mantissa = 0xC000000000000000
	case math.IsInf(value, 0):
		signExp |= 0x7FFF
		mantissa = 0x8000000000000000
	case abs == 0:
		mantissa = 0
	default:
		frac, exp := math.Frexp(abs)
		mantissa = uint64(math.Ldexp(frac, 64))
		signExp |= uint16(exp+extendedBias-1) & 0x7FFF
	}
	for i := 0; i < 8; i++ {
		b.data[index+i] = byte(mantissa >> (8 * uint(i)))
	}
	b.data[index+8] = byte(signExp)
	b.data[index+9] = byte(signExp >> 8)
	return nil
}

// byteBufferFloatSizes maps a float accessor suffix to its byte width.
var byteBufferFloatSizes = map[string]int{"single": 4, "double": 8, "extended": 10}

// byteBufferFloatSize reports the width of a float accessor suffix.
func byteBufferFloatSize(suffix string) (int, bool) {
	size, ok := byteBufferFloatSizes[suffix]
	return size, ok
}

// ByteBufferIsFloatAccessor reports whether a normalized accessor suffix names
// one of the floating point accessors (single, double, extended).
func ByteBufferIsFloatAccessor(suffix string) bool {
	_, ok := byteBufferFloatSizes[suffix]
	return ok
}

// GetFloatAt reads the named float accessor at index without moving the cursor.
func (b *ByteBufferValue) GetFloatAt(suffix string, index int) (float64, error) {
	switch suffix {
	case "single":
		return b.GetSingleAt(index)
	case "double":
		return b.GetDoubleAt(index)
	case "extended":
		return b.GetExtendedAt(index)
	}
	return 0, &ByteBufferError{Message: "unknown ByteBuffer float accessor " + suffix}
}

// GetFloat reads the named float accessor at the cursor and advances it.
func (b *ByteBufferValue) GetFloat(suffix string) (float64, error) {
	size, ok := byteBufferFloatSize(suffix)
	if !ok {
		return 0, &ByteBufferError{Message: "unknown ByteBuffer float accessor " + suffix}
	}
	value, err := b.GetFloatAt(suffix, b.pos)
	if err != nil {
		return 0, err
	}
	b.pos += size
	return value, nil
}

// SetFloatAt writes the named float accessor at index without moving the cursor.
func (b *ByteBufferValue) SetFloatAt(suffix string, index int, value float64) error {
	switch suffix {
	case "single":
		return b.SetSingleAt(index, value)
	case "double":
		return b.SetDoubleAt(index, value)
	case "extended":
		return b.SetExtendedAt(index, value)
	}
	return &ByteBufferError{Message: "unknown ByteBuffer float accessor " + suffix}
}

// SetFloat writes the named float accessor at the cursor and advances it.
// A failed write leaves the cursor untouched.
func (b *ByteBufferValue) SetFloat(suffix string, value float64) error {
	size, ok := byteBufferFloatSize(suffix)
	if !ok {
		return &ByteBufferError{Message: "unknown ByteBuffer float accessor " + suffix}
	}
	if err := b.SetFloatAt(suffix, b.pos, value); err != nil {
		return err
	}
	b.pos += size
	return nil
}

// ============================================================================
// Data strings, hex, base64, JSON
// ============================================================================

// DataStringToBytes converts a DWScript "data string" to raw bytes by taking the
// low byte of every UTF-16 code unit. This is the conversion DWScript's
// ByteBuffer(String) cast and AssignDataString use.
func DataStringToBytes(s string) []byte {
	units := utf16.Encode([]rune(s))
	out := make([]byte, len(units))
	for i, u := range units {
		out[i] = byte(u)
	}
	return out
}

// BytesToDataString converts raw bytes to a DWScript "data string", mapping each
// byte to the character with that code point.
func BytesToDataString(data []byte) string {
	var sb strings.Builder
	sb.Grow(len(data))
	for _, c := range data {
		sb.WriteRune(rune(c))
	}
	return sb.String()
}

// GetDataAt returns size bytes at index as a data string.
func (b *ByteBufferValue) GetDataAt(index, size int) (string, error) {
	if err := b.checkRange(index, size); err != nil {
		return "", err
	}
	return BytesToDataString(b.data[index : index+size]), nil
}

// SetDataAt writes a data string at index without moving the cursor.
func (b *ByteBufferValue) SetDataAt(index int, s string) error {
	raw := DataStringToBytes(s)
	if err := b.checkRange(index, len(raw)); err != nil {
		return err
	}
	copy(b.data[index:], raw)
	return nil
}

// SetData writes a data string at the cursor and advances it. A failed write
// leaves the cursor untouched.
func (b *ByteBufferValue) SetData(s string) error {
	raw := DataStringToBytes(s)
	if err := b.SetDataAt(b.pos, s); err != nil {
		return err
	}
	b.pos += len(raw)
	return nil
}

// ToDataString renders the whole buffer as a data string.
func (b *ByteBufferValue) ToDataString() string { return BytesToDataString(b.data) }

// ToHexString renders the whole buffer as lowercase hexadecimal.
func (b *ByteBufferValue) ToHexString() string { return hex.EncodeToString(b.data) }

// ToBase64 renders the whole buffer as standard padded base64.
func (b *ByteBufferValue) ToBase64() string { return base64.StdEncoding.EncodeToString(b.data) }

// ToJSON renders the whole buffer as a JSON array of byte values.
func (b *ByteBufferValue) ToJSON() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, c := range b.data {
		if i > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, "%d", c)
	}
	sb.WriteByte(']')
	return sb.String()
}

// replace installs data as the buffer's content and rewinds the cursor.
func (b *ByteBufferValue) replace(data []byte) {
	b.data = data
	b.pos = 0
}

// Assign copies the content of other into the buffer. The two buffers stay
// independent afterwards.
func (b *ByteBufferValue) Assign(other *ByteBufferValue) {
	if other == nil {
		b.replace([]byte{})
		return
	}
	clone := make([]byte, len(other.data))
	copy(clone, other.data)
	b.replace(clone)
}

// AssignDataString replaces the content with the bytes of a data string.
func (b *ByteBufferValue) AssignDataString(s string) {
	b.replace(DataStringToBytes(s))
}

// AssignHexString replaces the content with the bytes of a hex string.
func (b *ByteBufferValue) AssignHexString(s string) error {
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return &ByteBufferError{Message: "Invalid hexadecimal string"}
	}
	b.replace(decoded)
	return nil
}

// AssignBase64 replaces the content with the bytes of a base64 string.
func (b *ByteBufferValue) AssignBase64(s string) error {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return &ByteBufferError{Message: "Invalid base64 string"}
	}
	b.replace(decoded)
	return nil
}

// AssignJSON replaces the content with the bytes of a JSON array of numbers.
func (b *ByteBufferValue) AssignJSON(s string) error {
	var numbers []int64
	if err := json.Unmarshal([]byte(s), &numbers); err != nil {
		return &ByteBufferError{Message: "Invalid JSON byte array"}
	}
	decoded := make([]byte, len(numbers))
	for i, n := range numbers {
		if n < 0 || n > 255 {
			return newOverflowError(n, "Byte")
		}
		decoded[i] = byte(n)
	}
	b.replace(decoded)
	return nil
}

// Copy returns a new ByteBuffer holding count bytes starting at index. The range
// is clamped to the buffer, so an over-long count yields the available tail.
func (b *ByteBufferValue) Copy(index, count int) *ByteBufferValue {
	if index < 0 {
		index = 0
	}
	if index > len(b.data) {
		index = len(b.data)
	}
	if count < 0 {
		count = 0
	}
	// Clamp against the remaining tail rather than against index+count, which
	// wraps negative for a script-supplied count near MaxInt64.
	if count > len(b.data)-index {
		count = len(b.data) - index
	}
	return NewByteBufferValueFromBytes(b.data[index : index+count])
}

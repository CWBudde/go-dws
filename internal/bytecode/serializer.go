package bytecode

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Bytecode file format (.dwc) specification
// ==========================================
//
// Header (8 bytes):
//   - Magic number: "DWC\x00" (4 bytes)
//   - Version major: uint8 (1 byte)
//   - Version minor: uint8 (1 byte)
//   - Version patch: uint8 (1 byte)
//   - Reserved: uint8 (1 byte) - for future use
//
// Body:
//   - Chunk data (variable length, see SerializeChunk)
//
// Design goals:
//   - Forward compatibility: version checks allow graceful failures
//   - Compact: binary format, no padding
//   - Complete: captures all runtime state needed for execution

const (
	// MagicNumber identifies DWScript bytecode files
	MagicNumber = "DWC\x00"

	// Version of the bytecode format
	VersionMajor = 1
	VersionMinor = 0
	VersionPatch = 0
)

// SerializerVersion represents a bytecode format version
type SerializerVersion struct {
	Major uint8
	Minor uint8
	Patch uint8
}

// String returns a string representation of the version
func (v SerializerVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsCompatible checks if this version can read bytecode of another version
func (v SerializerVersion) IsCompatible(other SerializerVersion) bool {
	// Major version must match exactly
	if v.Major != other.Major {
		return false
	}
	// Can read older minor versions, but not newer ones
	if other.Minor > v.Minor {
		return false
	}
	return true
}

// CurrentVersion returns the current serializer version
func CurrentVersion() SerializerVersion {
	return SerializerVersion{
		Major: VersionMajor,
		Minor: VersionMinor,
		Patch: VersionPatch,
	}
}

// Serializer handles bytecode serialization/deserialization
type Serializer struct {
	version SerializerVersion
}

// NewSerializer creates a new serializer with the current version
func NewSerializer() *Serializer {
	return &Serializer{
		version: CurrentVersion(),
	}
}

// SerializeChunk writes a Chunk to binary format
// Format:
//   - Name: string (length-prefixed)
//   - LocalCount: int32
//   - Code: []Instruction (count + items)
//   - Constants: []Value (count + items)
//   - Lines: []LineInfo (count + items)
//   - TryInfos: map[int]TryInfo (count + key-value pairs)
//   - Helpers: map[string]*HelperInfo (count + key-value pairs)
func (s *Serializer) SerializeChunk(chunk *Chunk) ([]byte, error) {
	if chunk == nil {
		return nil, fmt.Errorf("cannot serialize nil chunk")
	}

	buf := new(bytes.Buffer)

	// Write header
	if err := s.writeHeader(buf); err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}

	// Write chunk name
	if err := s.writeString(buf, chunk.Name); err != nil {
		return nil, fmt.Errorf("failed to write chunk name: %w", err)
	}

	// Write local count
	if err := s.writeInt32(buf, int32(chunk.LocalCount)); err != nil {
		return nil, fmt.Errorf("failed to write local count: %w", err)
	}

	// Write instructions
	if err := s.writeInstructions(buf, chunk.Code); err != nil {
		return nil, fmt.Errorf("failed to write instructions: %w", err)
	}

	// Write constants
	if err := s.writeConstants(buf, chunk.Constants); err != nil {
		return nil, fmt.Errorf("failed to write constants: %w", err)
	}

	// Write line info
	if err := s.writeLineInfos(buf, chunk.Lines); err != nil {
		return nil, fmt.Errorf("failed to write line info: %w", err)
	}

	// Write try infos
	if err := s.writeTryInfos(buf, chunk.tryInfos); err != nil {
		return nil, fmt.Errorf("failed to write try infos: %w", err)
	}

	// Write helpers
	if err := s.writeHelpers(buf, chunk.Helpers); err != nil {
		return nil, fmt.Errorf("failed to write helpers: %w", err)
	}

	return buf.Bytes(), nil
}

// DeserializeChunk reads a Chunk from binary format
func (s *Serializer) DeserializeChunk(data []byte) (*Chunk, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("bytecode too short: expected at least 8 bytes, got %d", len(data))
	}

	buf := bytes.NewReader(data)

	// Read and validate header
	version, err := s.readHeader(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	if !s.version.IsCompatible(version) {
		return nil, fmt.Errorf("incompatible bytecode version: have %s, bytecode is %s", s.version, version)
	}

	// Read chunk name
	name, err := s.readString(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk name: %w", err)
	}

	chunk := NewChunk(name)

	// Read local count
	localCount, err := s.readInt32(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read local count: %w", err)
	}
	chunk.LocalCount = int(localCount)

	// Read instructions
	instructions, err := s.readInstructions(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read instructions: %w", err)
	}
	chunk.Code = instructions

	// Read constants
	constants, err := s.readConstants(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read constants: %w", err)
	}
	chunk.Constants = constants

	// Read line info
	lines, err := s.readLineInfos(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read line info: %w", err)
	}
	chunk.Lines = lines

	// Read try infos
	tryInfos, err := s.readTryInfos(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read try infos: %w", err)
	}
	chunk.tryInfos = tryInfos

	// Read helpers
	helpers, err := s.readHelpers(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read helpers: %w", err)
	}
	chunk.Helpers = helpers

	return chunk, nil
}

// ============================================================================
// Header serialization
// ============================================================================

func (s *Serializer) writeHeader(w io.Writer) error {
	// Write magic number
	if _, err := w.Write([]byte(MagicNumber)); err != nil {
		return err
	}
	// Write version
	if err := binary.Write(w, binary.LittleEndian, s.version.Major); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, s.version.Minor); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, s.version.Patch); err != nil {
		return err
	}
	// Write reserved byte
	return binary.Write(w, binary.LittleEndian, uint8(0))
}

func (s *Serializer) readHeader(r io.Reader) (SerializerVersion, error) {
	// Read magic number
	magic := make([]byte, 4)
	if _, err := io.ReadFull(r, magic); err != nil {
		return SerializerVersion{}, fmt.Errorf("failed to read magic number: %w", err)
	}
	if string(magic) != MagicNumber {
		return SerializerVersion{}, fmt.Errorf("invalid magic number: expected %q, got %q", MagicNumber, string(magic))
	}

	// Read version
	var version SerializerVersion
	if err := binary.Read(r, binary.LittleEndian, &version.Major); err != nil {
		return SerializerVersion{}, fmt.Errorf("failed to read major version: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &version.Minor); err != nil {
		return SerializerVersion{}, fmt.Errorf("failed to read minor version: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &version.Patch); err != nil {
		return SerializerVersion{}, fmt.Errorf("failed to read patch version: %w", err)
	}

	// Read reserved byte
	var reserved uint8
	if err := binary.Read(r, binary.LittleEndian, &reserved); err != nil {
		return SerializerVersion{}, fmt.Errorf("failed to read reserved byte: %w", err)
	}

	return version, nil
}

// ============================================================================
// Primitive type serialization
// ============================================================================

func (s *Serializer) writeString(w io.Writer, str string) error {
	// Write length
	if err := binary.Write(w, binary.LittleEndian, uint32(len(str))); err != nil {
		return err
	}
	// Write string data
	if len(str) > 0 {
		_, err := w.Write([]byte(str))
		return err
	}
	return nil
}

func (s *Serializer) readString(r io.Reader) (string, error) {
	// Read length
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	// Read string data
	if length > 0 {
		data := make([]byte, length)
		if _, err := io.ReadFull(r, data); err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", nil
}

func (s *Serializer) writeInt32(w io.Writer, val int32) error {
	return binary.Write(w, binary.LittleEndian, val)
}

func (s *Serializer) readInt32(r io.Reader) (int32, error) {
	var val int32
	err := binary.Read(r, binary.LittleEndian, &val)
	return val, err
}

func (s *Serializer) writeInt64(w io.Writer, val int64) error {
	return binary.Write(w, binary.LittleEndian, val)
}

func (s *Serializer) readInt64(r io.Reader) (int64, error) {
	var val int64
	err := binary.Read(r, binary.LittleEndian, &val)
	return val, err
}

func (s *Serializer) writeFloat64(w io.Writer, val float64) error {
	return binary.Write(w, binary.LittleEndian, val)
}

func (s *Serializer) readFloat64(r io.Reader) (float64, error) {
	var val float64
	err := binary.Read(r, binary.LittleEndian, &val)
	return val, err
}

func (s *Serializer) writeBool(w io.Writer, val bool) error {
	var b uint8
	if val {
		b = 1
	}
	return binary.Write(w, binary.LittleEndian, b)
}

func (s *Serializer) readBool(r io.Reader) (bool, error) {
	var b uint8
	if err := binary.Read(r, binary.LittleEndian, &b); err != nil {
		return false, err
	}
	return b != 0, nil
}

func (s *Serializer) writeUint16(w io.Writer, val uint16) error {
	return binary.Write(w, binary.LittleEndian, val)
}

func (s *Serializer) readUint16(r io.Reader) (uint16, error) {
	var val uint16
	err := binary.Read(r, binary.LittleEndian, &val)
	return val, err
}

// ============================================================================
// Instruction serialization
// ============================================================================

func (s *Serializer) writeInstructions(w io.Writer, instructions []Instruction) error {
	// Write count
	if err := binary.Write(w, binary.LittleEndian, uint32(len(instructions))); err != nil {
		return err
	}
	// Write instructions (each is uint32)
	for _, inst := range instructions {
		if err := binary.Write(w, binary.LittleEndian, uint32(inst)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Serializer) readInstructions(r io.Reader) ([]Instruction, error) {
	// Read count
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	// Read instructions
	instructions := make([]Instruction, count)
	for i := range instructions {
		var inst uint32
		if err := binary.Read(r, binary.LittleEndian, &inst); err != nil {
			return nil, err
		}
		instructions[i] = Instruction(inst)
	}
	return instructions, nil
}

// ============================================================================
// Value serialization
// ============================================================================

func (s *Serializer) writeValue(w io.Writer, val Value) error {
	// Write type tag
	if err := binary.Write(w, binary.LittleEndian, uint8(val.Type)); err != nil {
		return err
	}

	// Write value data based on type
	switch val.Type {
	case ValueNil:
		// No data for nil
		return nil

	case ValueBool:
		return s.writeBool(w, val.AsBool())

	case ValueInt:
		return s.writeInt64(w, val.AsInt())

	case ValueFloat:
		return s.writeFloat64(w, val.AsFloat())

	case ValueString:
		return s.writeString(w, val.AsString())

	case ValueFunction:
		// Serialize function object
		fn := val.Data.(*FunctionObject)
		if err := s.writeString(w, fn.Name); err != nil {
			return err
		}
		if err := s.writeInt32(w, int32(fn.Arity)); err != nil {
			return err
		}
		// Serialize the function's chunk
		chunkData, err := s.SerializeChunk(fn.Chunk)
		if err != nil {
			return err
		}
		// Write chunk size and data
		if err := s.writeInt32(w, int32(len(chunkData))); err != nil {
			return err
		}
		_, err = w.Write(chunkData)
		if err != nil {
			return err
		}
		// Serialize upvalue definitions
		if err := s.writeInt32(w, int32(len(fn.UpvalueDefs))); err != nil {
			return err
		}
		for _, upval := range fn.UpvalueDefs {
			if err := s.writeBool(w, upval.IsLocal); err != nil {
				return err
			}
			if err := s.writeInt32(w, int32(upval.Index)); err != nil {
				return err
			}
		}
		return nil

	case ValueBuiltin:
		// Serialize builtin name
		return s.writeString(w, val.Data.(string))

	case ValueArray, ValueObject, ValueClosure, ValueVariant:
		// These types cannot be serialized as constants
		// They are runtime-only values
		return fmt.Errorf("cannot serialize value type %s as constant", val.Type)

	default:
		return fmt.Errorf("unknown value type: %d", val.Type)
	}
}

func (s *Serializer) readValue(r io.Reader) (Value, error) {
	// Read type tag
	var typeTag uint8
	if err := binary.Read(r, binary.LittleEndian, &typeTag); err != nil {
		return Value{}, err
	}

	valueType := ValueType(typeTag)

	switch valueType {
	case ValueNil:
		return NilValue(), nil

	case ValueBool:
		b, err := s.readBool(r)
		if err != nil {
			return Value{}, err
		}
		return BoolValue(b), nil

	case ValueInt:
		i, err := s.readInt64(r)
		if err != nil {
			return Value{}, err
		}
		return IntValue(i), nil

	case ValueFloat:
		f, err := s.readFloat64(r)
		if err != nil {
			return Value{}, err
		}
		return FloatValue(f), nil

	case ValueString:
		str, err := s.readString(r)
		if err != nil {
			return Value{}, err
		}
		return StringValue(str), nil

	case ValueFunction:
		// Deserialize function object
		name, err := s.readString(r)
		if err != nil {
			return Value{}, err
		}
		arity, err := s.readInt32(r)
		if err != nil {
			return Value{}, err
		}
		// Read chunk size and data
		chunkSize, err := s.readInt32(r)
		if err != nil {
			return Value{}, err
		}
		chunkData := make([]byte, chunkSize)
		if _, err := io.ReadFull(r, chunkData); err != nil {
			return Value{}, err
		}
		chunk, err := s.DeserializeChunk(chunkData)
		if err != nil {
			return Value{}, err
		}
		// Read upvalue definitions
		upvalCount, err := s.readInt32(r)
		if err != nil {
			return Value{}, err
		}
		upvalueDefs := make([]UpvalueDef, upvalCount)
		for i := range upvalueDefs {
			isLocal, err := s.readBool(r)
			if err != nil {
				return Value{}, err
			}
			index, err := s.readInt32(r)
			if err != nil {
				return Value{}, err
			}
			upvalueDefs[i] = UpvalueDef{
				IsLocal: isLocal,
				Index:   int(index),
			}
		}
		fn := NewFunctionObject(name, chunk, int(arity))
		fn.UpvalueDefs = upvalueDefs
		return FunctionValue(fn), nil

	case ValueBuiltin:
		name, err := s.readString(r)
		if err != nil {
			return Value{}, err
		}
		return BuiltinValue(name), nil

	default:
		return Value{}, fmt.Errorf("unknown or unsupported value type: %d", typeTag)
	}
}

func (s *Serializer) writeConstants(w io.Writer, constants []Value) error {
	// Write count
	if err := binary.Write(w, binary.LittleEndian, uint32(len(constants))); err != nil {
		return err
	}
	// Write each constant
	for _, val := range constants {
		if err := s.writeValue(w, val); err != nil {
			return err
		}
	}
	return nil
}

func (s *Serializer) readConstants(r io.Reader) ([]Value, error) {
	// Read count
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	// Read each constant
	constants := make([]Value, count)
	for i := range constants {
		val, err := s.readValue(r)
		if err != nil {
			return nil, err
		}
		constants[i] = val
	}
	return constants, nil
}

// ============================================================================
// LineInfo serialization
// ============================================================================

func (s *Serializer) writeLineInfos(w io.Writer, lines []LineInfo) error {
	// Write count
	if err := binary.Write(w, binary.LittleEndian, uint32(len(lines))); err != nil {
		return err
	}
	// Write each line info
	for _, line := range lines {
		if err := s.writeInt32(w, int32(line.InstructionOffset)); err != nil {
			return err
		}
		if err := s.writeInt32(w, int32(line.Line)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Serializer) readLineInfos(r io.Reader) ([]LineInfo, error) {
	// Read count
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	// Read each line info
	lines := make([]LineInfo, count)
	for i := range lines {
		offset, err := s.readInt32(r)
		if err != nil {
			return nil, err
		}
		line, err := s.readInt32(r)
		if err != nil {
			return nil, err
		}
		lines[i] = LineInfo{
			InstructionOffset: int(offset),
			Line:              int(line),
		}
	}
	return lines, nil
}

// ============================================================================
// TryInfo serialization
// ============================================================================

func (s *Serializer) writeTryInfos(w io.Writer, tryInfos map[int]TryInfo) error {
	// Write count
	if err := binary.Write(w, binary.LittleEndian, uint32(len(tryInfos))); err != nil {
		return err
	}
	// Write each try info
	for offset, info := range tryInfos {
		if err := s.writeInt32(w, int32(offset)); err != nil {
			return err
		}
		if err := s.writeInt32(w, int32(info.CatchTarget)); err != nil {
			return err
		}
		if err := s.writeInt32(w, int32(info.FinallyTarget)); err != nil {
			return err
		}
		if err := s.writeBool(w, info.HasCatch); err != nil {
			return err
		}
		if err := s.writeBool(w, info.HasFinally); err != nil {
			return err
		}
	}
	return nil
}

func (s *Serializer) readTryInfos(r io.Reader) (map[int]TryInfo, error) {
	// Read count
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	// Read each try info
	tryInfos := make(map[int]TryInfo, count)
	for i := uint32(0); i < count; i++ {
		offset, err := s.readInt32(r)
		if err != nil {
			return nil, err
		}
		catchTarget, err := s.readInt32(r)
		if err != nil {
			return nil, err
		}
		finallyTarget, err := s.readInt32(r)
		if err != nil {
			return nil, err
		}
		hasCatch, err := s.readBool(r)
		if err != nil {
			return nil, err
		}
		hasFinally, err := s.readBool(r)
		if err != nil {
			return nil, err
		}
		tryInfos[int(offset)] = TryInfo{
			CatchTarget:   int(catchTarget),
			FinallyTarget: int(finallyTarget),
			HasCatch:      hasCatch,
			HasFinally:    hasFinally,
		}
	}
	return tryInfos, nil
}

// ============================================================================
// HelperInfo serialization
// ============================================================================

func (s *Serializer) writeHelpers(w io.Writer, helpers map[string]*HelperInfo) error {
	// Write count
	if err := binary.Write(w, binary.LittleEndian, uint32(len(helpers))); err != nil {
		return err
	}
	// Write each helper
	for name, helper := range helpers {
		if err := s.writeString(w, name); err != nil {
			return err
		}
		if err := s.writeHelperInfo(w, helper); err != nil {
			return err
		}
	}
	return nil
}

func (s *Serializer) readHelpers(r io.Reader) (map[string]*HelperInfo, error) {
	// Read count
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	// Read each helper
	helpers := make(map[string]*HelperInfo, count)
	for i := uint32(0); i < count; i++ {
		name, err := s.readString(r)
		if err != nil {
			return nil, err
		}
		helper, err := s.readHelperInfo(r)
		if err != nil {
			return nil, err
		}
		helpers[name] = helper
	}
	return helpers, nil
}

func (s *Serializer) writeHelperInfo(w io.Writer, helper *HelperInfo) error {
	if helper == nil {
		return fmt.Errorf("cannot serialize nil helper info")
	}

	if err := s.writeString(w, helper.Name); err != nil {
		return err
	}
	if err := s.writeString(w, helper.TargetType); err != nil {
		return err
	}
	if err := s.writeString(w, helper.ParentHelper); err != nil {
		return err
	}

	// Write methods map
	if err := binary.Write(w, binary.LittleEndian, uint32(len(helper.Methods))); err != nil {
		return err
	}
	for name, slot := range helper.Methods {
		if err := s.writeString(w, name); err != nil {
			return err
		}
		if err := s.writeUint16(w, slot); err != nil {
			return err
		}
	}

	// Write properties
	if err := binary.Write(w, binary.LittleEndian, uint32(len(helper.Properties))); err != nil {
		return err
	}
	for _, prop := range helper.Properties {
		if err := s.writeString(w, prop); err != nil {
			return err
		}
	}

	// Write class vars
	if err := binary.Write(w, binary.LittleEndian, uint32(len(helper.ClassVars))); err != nil {
		return err
	}
	for _, classVar := range helper.ClassVars {
		if err := s.writeString(w, classVar); err != nil {
			return err
		}
	}

	// Write class consts
	if err := binary.Write(w, binary.LittleEndian, uint32(len(helper.ClassConsts))); err != nil {
		return err
	}
	for name, val := range helper.ClassConsts {
		if err := s.writeString(w, name); err != nil {
			return err
		}
		if err := s.writeValue(w, val); err != nil {
			return err
		}
	}

	return nil
}

func (s *Serializer) readHelperInfo(r io.Reader) (*HelperInfo, error) {
	helper := &HelperInfo{
		Methods:     make(map[string]uint16),
		ClassConsts: make(map[string]Value),
	}

	var err error
	helper.Name, err = s.readString(r)
	if err != nil {
		return nil, err
	}
	helper.TargetType, err = s.readString(r)
	if err != nil {
		return nil, err
	}
	helper.ParentHelper, err = s.readString(r)
	if err != nil {
		return nil, err
	}

	// Read methods map
	var methodCount uint32
	if err := binary.Read(r, binary.LittleEndian, &methodCount); err != nil {
		return nil, err
	}
	for i := uint32(0); i < methodCount; i++ {
		name, err := s.readString(r)
		if err != nil {
			return nil, err
		}
		slot, err := s.readUint16(r)
		if err != nil {
			return nil, err
		}
		helper.Methods[name] = slot
	}

	// Read properties
	var propCount uint32
	if err := binary.Read(r, binary.LittleEndian, &propCount); err != nil {
		return nil, err
	}
	helper.Properties = make([]string, propCount)
	for i := uint32(0); i < propCount; i++ {
		helper.Properties[i], err = s.readString(r)
		if err != nil {
			return nil, err
		}
	}

	// Read class vars
	var classVarCount uint32
	if err := binary.Read(r, binary.LittleEndian, &classVarCount); err != nil {
		return nil, err
	}
	helper.ClassVars = make([]string, classVarCount)
	for i := uint32(0); i < classVarCount; i++ {
		helper.ClassVars[i], err = s.readString(r)
		if err != nil {
			return nil, err
		}
	}

	// Read class consts
	var classConstCount uint32
	if err := binary.Read(r, binary.LittleEndian, &classConstCount); err != nil {
		return nil, err
	}
	for i := uint32(0); i < classConstCount; i++ {
		name, err := s.readString(r)
		if err != nil {
			return nil, err
		}
		val, err := s.readValue(r)
		if err != nil {
			return nil, err
		}
		helper.ClassConsts[name] = val
	}

	return helper, nil
}

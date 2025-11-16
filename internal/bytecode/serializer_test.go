package bytecode

import (
	"bytes"
	"testing"
)

func TestSerializer_SimpleProgram(t *testing.T) {
	// Create a simple chunk with basic operations
	chunk := NewChunk("test")
	chunk.LocalCount = 2

	// var x := 42; x + 10
	chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1) // Load 42
	chunk.WriteInstruction(MakeInstruction(OpStoreLocal, 0, 0), 1)
	chunk.WriteInstruction(MakeInstruction(OpLoadLocal, 0, 0), 1)
	chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 1), 1) // Load 10
	chunk.WriteInstruction(MakeSimpleInstruction(OpAddInt), 1)
	chunk.WriteInstruction(MakeSimpleInstruction(OpReturn), 1)
	chunk.Constants = []Value{IntValue(42), IntValue(10)}

	// Serialize
	serializer := NewSerializer()
	data, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	// Deserialize
	deserialized, err := serializer.DeserializeChunk(data)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	// Verify
	if deserialized.Name != chunk.Name {
		t.Errorf("Name mismatch: expected %q, got %q", chunk.Name, deserialized.Name)
	}
	if deserialized.LocalCount != chunk.LocalCount {
		t.Errorf("LocalCount mismatch: expected %d, got %d", chunk.LocalCount, deserialized.LocalCount)
	}
	if len(deserialized.Code) != len(chunk.Code) {
		t.Fatalf("Code length mismatch: expected %d, got %d", len(chunk.Code), len(deserialized.Code))
	}
	for i := range chunk.Code {
		if deserialized.Code[i] != chunk.Code[i] {
			t.Errorf("Code[%d] mismatch: expected %d, got %d", i, chunk.Code[i], deserialized.Code[i])
		}
	}
	if len(deserialized.Constants) != len(chunk.Constants) {
		t.Fatalf("Constants length mismatch: expected %d, got %d", len(chunk.Constants), len(deserialized.Constants))
	}
	for i := range chunk.Constants {
		if deserialized.Constants[i].Type != chunk.Constants[i].Type {
			t.Errorf("Constants[%d] type mismatch: expected %v, got %v", i, chunk.Constants[i].Type, deserialized.Constants[i].Type)
		}
		if deserialized.Constants[i].AsInt() != chunk.Constants[i].AsInt() {
			t.Errorf("Constants[%d] value mismatch: expected %d, got %d", i, chunk.Constants[i].AsInt(), deserialized.Constants[i].AsInt())
		}
	}
}

func TestSerializer_AllValueTypes(t *testing.T) {
	tests := []struct {
		name  string
		value Value
	}{
		{"nil", NilValue()},
		{"bool_true", BoolValue(true)},
		{"bool_false", BoolValue(false)},
		{"int_positive", IntValue(42)},
		{"int_negative", IntValue(-42)},
		{"int_zero", IntValue(0)},
		{"float_positive", FloatValue(3.14)},
		{"float_negative", FloatValue(-3.14)},
		{"float_zero", FloatValue(0.0)},
		{"string_simple", StringValue("hello")},
		{"string_empty", StringValue("")},
		{"string_unicode", StringValue("hello 世界 🌍")},
		{"builtin", BuiltinValue("Print")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := NewChunk("test_" + tt.name)
			chunk.Constants = []Value{tt.value}

			serializer := NewSerializer()
			data, err := serializer.SerializeChunk(chunk)
			if err != nil {
				t.Fatalf("SerializeChunk failed: %v", err)
			}

			deserialized, err := serializer.DeserializeChunk(data)
			if err != nil {
				t.Fatalf("DeserializeChunk failed: %v", err)
			}

			if len(deserialized.Constants) != 1 {
				t.Fatalf("Expected 1 constant, got %d", len(deserialized.Constants))
			}

			got := deserialized.Constants[0]
			if got.Type != tt.value.Type {
				t.Errorf("Type mismatch: expected %v, got %v", tt.value.Type, got.Type)
			}

			// Compare values based on type
			switch tt.value.Type {
			case ValueNil:
				// No value to compare
			case ValueBool:
				if got.AsBool() != tt.value.AsBool() {
					t.Errorf("Bool mismatch: expected %v, got %v", tt.value.AsBool(), got.AsBool())
				}
			case ValueInt:
				if got.AsInt() != tt.value.AsInt() {
					t.Errorf("Int mismatch: expected %d, got %d", tt.value.AsInt(), got.AsInt())
				}
			case ValueFloat:
				if got.AsFloat() != tt.value.AsFloat() {
					t.Errorf("Float mismatch: expected %f, got %f", tt.value.AsFloat(), got.AsFloat())
				}
			case ValueString:
				if got.AsString() != tt.value.AsString() {
					t.Errorf("String mismatch: expected %q, got %q", tt.value.AsString(), got.AsString())
				}
			case ValueBuiltin:
				if got.Data.(string) != tt.value.Data.(string) {
					t.Errorf("Builtin mismatch: expected %q, got %q", tt.value.Data.(string), got.Data.(string))
				}
			}
		})
	}
}

func TestSerializer_FunctionValue(t *testing.T) {
	// Create a function with its own chunk
	funcChunk := NewChunk("testFunc")
	funcChunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
	funcChunk.WriteInstruction(MakeSimpleInstruction(OpReturn), 1)
	funcChunk.Constants = []Value{IntValue(99)}
	funcChunk.LocalCount = 1

	fn := NewFunctionObject("testFunc", funcChunk, 0)
	fn.UpvalueDefs = []UpvalueDef{
		{IsLocal: true, Index: 0},
		{IsLocal: false, Index: 1},
	}

	// Create a chunk containing the function as a constant
	mainChunk := NewChunk("main")
	mainChunk.Constants = []Value{FunctionValue(fn)}

	// Serialize and deserialize
	serializer := NewSerializer()
	data, err := serializer.SerializeChunk(mainChunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	deserialized, err := serializer.DeserializeChunk(data)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	// Verify
	if len(deserialized.Constants) != 1 {
		t.Fatalf("Expected 1 constant, got %d", len(deserialized.Constants))
	}

	if deserialized.Constants[0].Type != ValueFunction {
		t.Fatalf("Expected function value, got %v", deserialized.Constants[0].Type)
	}

	deserializedFn := deserialized.Constants[0].Data.(*FunctionObject)
	if deserializedFn.Name != fn.Name {
		t.Errorf("Function name mismatch: expected %q, got %q", fn.Name, deserializedFn.Name)
	}
	if deserializedFn.Arity != fn.Arity {
		t.Errorf("Function arity mismatch: expected %d, got %d", fn.Arity, deserializedFn.Arity)
	}
	if deserializedFn.Chunk.LocalCount != fn.Chunk.LocalCount {
		t.Errorf("Function local count mismatch: expected %d, got %d", fn.Chunk.LocalCount, deserializedFn.Chunk.LocalCount)
	}
	if len(deserializedFn.UpvalueDefs) != len(fn.UpvalueDefs) {
		t.Fatalf("Upvalue defs length mismatch: expected %d, got %d", len(fn.UpvalueDefs), len(deserializedFn.UpvalueDefs))
	}
	for i := range fn.UpvalueDefs {
		if deserializedFn.UpvalueDefs[i] != fn.UpvalueDefs[i] {
			t.Errorf("UpvalueDef[%d] mismatch: expected %+v, got %+v", i, fn.UpvalueDefs[i], deserializedFn.UpvalueDefs[i])
		}
	}
}

func TestSerializer_LineInfo(t *testing.T) {
	chunk := NewChunk("test")
	chunk.WriteInstruction(MakeSimpleInstruction(OpLoadNil), 1)
	chunk.WriteInstruction(MakeSimpleInstruction(OpLoadNil), 2)
	chunk.WriteInstruction(MakeSimpleInstruction(OpLoadNil), 5)

	serializer := NewSerializer()
	data, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	deserialized, err := serializer.DeserializeChunk(data)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	if len(deserialized.Lines) != len(chunk.Lines) {
		t.Fatalf("Lines length mismatch: expected %d, got %d", len(chunk.Lines), len(deserialized.Lines))
	}

	for i := range chunk.Lines {
		if deserialized.Lines[i] != chunk.Lines[i] {
			t.Errorf("Lines[%d] mismatch: expected %+v, got %+v", i, chunk.Lines[i], deserialized.Lines[i])
		}
	}
}

func TestSerializer_TryInfo(t *testing.T) {
	chunk := NewChunk("test")
	chunk.SetTryInfo(5, TryInfo{
		CatchTarget:   10,
		FinallyTarget: 20,
		HasCatch:      true,
		HasFinally:    true,
	})
	chunk.SetTryInfo(15, TryInfo{
		CatchTarget: 25,
		HasCatch:    true,
		HasFinally:  false,
	})

	serializer := NewSerializer()
	data, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	deserialized, err := serializer.DeserializeChunk(data)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	if len(deserialized.tryInfos) != len(chunk.tryInfos) {
		t.Fatalf("TryInfos length mismatch: expected %d, got %d", len(chunk.tryInfos), len(deserialized.tryInfos))
	}

	for offset, info := range chunk.tryInfos {
		deserializedInfo, ok := deserialized.tryInfos[offset]
		if !ok {
			t.Errorf("TryInfo at offset %d not found in deserialized chunk", offset)
			continue
		}
		if deserializedInfo != info {
			t.Errorf("TryInfo[%d] mismatch: expected %+v, got %+v", offset, info, deserializedInfo)
		}
	}
}

func TestSerializer_HelperInfo(t *testing.T) {
	chunk := NewChunk("test")
	chunk.Helpers = map[string]*HelperInfo{
		"TStringHelper": {
			Name:         "TStringHelper",
			TargetType:   "String",
			ParentHelper: "",
			Methods: map[string]uint16{
				"Length":    0,
				"ToUpper":   1,
				"Substring": 2,
			},
			Properties:  []string{"Chars", "Length"},
			ClassVars:   []string{"DefaultEncoding"},
			ClassConsts: map[string]Value{"MaxLength": IntValue(1000)},
		},
		"TIntHelper": {
			Name:       "TIntHelper",
			TargetType: "Integer",
			Methods:    map[string]uint16{"ToString": 3},
		},
	}

	serializer := NewSerializer()
	data, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	deserialized, err := serializer.DeserializeChunk(data)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	if len(deserialized.Helpers) != len(chunk.Helpers) {
		t.Fatalf("Helpers length mismatch: expected %d, got %d", len(chunk.Helpers), len(deserialized.Helpers))
	}

	for name, helper := range chunk.Helpers {
		deserializedHelper, ok := deserialized.Helpers[name]
		if !ok {
			t.Errorf("Helper %q not found in deserialized chunk", name)
			continue
		}

		if deserializedHelper.Name != helper.Name {
			t.Errorf("Helper %q name mismatch: expected %q, got %q", name, helper.Name, deserializedHelper.Name)
		}
		if deserializedHelper.TargetType != helper.TargetType {
			t.Errorf("Helper %q target type mismatch: expected %q, got %q", name, helper.TargetType, deserializedHelper.TargetType)
		}
		if deserializedHelper.ParentHelper != helper.ParentHelper {
			t.Errorf("Helper %q parent helper mismatch: expected %q, got %q", name, helper.ParentHelper, deserializedHelper.ParentHelper)
		}

		if len(deserializedHelper.Methods) != len(helper.Methods) {
			t.Errorf("Helper %q methods length mismatch: expected %d, got %d", name, len(helper.Methods), len(deserializedHelper.Methods))
		}
		for methodName, slot := range helper.Methods {
			if deserializedHelper.Methods[methodName] != slot {
				t.Errorf("Helper %q method %q slot mismatch: expected %d, got %d", name, methodName, slot, deserializedHelper.Methods[methodName])
			}
		}

		if len(deserializedHelper.Properties) != len(helper.Properties) {
			t.Errorf("Helper %q properties length mismatch: expected %d, got %d", name, len(helper.Properties), len(deserializedHelper.Properties))
		}

		if len(deserializedHelper.ClassVars) != len(helper.ClassVars) {
			t.Errorf("Helper %q class vars length mismatch: expected %d, got %d", name, len(helper.ClassVars), len(deserializedHelper.ClassVars))
		}

		if len(deserializedHelper.ClassConsts) != len(helper.ClassConsts) {
			t.Errorf("Helper %q class consts length mismatch: expected %d, got %d", name, len(helper.ClassConsts), len(deserializedHelper.ClassConsts))
		}
	}
}

func TestSerializer_VersionCompatibility(t *testing.T) {
	tests := []struct {
		name       string
		current    SerializerVersion
		bytecode   SerializerVersion
		compatible bool
	}{
		{
			name:       "same_version",
			current:    SerializerVersion{1, 0, 0},
			bytecode:   SerializerVersion{1, 0, 0},
			compatible: true,
		},
		{
			name:       "same_major_older_minor",
			current:    SerializerVersion{1, 2, 0},
			bytecode:   SerializerVersion{1, 1, 0},
			compatible: true,
		},
		{
			name:       "same_major_newer_minor",
			current:    SerializerVersion{1, 1, 0},
			bytecode:   SerializerVersion{1, 2, 0},
			compatible: false,
		},
		{
			name:       "different_major",
			current:    SerializerVersion{2, 0, 0},
			bytecode:   SerializerVersion{1, 0, 0},
			compatible: false,
		},
		{
			name:       "same_major_minor_different_patch",
			current:    SerializerVersion{1, 0, 5},
			bytecode:   SerializerVersion{1, 0, 3},
			compatible: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.IsCompatible(tt.bytecode)
			if result != tt.compatible {
				t.Errorf("IsCompatible(%v, %v) = %v, want %v", tt.current, tt.bytecode, result, tt.compatible)
			}
		})
	}
}

func TestSerializer_InvalidMagicNumber(t *testing.T) {
	// Create corrupted data with wrong magic number
	data := []byte("XXX\x00\x01\x00\x00\x00")

	serializer := NewSerializer()
	_, err := serializer.DeserializeChunk(data)
	if err == nil {
		t.Fatal("Expected error for invalid magic number, got nil")
	}
}

func TestSerializer_TooShortData(t *testing.T) {
	// Data too short to contain header
	data := []byte{0x01, 0x02}

	serializer := NewSerializer()
	_, err := serializer.DeserializeChunk(data)
	if err == nil {
		t.Fatal("Expected error for too short data, got nil")
	}
}

func TestSerializer_IncompatibleVersion(t *testing.T) {
	// Create a chunk with incompatible version
	chunk := NewChunk("test")

	// Create serializer with old version
	oldSerializer := &Serializer{
		version: SerializerVersion{Major: 1, Minor: 0, Patch: 0},
	}
	data, err := oldSerializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	// Manually modify version in the data to be newer
	// Header: magic (4 bytes) + major + minor + patch + reserved
	data[5] = 99 // Set minor version to 99

	// Try to deserialize with old version
	_, err = oldSerializer.DeserializeChunk(data)
	if err == nil {
		t.Fatal("Expected error for incompatible version, got nil")
	}
}

func TestSerializer_EmptyChunk(t *testing.T) {
	chunk := NewChunk("empty")

	serializer := NewSerializer()
	data, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	deserialized, err := serializer.DeserializeChunk(data)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	if deserialized.Name != chunk.Name {
		t.Errorf("Name mismatch: expected %q, got %q", chunk.Name, deserialized.Name)
	}
	if len(deserialized.Code) != 0 {
		t.Errorf("Expected empty code, got %d instructions", len(deserialized.Code))
	}
	if len(deserialized.Constants) != 0 {
		t.Errorf("Expected empty constants, got %d constants", len(deserialized.Constants))
	}
}

func TestSerializer_LargeChunk(t *testing.T) {
	chunk := NewChunk("large")

	// Add many instructions
	for i := 0; i < 10000; i++ {
		chunk.WriteInstruction(MakeSimpleInstruction(OpLoadNil), i)
	}

	// Add many constants
	for i := 0; i < 1000; i++ {
		chunk.Constants = append(chunk.Constants, IntValue(int64(i)))
	}

	serializer := NewSerializer()
	data, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	deserialized, err := serializer.DeserializeChunk(data)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	if len(deserialized.Code) != len(chunk.Code) {
		t.Errorf("Code length mismatch: expected %d, got %d", len(chunk.Code), len(deserialized.Code))
	}
	if len(deserialized.Constants) != len(chunk.Constants) {
		t.Errorf("Constants length mismatch: expected %d, got %d", len(chunk.Constants), len(deserialized.Constants))
	}
	if len(deserialized.Lines) != len(chunk.Lines) {
		t.Errorf("Lines length mismatch: expected %d, got %d", len(chunk.Lines), len(deserialized.Lines))
	}
}

func TestSerializer_RoundTrip(t *testing.T) {
	// Create a complex chunk
	chunk := NewChunk("roundtrip_test")
	chunk.LocalCount = 5

	// Add various instructions
	chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
	chunk.WriteInstruction(MakeInstruction(OpStoreLocal, 0, 0), 1)
	chunk.WriteInstruction(MakeInstruction(OpLoadLocal, 0, 0), 2)
	chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 1), 2)
	chunk.WriteInstruction(MakeSimpleInstruction(OpAddInt), 2)
	chunk.WriteInstruction(MakeInstruction(OpJump, 0, 10), 3)
	chunk.WriteInstruction(MakeSimpleInstruction(OpReturn), 4)

	// Add various constants
	chunk.Constants = []Value{
		IntValue(42),
		IntValue(10),
		FloatValue(3.14),
		StringValue("hello"),
		BoolValue(true),
		NilValue(),
	}

	// Add try info
	chunk.SetTryInfo(3, TryInfo{
		CatchTarget:   5,
		FinallyTarget: 8,
		HasCatch:      true,
		HasFinally:    true,
	})

	serializer := NewSerializer()

	// First round trip
	data1, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("First SerializeChunk failed: %v", err)
	}

	deserialized1, err := serializer.DeserializeChunk(data1)
	if err != nil {
		t.Fatalf("First DeserializeChunk failed: %v", err)
	}

	// Second round trip (serialize the deserialized chunk)
	data2, err := serializer.SerializeChunk(deserialized1)
	if err != nil {
		t.Fatalf("Second SerializeChunk failed: %v", err)
	}

	deserialized2, err := serializer.DeserializeChunk(data2)
	if err != nil {
		t.Fatalf("Second DeserializeChunk failed: %v", err)
	}

	// The two serialized forms should be identical
	if !bytes.Equal(data1, data2) {
		t.Error("Round trip data mismatch: serialized forms are not identical")
	}

	// Verify key properties are preserved
	if deserialized2.Name != chunk.Name {
		t.Errorf("Name mismatch after round trip: expected %q, got %q", chunk.Name, deserialized2.Name)
	}
	if deserialized2.LocalCount != chunk.LocalCount {
		t.Errorf("LocalCount mismatch after round trip: expected %d, got %d", chunk.LocalCount, deserialized2.LocalCount)
	}
	if len(deserialized2.Code) != len(chunk.Code) {
		t.Errorf("Code length mismatch after round trip: expected %d, got %d", len(chunk.Code), len(deserialized2.Code))
	}
	if len(deserialized2.Constants) != len(chunk.Constants) {
		t.Errorf("Constants length mismatch after round trip: expected %d, got %d", len(chunk.Constants), len(deserialized2.Constants))
	}
}

// ============================================================================
// Bounds checking tests - prevent memory exhaustion attacks
// ============================================================================

func TestSerializer_BoundsChecking_NegativeChunkSize(t *testing.T) {
	// Create a malicious bytecode with negative chunkSize in a function constant
	// This tests the specific issue mentioned in the PR comment
	buf := new(bytes.Buffer)
	s := NewSerializer()

	// Write header
	buf.Write([]byte(MagicNumber))
	buf.WriteByte(1) // major
	buf.WriteByte(0) // minor
	buf.WriteByte(0) // patch
	buf.WriteByte(0) // reserved

	// Write chunk name
	s.writeString(buf, "malicious")
	s.writeInt32(buf, 0) // local count

	// Write empty instructions
	s.writeInt32(buf, 0) // instruction count

	// Write one constant - a function with negative chunkSize
	s.writeInt32(buf, 1) // constant count
	buf.WriteByte(byte(ValueFunction))
	s.writeString(buf, "badFunc") // function name
	s.writeInt32(buf, 0)          // arity
	s.writeInt32(buf, -1000)      // NEGATIVE chunkSize - should be rejected

	// Try to deserialize - should return error, not panic
	_, err := s.DeserializeChunk(buf.Bytes())
	if err == nil {
		t.Fatal("Expected error for negative chunk size, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("negative")) {
		t.Errorf("Expected error message to mention 'negative', got: %v", err)
	}
}

func TestSerializer_BoundsChecking_ExcessiveChunkSize(t *testing.T) {
	// Test extremely large chunkSize (would cause OOM if not validated)
	buf := new(bytes.Buffer)
	s := NewSerializer()

	// Write header
	buf.Write([]byte(MagicNumber))
	buf.WriteByte(1) // major
	buf.WriteByte(0) // minor
	buf.WriteByte(0) // patch
	buf.WriteByte(0) // reserved

	// Write chunk name
	s.writeString(buf, "malicious")
	s.writeInt32(buf, 0) // local count

	// Write empty instructions
	s.writeInt32(buf, 0) // instruction count

	// Write one constant - a function with huge chunkSize (1GB)
	s.writeInt32(buf, 1) // constant count
	buf.WriteByte(byte(ValueFunction))
	s.writeString(buf, "badFunc")     // function name
	s.writeInt32(buf, 0)              // arity
	s.writeInt32(buf, 1024*1024*1024) // 1GB chunkSize - should be rejected

	// Try to deserialize - should return error, not cause OOM
	_, err := s.DeserializeChunk(buf.Bytes())
	if err == nil {
		t.Fatal("Expected error for excessive chunk size, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("exceeds maximum")) {
		t.Errorf("Expected error message to mention 'exceeds maximum', got: %v", err)
	}
}

func TestSerializer_BoundsChecking_NegativeUpvalueCount(t *testing.T) {
	// Test negative upvalCount in function constant
	// This is the other issue specifically mentioned in the PR comment
	buf := new(bytes.Buffer)
	s := NewSerializer()

	// Write header
	buf.Write([]byte(MagicNumber))
	buf.WriteByte(1) // major
	buf.WriteByte(0) // minor
	buf.WriteByte(0) // patch
	buf.WriteByte(0) // reserved

	// Write chunk name
	s.writeString(buf, "malicious")
	s.writeInt32(buf, 0) // local count

	// Write empty instructions
	s.writeInt32(buf, 0) // instruction count

	// Write one constant - a function with negative upvalCount
	s.writeInt32(buf, 1) // constant count
	buf.WriteByte(byte(ValueFunction))
	s.writeString(buf, "badFunc") // function name
	s.writeInt32(buf, 0)          // arity

	// Write valid inner chunk
	innerChunk := NewChunk("inner")
	innerData, _ := s.SerializeChunk(innerChunk)
	s.writeInt32(buf, int32(len(innerData)))
	buf.Write(innerData)

	s.writeInt32(buf, -500) // NEGATIVE upvalCount - should be rejected

	// Try to deserialize - should return error, not panic
	_, err := s.DeserializeChunk(buf.Bytes())
	if err == nil {
		t.Fatal("Expected error for negative upvalue count, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("negative")) {
		t.Errorf("Expected error message to mention 'negative', got: %v", err)
	}
}

func TestSerializer_BoundsChecking_ExcessiveUpvalueCount(t *testing.T) {
	// Test extremely large upvalCount
	buf := new(bytes.Buffer)
	s := NewSerializer()

	// Write header
	buf.Write([]byte(MagicNumber))
	buf.WriteByte(1) // major
	buf.WriteByte(0) // minor
	buf.WriteByte(0) // patch
	buf.WriteByte(0) // reserved

	// Write chunk name
	s.writeString(buf, "malicious")
	s.writeInt32(buf, 0) // local count

	// Write empty instructions
	s.writeInt32(buf, 0) // instruction count

	// Write one constant - a function with huge upvalCount
	s.writeInt32(buf, 1) // constant count
	buf.WriteByte(byte(ValueFunction))
	s.writeString(buf, "badFunc") // function name
	s.writeInt32(buf, 0)          // arity

	// Write valid inner chunk
	innerChunk := NewChunk("inner")
	innerData, _ := s.SerializeChunk(innerChunk)
	s.writeInt32(buf, int32(len(innerData)))
	buf.Write(innerData)

	s.writeInt32(buf, 100000) // Huge upvalCount - should be rejected

	// Try to deserialize - should return error
	_, err := s.DeserializeChunk(buf.Bytes())
	if err == nil {
		t.Fatal("Expected error for excessive upvalue count, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("exceeds maximum")) {
		t.Errorf("Expected error message to mention 'exceeds maximum', got: %v", err)
	}
}

func TestSerializer_BoundsChecking_ExcessiveStringLength(t *testing.T) {
	// Test extremely large string length
	buf := new(bytes.Buffer)
	s := NewSerializer()

	// Write header
	buf.Write([]byte(MagicNumber))
	buf.WriteByte(1) // major
	buf.WriteByte(0) // minor
	buf.WriteByte(0) // patch
	buf.WriteByte(0) // reserved

	// Write chunk name with excessive length (100MB)
	buf.WriteByte(100)
	buf.WriteByte(0)
	buf.WriteByte(0)
	buf.WriteByte(96) // 100MB in little-endian uint32

	// Try to deserialize - should return error, not cause OOM
	_, err := s.DeserializeChunk(buf.Bytes())
	if err == nil {
		t.Fatal("Expected error for excessive string length, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("exceeds maximum")) {
		t.Errorf("Expected error message to mention 'exceeds maximum', got: %v", err)
	}
}

func TestSerializer_BoundsChecking_ExcessiveInstructionCount(t *testing.T) {
	// Test extremely large instruction count
	buf := new(bytes.Buffer)
	s := NewSerializer()

	// Write header
	buf.Write([]byte(MagicNumber))
	buf.WriteByte(1) // major
	buf.WriteByte(0) // minor
	buf.WriteByte(0) // patch
	buf.WriteByte(0) // reserved

	// Write chunk name
	s.writeString(buf, "malicious")
	s.writeInt32(buf, 0) // local count

	// Write excessive instruction count (10 million)
	s.writeInt32(buf, 10_000_000) // Should be rejected

	// Try to deserialize - should return error
	_, err := s.DeserializeChunk(buf.Bytes())
	if err == nil {
		t.Fatal("Expected error for excessive instruction count, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("exceeds maximum")) {
		t.Errorf("Expected error message to mention 'exceeds maximum', got: %v", err)
	}
}

func TestSerializer_BoundsChecking_ExcessiveConstantCount(t *testing.T) {
	// Test extremely large constant count
	buf := new(bytes.Buffer)
	s := NewSerializer()

	// Write header
	buf.Write([]byte(MagicNumber))
	buf.WriteByte(1) // major
	buf.WriteByte(0) // minor
	buf.WriteByte(0) // patch
	buf.WriteByte(0) // reserved

	// Write chunk name
	s.writeString(buf, "malicious")
	s.writeInt32(buf, 0) // local count

	// Write empty instructions
	s.writeInt32(buf, 0) // instruction count

	// Write excessive constant count (1 million)
	s.writeInt32(buf, 1_000_000) // Should be rejected

	// Try to deserialize - should return error
	_, err := s.DeserializeChunk(buf.Bytes())
	if err == nil {
		t.Fatal("Expected error for excessive constant count, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("exceeds maximum")) {
		t.Errorf("Expected error message to mention 'exceeds maximum', got: %v", err)
	}
}

func TestSerializer_BoundsChecking_ValidMaxValues(t *testing.T) {
	// Test that values at the maximum allowed bounds still work
	// This ensures we're not being too restrictive
	tests := []struct {
		createFn  func() ([]byte, error)
		name      string
		shouldErr bool
	}{
		{
			name: "max_valid_constants",
			createFn: func() ([]byte, error) {
				chunk := NewChunk("test")
				// Add exactly the max allowed constants (this should succeed)
				// Use a smaller number for the test to keep it fast
				for i := 0; i < 1000; i++ {
					chunk.Constants = append(chunk.Constants, IntValue(int64(i)))
				}
				s := NewSerializer()
				return s.SerializeChunk(chunk)
			},
			shouldErr: false,
		},
		{
			name: "max_valid_instructions",
			createFn: func() ([]byte, error) {
				chunk := NewChunk("test")
				// Add many instructions (but not exceeding max)
				for i := 0; i < 10000; i++ {
					chunk.WriteInstruction(MakeSimpleInstruction(OpLoadNil), i)
				}
				s := NewSerializer()
				return s.SerializeChunk(chunk)
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.createFn()
			if err != nil {
				if !tt.shouldErr {
					t.Fatalf("Unexpected error creating chunk: %v", err)
				}
				return
			}

			s := NewSerializer()
			_, err = s.DeserializeChunk(data)
			if tt.shouldErr && err == nil {
				t.Fatal("Expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Fatalf("Unexpected error deserializing: %v", err)
			}
		})
	}
}

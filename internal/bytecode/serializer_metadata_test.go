package bytecode

import (
	"testing"
)

// TestSerializer_ClassMetadata tests serialization of class metadata with field initializers
func TestSerializer_ClassMetadata(t *testing.T) {
	chunk := NewChunk("test")

	// Create class metadata with field initializers
	initChunk1 := NewChunk("TestClass.field1$init")
	initChunk1.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
	initChunk1.WriteInstruction(MakeSimpleInstruction(OpReturn), 1)
	initChunk1.Constants = []Value{IntValue(42)}

	initChunk2 := NewChunk("TestClass.field2$init")
	initChunk2.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
	initChunk2.WriteInstruction(MakeSimpleInstruction(OpReturn), 1)
	initChunk2.Constants = []Value{StringValue("hello")}

	chunk.Classes = map[string]*ClassMetadata{
		"testclass": {
			Name: "TestClass",
			Fields: []*FieldMetadata{
				{
					Name:        "field1",
					Initializer: initChunk1,
				},
				{
					Name:        "field2",
					Initializer: initChunk2,
				},
				{
					Name:        "field3",
					Initializer: nil, // No initializer
				},
			},
		},
	}

	// Serialize and deserialize
	serializer := NewSerializer()
	data, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	deserialized, err := serializer.DeserializeChunk(data)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	// Verify class metadata
	if len(deserialized.Classes) != 1 {
		t.Fatalf("Expected 1 class, got %d", len(deserialized.Classes))
	}

	classMeta, ok := deserialized.Classes["testclass"]
	if !ok {
		t.Fatal("Class 'testclass' not found")
	}

	if classMeta.Name != "TestClass" {
		t.Errorf("Class name mismatch: expected %q, got %q", "TestClass", classMeta.Name)
	}

	if len(classMeta.Fields) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(classMeta.Fields))
	}

	// Verify field1
	if classMeta.Fields[0].Name != "field1" {
		t.Errorf("Field 0 name mismatch: expected %q, got %q", "field1", classMeta.Fields[0].Name)
	}
	if classMeta.Fields[0].Initializer == nil {
		t.Error("Field 0 initializer should not be nil")
	} else {
		if len(classMeta.Fields[0].Initializer.Constants) != 1 {
			t.Errorf("Expected 1 constant in field1 initializer, got %d", len(classMeta.Fields[0].Initializer.Constants))
		}
		if classMeta.Fields[0].Initializer.Constants[0].AsInt() != 42 {
			t.Errorf("Field1 initializer constant mismatch: expected 42, got %d", classMeta.Fields[0].Initializer.Constants[0].AsInt())
		}
	}

	// Verify field2
	if classMeta.Fields[1].Name != "field2" {
		t.Errorf("Field 1 name mismatch: expected %q, got %q", "field2", classMeta.Fields[1].Name)
	}
	if classMeta.Fields[1].Initializer == nil {
		t.Error("Field 1 initializer should not be nil")
	} else {
		if len(classMeta.Fields[1].Initializer.Constants) != 1 {
			t.Errorf("Expected 1 constant in field2 initializer, got %d", len(classMeta.Fields[1].Initializer.Constants))
		}
		if classMeta.Fields[1].Initializer.Constants[0].AsString() != "hello" {
			t.Errorf("Field2 initializer constant mismatch: expected %q, got %q", "hello", classMeta.Fields[1].Initializer.Constants[0].AsString())
		}
	}

	// Verify field3 (no initializer)
	if classMeta.Fields[2].Name != "field3" {
		t.Errorf("Field 2 name mismatch: expected %q, got %q", "field3", classMeta.Fields[2].Name)
	}
	if classMeta.Fields[2].Initializer != nil {
		t.Error("Field 2 initializer should be nil")
	}
}

// TestSerializer_RecordMetadata tests serialization of record metadata with methods
func TestSerializer_RecordMetadata(t *testing.T) {
	chunk := NewChunk("test")

	// Create record metadata with methods
	chunk.Records = map[string]*RecordMetadata{
		"tpoint": {
			Name: "TPoint",
			Methods: map[string]uint16{
				"distance": 0,
				"length":   1,
			},
			Fields: []*FieldMetadata{
				{
					Name:        "x",
					Initializer: nil,
				},
				{
					Name:        "y",
					Initializer: nil,
				},
			},
		},
		"trect": {
			Name: "TRect",
			Methods: map[string]uint16{
				"area":      2,
				"perimeter": 3,
			},
			Fields: []*FieldMetadata{},
		},
	}

	// Serialize and deserialize
	serializer := NewSerializer()
	data, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	deserialized, err := serializer.DeserializeChunk(data)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	// Verify record metadata
	if len(deserialized.Records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(deserialized.Records))
	}

	// Verify TPoint
	pointMeta, ok := deserialized.Records["tpoint"]
	if !ok {
		t.Fatal("Record 'tpoint' not found")
	}

	if pointMeta.Name != "TPoint" {
		t.Errorf("Record name mismatch: expected %q, got %q", "TPoint", pointMeta.Name)
	}

	if len(pointMeta.Methods) != 2 {
		t.Fatalf("Expected 2 methods, got %d", len(pointMeta.Methods))
	}

	if pointMeta.Methods["distance"] != 0 {
		t.Errorf("Method 'distance' slot mismatch: expected 0, got %d", pointMeta.Methods["distance"])
	}

	if pointMeta.Methods["length"] != 1 {
		t.Errorf("Method 'length' slot mismatch: expected 1, got %d", pointMeta.Methods["length"])
	}

	if len(pointMeta.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(pointMeta.Fields))
	}

	// Verify TRect
	rectMeta, ok := deserialized.Records["trect"]
	if !ok {
		t.Fatal("Record 'trect' not found")
	}

	if rectMeta.Name != "TRect" {
		t.Errorf("Record name mismatch: expected %q, got %q", "TRect", rectMeta.Name)
	}

	if len(rectMeta.Methods) != 2 {
		t.Fatalf("Expected 2 methods, got %d", len(rectMeta.Methods))
	}

	if len(rectMeta.Fields) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(rectMeta.Fields))
	}
}

// TestSerializer_RecordWithFieldInitializers tests serialization of record metadata with field initializers
func TestSerializer_RecordWithFieldInitializers(t *testing.T) {
	chunk := NewChunk("test")

	// Create a field initializer
	initChunk := NewChunk("TConfig.timeout$init")
	initChunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
	initChunk.WriteInstruction(MakeSimpleInstruction(OpReturn), 1)
	initChunk.Constants = []Value{IntValue(30)}

	chunk.Records = map[string]*RecordMetadata{
		"tconfig": {
			Name:    "TConfig",
			Methods: map[string]uint16{},
			Fields: []*FieldMetadata{
				{
					Name:        "timeout",
					Initializer: initChunk,
				},
				{
					Name:        "debug",
					Initializer: nil,
				},
			},
		},
	}

	// Serialize and deserialize
	serializer := NewSerializer()
	data, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	deserialized, err := serializer.DeserializeChunk(data)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	// Verify
	recordMeta, ok := deserialized.Records["tconfig"]
	if !ok {
		t.Fatal("Record 'tconfig' not found")
	}

	if len(recordMeta.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(recordMeta.Fields))
	}

	// Verify field with initializer
	if recordMeta.Fields[0].Name != "timeout" {
		t.Errorf("Field name mismatch: expected %q, got %q", "timeout", recordMeta.Fields[0].Name)
	}
	if recordMeta.Fields[0].Initializer == nil {
		t.Error("Field initializer should not be nil")
	} else {
		if len(recordMeta.Fields[0].Initializer.Constants) != 1 {
			t.Errorf("Expected 1 constant, got %d", len(recordMeta.Fields[0].Initializer.Constants))
		}
		if recordMeta.Fields[0].Initializer.Constants[0].AsInt() != 30 {
			t.Errorf("Constant mismatch: expected 30, got %d", recordMeta.Fields[0].Initializer.Constants[0].AsInt())
		}
	}

	// Verify field without initializer
	if recordMeta.Fields[1].Name != "debug" {
		t.Errorf("Field name mismatch: expected %q, got %q", "debug", recordMeta.Fields[1].Name)
	}
	if recordMeta.Fields[1].Initializer != nil {
		t.Error("Field initializer should be nil")
	}
}

// TestSerializer_EmptyMetadata tests serialization with empty metadata maps
func TestSerializer_EmptyMetadata(t *testing.T) {
	chunk := NewChunk("test")
	chunk.Classes = map[string]*ClassMetadata{}
	chunk.Records = map[string]*RecordMetadata{}

	serializer := NewSerializer()
	data, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	deserialized, err := serializer.DeserializeChunk(data)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	if len(deserialized.Classes) != 0 {
		t.Errorf("Expected 0 classes, got %d", len(deserialized.Classes))
	}

	if len(deserialized.Records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(deserialized.Records))
	}
}

// TestSerializer_ComplexClassWithMultipleFields tests complex class with many fields
func TestSerializer_ComplexClassWithMultipleFields(t *testing.T) {
	chunk := NewChunk("test")

	// Create multiple field initializers
	fields := make([]*FieldMetadata, 10)
	for i := 0; i < 10; i++ {
		var initChunk *Chunk
		if i%2 == 0 {
			// Even fields get initializers
			initChunk = NewChunk("TestClass.field$init")
			initChunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
			initChunk.WriteInstruction(MakeSimpleInstruction(OpReturn), 1)
			initChunk.Constants = []Value{IntValue(int64(i * 10))}
		}
		fields[i] = &FieldMetadata{
			Name:        "field" + string(rune('0'+i)),
			Initializer: initChunk,
		}
	}

	chunk.Classes = map[string]*ClassMetadata{
		"testclass": {
			Name:   "TestClass",
			Fields: fields,
		},
	}

	// Serialize and deserialize
	serializer := NewSerializer()
	data, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	deserialized, err := serializer.DeserializeChunk(data)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	// Verify
	classMeta, ok := deserialized.Classes["testclass"]
	if !ok {
		t.Fatal("Class not found")
	}

	if len(classMeta.Fields) != 10 {
		t.Fatalf("Expected 10 fields, got %d", len(classMeta.Fields))
	}

	// Verify even fields have initializers, odd fields don't
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			if classMeta.Fields[i].Initializer == nil {
				t.Errorf("Field %d should have initializer", i)
			}
		} else {
			if classMeta.Fields[i].Initializer != nil {
				t.Errorf("Field %d should not have initializer", i)
			}
		}
	}
}

// TestSerializer_RoundTripWithMetadata tests round-trip serialization with all metadata types
func TestSerializer_RoundTripWithMetadata(t *testing.T) {
	chunk := NewChunk("test")

	// Add class metadata
	initChunk := NewChunk("init")
	initChunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
	initChunk.WriteInstruction(MakeSimpleInstruction(OpReturn), 1)
	initChunk.Constants = []Value{IntValue(100)}

	chunk.Classes = map[string]*ClassMetadata{
		"myclass": {
			Name: "MyClass",
			Fields: []*FieldMetadata{
				{Name: "field1", Initializer: initChunk},
			},
		},
	}

	// Add record metadata
	chunk.Records = map[string]*RecordMetadata{
		"myrecord": {
			Name:    "MyRecord",
			Methods: map[string]uint16{"method1": 5},
			Fields:  []*FieldMetadata{},
		},
	}

	// Add helper metadata
	chunk.Helpers = map[string]*HelperInfo{
		"myhelper": {
			Name:        "MyHelper",
			TargetType:  "Integer",
			Methods:     map[string]uint16{"helper1": 10},
			Properties:  []string{"prop1"},
			ClassVars:   []string{"var1"},
			ClassConsts: map[string]Value{"const1": IntValue(42)},
		},
	}

	serializer := NewSerializer()

	// First round
	data1, err := serializer.SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("First serialize failed: %v", err)
	}

	deserialized1, err := serializer.DeserializeChunk(data1)
	if err != nil {
		t.Fatalf("First deserialize failed: %v", err)
	}

	// Second round
	data2, err := serializer.SerializeChunk(deserialized1)
	if err != nil {
		t.Fatalf("Second serialize failed: %v", err)
	}

	deserialized2, err := serializer.DeserializeChunk(data2)
	if err != nil {
		t.Fatalf("Second deserialize failed: %v", err)
	}

	// Verify all metadata survived two round trips
	if len(deserialized2.Classes) != 1 {
		t.Errorf("Expected 1 class, got %d", len(deserialized2.Classes))
	}
	if len(deserialized2.Records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(deserialized2.Records))
	}
	if len(deserialized2.Helpers) != 1 {
		t.Errorf("Expected 1 helper, got %d", len(deserialized2.Helpers))
	}
}

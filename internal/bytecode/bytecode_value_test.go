package bytecode

import (
	"strings"
	"testing"
)

// TestValueRecordMethods tests record-related Value methods for coverage
func TestValueRecordMethods(t *testing.T) {
	t.Run("IsRecord true", func(t *testing.T) {
		rec := NewRecordInstance("TPoint")
		val := RecordValue(rec)
		if !val.IsRecord() {
			t.Errorf("RecordValue().IsRecord() = false, want true")
		}
	})

	t.Run("IsRecord false", func(t *testing.T) {
		val := IntValue(42)
		if val.IsRecord() {
			t.Errorf("IntValue().IsRecord() = true, want false")
		}
	})

	t.Run("AsRecord valid", func(t *testing.T) {
		rec := NewRecordInstance("TPoint")
		rec.SetField("x", IntValue(10))
		rec.SetField("y", IntValue(20))
		val := RecordValue(rec)

		asRec := val.AsRecord()
		if asRec == nil {
			t.Fatalf("RecordValue().AsRecord() = nil, want record instance")
		}
		if asRec.TypeName != "TPoint" {
			t.Errorf("AsRecord().TypeName = %v, want 'TPoint'", asRec.TypeName)
		}

		x, ok := asRec.GetField("x")
		if !ok || x.AsInt() != 10 {
			t.Errorf("AsRecord().GetField('x') = %v, want 10", x)
		}
	})

	t.Run("AsRecord invalid", func(t *testing.T) {
		val := IntValue(42)
		asRec := val.AsRecord()
		if asRec != nil {
			t.Errorf("IntValue().AsRecord() = %v, want nil", asRec)
		}
	})

	t.Run("IsVariant true", func(t *testing.T) {
		wrapped := IntValue(42)
		val := VariantValue(wrapped)
		if !val.IsVariant() {
			t.Errorf("VariantValue().IsVariant() = false, want true")
		}
	})

	t.Run("IsVariant false", func(t *testing.T) {
		val := IntValue(42)
		if val.IsVariant() {
			t.Errorf("IntValue().IsVariant() = true, want false")
		}
	})

	t.Run("AsVariant valid", func(t *testing.T) {
		wrapped := StringValue("hello")
		val := VariantValue(wrapped)

		asVariant := val.AsVariant()
		if !asVariant.IsString() {
			t.Errorf("VariantValue().AsVariant() should return wrapped string value")
		}
		if asVariant.AsString() != "hello" {
			t.Errorf("AsVariant().AsString() = %v, want 'hello'", asVariant.AsString())
		}
	})

	t.Run("AsVariant invalid", func(t *testing.T) {
		val := IntValue(42)
		asVariant := val.AsVariant()
		if !asVariant.IsNil() {
			t.Errorf("IntValue().AsVariant() = %v, want nil", asVariant)
		}
	})
}

// TestValueStringMethod tests the String() method for various value types
func TestValueStringMethod(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		contains string // Check if output contains this string
	}{
		{
			name:     "nil value",
			value:    NilValue(),
			contains: "nil",
		},
		{
			name:     "bool true",
			value:    BoolValue(true),
			contains: "true",
		},
		{
			name:     "bool false",
			value:    BoolValue(false),
			contains: "false",
		},
		{
			name:     "int value",
			value:    IntValue(42),
			contains: "42",
		},
		{
			name:     "negative int",
			value:    IntValue(-100),
			contains: "-100",
		},
		{
			name:     "float value",
			value:    FloatValue(3.14),
			contains: "3.14",
		},
		{
			name:     "string value",
			value:    StringValue("hello"),
			contains: "hello",
		},
		{
			name:     "empty string",
			value:    StringValue(""),
			contains: "\"\"",
		},
		{
			name:     "array empty",
			value:    ArrayValue(NewArrayInstance([]Value{})),
			contains: "[]",
		},
		{
			name:     "array with values",
			value:    ArrayValue(NewArrayInstance([]Value{IntValue(1), IntValue(2)})),
			contains: "1",
		},
		{
			name:     "function with name",
			value:    FunctionValue(NewFunctionObject("testFunc", NewChunk("test"), 0)),
			contains: "testFunc",
		},
		{
			name:     "function without name",
			value:    FunctionValue(NewFunctionObject("", NewChunk("test"), 0)),
			contains: "<function>",
		},
		{
			name: "closure with name",
			value: ClosureValue(&Closure{
				Function: NewFunctionObject("closureFunc", NewChunk("test"), 0),
			}),
			contains: "closureFunc",
		},
		{
			name:     "closure without name",
			value:    ClosureValue(&Closure{Function: NewFunctionObject("", NewChunk("test"), 0)}),
			contains: "<closure>",
		},
		{
			name:     "object with class name",
			value:    ObjectValue(NewObjectInstance("MyClass")),
			contains: "MyClass",
		},
		{
			name:     "object without class name",
			value:    ObjectValue(NewObjectInstance("")),
			contains: "<object>",
		},
		{
			name:     "record with type name",
			value:    RecordValue(NewRecordInstance("TPoint")),
			contains: "TPoint",
		},
		{
			name:     "record without type name",
			value:    RecordValue(NewRecordInstance("")),
			contains: "<record>",
		},
		{
			name:     "builtin with name",
			value:    BuiltinValue("Print"),
			contains: "Print",
		},
		{
			name:     "builtin without name",
			value:    BuiltinValue(""),
			contains: "<builtin>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.value.String()
			if !strings.Contains(str, tt.contains) {
				t.Errorf("Value.String() = %q, want to contain %q", str, tt.contains)
			}
		})
	}
}

// TestRecordInstance tests RecordInstance methods for coverage
func TestRecordInstance(t *testing.T) {
	t.Run("NewRecordInstance", func(t *testing.T) {
		rec := NewRecordInstance("TPoint")
		if rec == nil {
			t.Fatalf("NewRecordInstance() = nil, want record instance")
		}
		if rec.TypeName != "TPoint" {
			t.Errorf("TypeName = %v, want 'TPoint'", rec.TypeName)
		}
	})

	t.Run("SetField and GetField", func(t *testing.T) {
		rec := NewRecordInstance("TPoint")
		rec.SetField("X", IntValue(10))
		rec.SetField("Y", IntValue(20))

		x, ok := rec.GetField("x") // Case insensitive
		if !ok {
			t.Fatalf("GetField('x') not found")
		}
		if x.AsInt() != 10 {
			t.Errorf("GetField('x') = %v, want 10", x.AsInt())
		}

		y, ok := rec.GetField("Y")
		if !ok {
			t.Fatalf("GetField('Y') not found")
		}
		if y.AsInt() != 20 {
			t.Errorf("GetField('Y') = %v, want 20", y.AsInt())
		}
	})

	t.Run("GetField non-existent", func(t *testing.T) {
		rec := NewRecordInstance("TPoint")
		_, ok := rec.GetField("Z")
		if ok {
			t.Errorf("GetField('Z') should not exist")
		}
	})

	t.Run("GetField nil receiver", func(t *testing.T) {
		var rec *RecordInstance = nil
		val, ok := rec.GetField("X")
		if ok {
			t.Errorf("GetField on nil receiver should return false")
		}
		if !val.IsNil() {
			t.Errorf("GetField on nil receiver should return NilValue")
		}
	})

	t.Run("SetField nil receiver", func(t *testing.T) {
		var rec *RecordInstance = nil
		// Should not panic
		rec.SetField("X", IntValue(10))
	})

	t.Run("SetField with nil fields map", func(t *testing.T) {
		rec := &RecordInstance{TypeName: "Test", fields: nil}
		rec.SetField("X", IntValue(10))
		val, ok := rec.GetField("X")
		if !ok || val.AsInt() != 10 {
			t.Errorf("SetField with nil fields map should create map")
		}
	})
}

// TestArrayInstance tests ArrayInstance methods for improved coverage
func TestArrayInstance(t *testing.T) {
	t.Run("NewArrayInstance empty", func(t *testing.T) {
		arr := NewArrayInstance([]Value{})
		if arr.Length() != 0 {
			t.Errorf("NewArrayInstance([]).Length() = %v, want 0", arr.Length())
		}
	})

	t.Run("NewArrayInstance with values", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1), IntValue(2), IntValue(3)})
		if arr.Length() != 3 {
			t.Errorf("NewArrayInstance([1,2,3]).Length() = %v, want 3", arr.Length())
		}
	})

	t.Run("NewArrayInstanceWithLength", func(t *testing.T) {
		arr := NewArrayInstanceWithLength(5)
		if arr.Length() != 5 {
			t.Errorf("NewArrayInstanceWithLength(5).Length() = %v, want 5", arr.Length())
		}
		// All elements should be nil
		for i := 0; i < arr.Length(); i++ {
			val, ok := arr.Get(i)
			if !ok || !val.IsNil() {
				t.Errorf("NewArrayInstanceWithLength should initialize all elements to nil")
			}
		}
	})

	t.Run("NewArrayInstanceWithLength negative", func(t *testing.T) {
		arr := NewArrayInstanceWithLength(-5)
		if arr.Length() != 0 {
			t.Errorf("NewArrayInstanceWithLength(-5).Length() = %v, want 0", arr.Length())
		}
	})

	t.Run("Get out of bounds", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1)})
		_, ok := arr.Get(10)
		if ok {
			t.Errorf("Get(10) on length 1 array should return false")
		}
	})

	t.Run("Get negative index", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1)})
		_, ok := arr.Get(-1)
		if ok {
			t.Errorf("Get(-1) should return false")
		}
	})

	t.Run("Set out of bounds", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1)})
		ok := arr.Set(10, IntValue(99))
		if ok {
			t.Errorf("Set(10) on length 1 array should return false")
		}
	})

	t.Run("Resize grow", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1), IntValue(2)})
		arr.Resize(5)
		if arr.Length() != 5 {
			t.Errorf("Resize(5) length = %v, want 5", arr.Length())
		}
		// New elements should be nil
		val, ok := arr.Get(4)
		if !ok || !val.IsNil() {
			t.Errorf("Resize should fill new elements with nil")
		}
	})

	t.Run("Resize shrink", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1), IntValue(2), IntValue(3)})
		arr.Resize(1)
		if arr.Length() != 1 {
			t.Errorf("Resize(1) length = %v, want 1", arr.Length())
		}
	})

	t.Run("Resize negative", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1), IntValue(2)})
		arr.Resize(-5)
		if arr.Length() != 0 {
			t.Errorf("Resize(-5) length = %v, want 0", arr.Length())
		}
	})

	t.Run("Resize nil receiver", func(t *testing.T) {
		var arr *ArrayInstance = nil
		// Should not panic
		arr.Resize(10)
	})

	t.Run("Length nil receiver", func(t *testing.T) {
		var arr *ArrayInstance = nil
		if arr.Length() != 0 {
			t.Errorf("nil.Length() = %v, want 0", arr.Length())
		}
	})

	t.Run("String empty array", func(t *testing.T) {
		arr := NewArrayInstance([]Value{})
		str := arr.String()
		if str != "[]" {
			t.Errorf("empty array.String() = %v, want '[]'", str)
		}
	})

	t.Run("String nil array", func(t *testing.T) {
		var arr *ArrayInstance = nil
		str := arr.String()
		if str != "[]" {
			t.Errorf("nil array.String() = %v, want '[]'", str)
		}
	})

	t.Run("String with values", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1), IntValue(2), IntValue(3)})
		str := arr.String()
		if !strings.Contains(str, "1") || !strings.Contains(str, "2") || !strings.Contains(str, "3") {
			t.Errorf("array.String() = %v, want to contain 1, 2, 3", str)
		}
	})
}

// TestObjectInstanceMethods tests ObjectInstance methods for improved coverage
func TestObjectInstanceMethods(t *testing.T) {
	t.Run("GetField nil receiver", func(t *testing.T) {
		var obj *ObjectInstance = nil
		_, ok := obj.GetField("test")
		if ok {
			t.Errorf("GetField on nil receiver should return false")
		}
	})

	t.Run("SetField nil receiver", func(t *testing.T) {
		var obj *ObjectInstance = nil
		// Should not panic
		obj.SetField("test", IntValue(42))
	})

	t.Run("SetField with nil fields map", func(t *testing.T) {
		obj := &ObjectInstance{ClassName: "Test", fields: nil}
		obj.SetField("X", IntValue(10))
		val, ok := obj.GetField("X")
		if !ok || val.AsInt() != 10 {
			t.Errorf("SetField with nil fields map should create map")
		}
	})

	t.Run("GetProperty nil receiver", func(t *testing.T) {
		var obj *ObjectInstance = nil
		_, ok := obj.GetProperty("test")
		if ok {
			t.Errorf("GetProperty on nil receiver should return false")
		}
	})

	t.Run("SetProperty nil receiver", func(t *testing.T) {
		var obj *ObjectInstance = nil
		// Should not panic
		obj.SetProperty("test", IntValue(42))
	})

	t.Run("SetProperty with nil props map", func(t *testing.T) {
		obj := &ObjectInstance{ClassName: "Test", props: nil}
		obj.SetProperty("X", IntValue(10))
		val, ok := obj.GetProperty("X")
		if !ok || val.AsInt() != 10 {
			t.Errorf("SetProperty with nil props map should create map")
		}
	})

	t.Run("GetProperty falls back to field", func(t *testing.T) {
		obj := NewObjectInstance("Test")
		obj.SetField("X", IntValue(42))
		val, ok := obj.GetProperty("X")
		if !ok || val.AsInt() != 42 {
			t.Errorf("GetProperty should fall back to field if property not found")
		}
	})

	t.Run("GetProperty prefers property over field", func(t *testing.T) {
		obj := NewObjectInstance("Test")
		obj.SetField("X", IntValue(10))
		obj.SetProperty("X", IntValue(20))
		val, ok := obj.GetProperty("X")
		if !ok || val.AsInt() != 20 {
			t.Errorf("GetProperty should prefer property over field")
		}
	})
}

// TestUpvalueMethods tests Upvalue methods for coverage
func TestUpvalueMethods(t *testing.T) {
	t.Run("newOpenUpvalue", func(t *testing.T) {
		val := IntValue(42)
		uv := newOpenUpvalue(&val)
		if uv == nil {
			t.Fatalf("newOpenUpvalue() = nil")
		}
		if uv.get().AsInt() != 42 {
			t.Errorf("upvalue.get() = %v, want 42", uv.get())
		}
	})

	t.Run("upvalue set", func(t *testing.T) {
		val := IntValue(42)
		uv := newOpenUpvalue(&val)
		uv.set(IntValue(100))
		if uv.get().AsInt() != 100 {
			t.Errorf("upvalue.get() after set = %v, want 100", uv.get())
		}
	})

	t.Run("upvalue close", func(t *testing.T) {
		val := IntValue(42)
		uv := newOpenUpvalue(&val)
		uv.close()
		// After close, location should be nil
		if uv.location != nil {
			t.Errorf("upvalue.close() should set location to nil")
		}
		// But value should still be accessible
		if uv.get().AsInt() != 42 {
			t.Errorf("upvalue.get() after close = %v, want 42", uv.get())
		}
	})

	t.Run("upvalue set after close", func(t *testing.T) {
		val := IntValue(42)
		uv := newOpenUpvalue(&val)
		uv.close()
		uv.set(IntValue(100))
		if uv.get().AsInt() != 100 {
			t.Errorf("upvalue.get() after close and set = %v, want 100", uv.get())
		}
	})

	t.Run("upvalue get nil receiver", func(t *testing.T) {
		var uv *Upvalue = nil
		val := uv.get()
		if !val.IsNil() {
			t.Errorf("nil upvalue.get() should return NilValue")
		}
	})

	t.Run("upvalue set nil receiver", func(t *testing.T) {
		var uv *Upvalue = nil
		// Should not panic
		uv.set(IntValue(42))
	})

	t.Run("upvalue close nil receiver", func(t *testing.T) {
		var uv *Upvalue = nil
		// Should not panic
		uv.close()
	})
}

// TestFunctionObject tests FunctionObject methods
func TestFunctionObject(t *testing.T) {
	t.Run("NewFunctionObject", func(t *testing.T) {
		chunk := NewChunk("test")
		fn := NewFunctionObject("testFunc", chunk, 2)
		if fn.Name != "testFunc" {
			t.Errorf("Function name = %v, want 'testFunc'", fn.Name)
		}
		if fn.Arity != 2 {
			t.Errorf("Function arity = %v, want 2", fn.Arity)
		}
		if fn.Chunk != chunk {
			t.Errorf("Function chunk mismatch")
		}
	})

	t.Run("UpvalueCount", func(t *testing.T) {
		fn := NewFunctionObject("test", NewChunk("test"), 0)
		fn.UpvalueDefs = []UpvalueDef{
			{IsLocal: true, Index: 0},
			{IsLocal: false, Index: 1},
		}
		if fn.UpvalueCount() != 2 {
			t.Errorf("UpvalueCount() = %v, want 2", fn.UpvalueCount())
		}
	})

	t.Run("UpvalueCount nil receiver", func(t *testing.T) {
		var fn *FunctionObject = nil
		if fn.UpvalueCount() != 0 {
			t.Errorf("nil UpvalueCount() = %v, want 0", fn.UpvalueCount())
		}
	})
}

// TestValueAsConversionMethods tests As* conversion methods
func TestValueAsConversionMethods(t *testing.T) {
	t.Run("AsBool valid", func(t *testing.T) {
		val := BoolValue(true)
		if !val.AsBool() {
			t.Errorf("BoolValue(true).AsBool() = false, want true")
		}
	})

	t.Run("AsBool invalid", func(t *testing.T) {
		val := IntValue(42)
		if val.AsBool() {
			t.Errorf("IntValue.AsBool() = true, want false")
		}
	})

	t.Run("AsInt valid", func(t *testing.T) {
		val := IntValue(42)
		if val.AsInt() != 42 {
			t.Errorf("IntValue(42).AsInt() = %v, want 42", val.AsInt())
		}
	})

	t.Run("AsInt invalid", func(t *testing.T) {
		val := StringValue("hello")
		if val.AsInt() != 0 {
			t.Errorf("StringValue.AsInt() = %v, want 0", val.AsInt())
		}
	})

	t.Run("AsFloat from float", func(t *testing.T) {
		val := FloatValue(3.14)
		if val.AsFloat() != 3.14 {
			t.Errorf("FloatValue(3.14).AsFloat() = %v, want 3.14", val.AsFloat())
		}
	})

	t.Run("AsFloat from int", func(t *testing.T) {
		val := IntValue(42)
		if val.AsFloat() != 42.0 {
			t.Errorf("IntValue(42).AsFloat() = %v, want 42.0", val.AsFloat())
		}
	})

	t.Run("AsFloat invalid", func(t *testing.T) {
		val := StringValue("hello")
		if val.AsFloat() != 0.0 {
			t.Errorf("StringValue.AsFloat() = %v, want 0.0", val.AsFloat())
		}
	})

	t.Run("AsString valid", func(t *testing.T) {
		val := StringValue("hello")
		if val.AsString() != "hello" {
			t.Errorf("StringValue('hello').AsString() = %v, want 'hello'", val.AsString())
		}
	})

	t.Run("AsString invalid", func(t *testing.T) {
		val := IntValue(42)
		if val.AsString() != "" {
			t.Errorf("IntValue.AsString() = %v, want empty string", val.AsString())
		}
	})

	t.Run("AsArray valid", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1), IntValue(2)})
		val := ArrayValue(arr)
		result := val.AsArray()
		if result == nil {
			t.Errorf("ArrayValue.AsArray() = nil, want array instance")
		}
		if result != nil && result.Length() != 2 {
			t.Errorf("AsArray().Length() = %v, want 2", result.Length())
		}
	})

	t.Run("AsArray invalid", func(t *testing.T) {
		val := IntValue(42)
		result := val.AsArray()
		if result != nil {
			t.Errorf("IntValue.AsArray() = %v, want nil", result)
		}
	})

	t.Run("AsFunction valid", func(t *testing.T) {
		fn := NewFunctionObject("test", NewChunk("test"), 0)
		val := FunctionValue(fn)
		result := val.AsFunction()
		if result == nil {
			t.Errorf("FunctionValue.AsFunction() = nil, want function object")
		}
		if result != nil && result.Name != "test" {
			t.Errorf("AsFunction().Name = %v, want 'test'", result.Name)
		}
	})

	t.Run("AsFunction invalid", func(t *testing.T) {
		val := IntValue(42)
		result := val.AsFunction()
		if result != nil {
			t.Errorf("IntValue.AsFunction() = %v, want nil", result)
		}
	})

	t.Run("AsClosure valid", func(t *testing.T) {
		closure := &Closure{Function: NewFunctionObject("test", NewChunk("test"), 0)}
		val := ClosureValue(closure)
		result := val.AsClosure()
		if result == nil {
			t.Errorf("ClosureValue.AsClosure() = nil, want closure")
		}
	})

	t.Run("AsClosure invalid", func(t *testing.T) {
		val := IntValue(42)
		result := val.AsClosure()
		if result != nil {
			t.Errorf("IntValue.AsClosure() = %v, want nil", result)
		}
	})

	t.Run("AsObject valid", func(t *testing.T) {
		obj := NewObjectInstance("MyClass")
		val := ObjectValue(obj)
		result := val.AsObject()
		if result == nil {
			t.Errorf("ObjectValue.AsObject() = nil, want object instance")
		}
		if result != nil && result.ClassName != "MyClass" {
			t.Errorf("AsObject().ClassName = %v, want 'MyClass'", result.ClassName)
		}
	})

	t.Run("AsObject invalid", func(t *testing.T) {
		val := IntValue(42)
		result := val.AsObject()
		if result != nil {
			t.Errorf("IntValue.AsObject() = %v, want nil", result)
		}
	})

	t.Run("AsBuiltin valid", func(t *testing.T) {
		val := BuiltinValue("Print")
		result := val.AsBuiltin()
		if result != "Print" {
			t.Errorf("BuiltinValue('Print').AsBuiltin() = %v, want 'Print'", result)
		}
	})

	t.Run("AsBuiltin invalid", func(t *testing.T) {
		val := IntValue(42)
		result := val.AsBuiltin()
		if result != "" {
			t.Errorf("IntValue.AsBuiltin() = %v, want empty string", result)
		}
	})
}

// TestZeroValueForType tests zeroValueForType helper
func TestZeroValueForType(t *testing.T) {
	tests := []struct {
		want Value
		vt   ValueType
	}{
		{vt: ValueInt, want: IntValue(0)},
		{vt: ValueFloat, want: FloatValue(0.0)},
		{vt: ValueString, want: StringValue("")},
		{vt: ValueBool, want: BoolValue(false)},
		{vt: ValueNil, want: NilValue()},
		{vt: ValueArray, want: NilValue()},
		{vt: ValueObject, want: NilValue()},
	}

	for _, tt := range tests {
		t.Run(tt.vt.String(), func(t *testing.T) {
			got := zeroValueForType(tt.vt)
			if got.Type != tt.want.Type {
				t.Errorf("zeroValueForType(%v) type = %v, want %v", tt.vt, got.Type, tt.want.Type)
			}
		})
	}
}

// TestValueTypeString tests ValueType.String() method
func TestValueTypeString(t *testing.T) {
	tests := []struct {
		want string
		vt   ValueType
	}{
		{vt: ValueNil, want: "nil"},
		{vt: ValueBool, want: "bool"},
		{vt: ValueInt, want: "int"},
		{vt: ValueFloat, want: "float"},
		{vt: ValueString, want: "string"},
		{vt: ValueArray, want: "array"},
		{vt: ValueObject, want: "object"},
		{vt: ValueRecord, want: "record"},
		{vt: ValueFunction, want: "function"},
		{vt: ValueClosure, want: "closure"},
		{vt: ValueBuiltin, want: "builtin"},
		{vt: ValueVariant, want: "variant"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.vt.String()
			if got != tt.want {
				t.Errorf("ValueType(%v).String() = %v, want %v", tt.vt, got, tt.want)
			}
		})
	}

	t.Run("unknown type", func(t *testing.T) {
		vt := ValueType(255)
		if vt.String() != "unknown" {
			t.Errorf("Unknown ValueType.String() = %v, want 'unknown'", vt.String())
		}
	})
}

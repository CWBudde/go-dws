package runtime

import (
	"testing"
)

// ============================================================================
// Type Checking Tests
// ============================================================================

func TestTypeCheckers(t *testing.T) {
	intVal := NewInteger(42)
	floatVal := NewFloat(3.14)
	strVal := NewString("hello")
	boolVal := NewBoolean(true)
	nilVal := &NilValue{}

	t.Run("IsInteger", func(t *testing.T) {
		if !IsInteger(intVal) {
			t.Error("IsInteger() should return true for IntegerValue")
		}
		if IsInteger(floatVal) {
			t.Error("IsInteger() should return false for FloatValue")
		}
	})

	t.Run("IsFloat", func(t *testing.T) {
		if !IsFloat(floatVal) {
			t.Error("IsFloat() should return true for FloatValue")
		}
		if IsFloat(intVal) {
			t.Error("IsFloat() should return false for IntegerValue")
		}
	})

	t.Run("IsString", func(t *testing.T) {
		if !IsString(strVal) {
			t.Error("IsString() should return true for StringValue")
		}
		if IsString(intVal) {
			t.Error("IsString() should return false for IntegerValue")
		}
	})

	t.Run("IsBoolean", func(t *testing.T) {
		if !IsBoolean(boolVal) {
			t.Error("IsBoolean() should return true for BooleanValue")
		}
		if IsBoolean(intVal) {
			t.Error("IsBoolean() should return false for IntegerValue")
		}
	})

	t.Run("IsNil", func(t *testing.T) {
		if !IsNil(nilVal) {
			t.Error("IsNil() should return true for NilValue")
		}
		if IsNil(intVal) {
			t.Error("IsNil() should return false for IntegerValue")
		}
	})
}

func TestAsTypeConversions(t *testing.T) {
	intVal := NewInteger(42)
	floatVal := NewFloat(3.14)

	t.Run("AsInteger", func(t *testing.T) {
		if result := AsInteger(intVal); result == nil {
			t.Error("AsInteger() should return non-nil for IntegerValue")
		}
		if result := AsInteger(floatVal); result != nil {
			t.Error("AsInteger() should return nil for FloatValue")
		}
	})

	t.Run("AsFloat", func(t *testing.T) {
		if result := AsFloat(floatVal); result == nil {
			t.Error("AsFloat() should return non-nil for FloatValue")
		}
		if result := AsFloat(intVal); result != nil {
			t.Error("AsFloat() should return nil for IntegerValue")
		}
	})
}

func TestInterfaceTypeCheckers(t *testing.T) {
	intVal := NewInteger(42)
	strVal := NewString("hello")

	t.Run("IsNumeric", func(t *testing.T) {
		if !IsNumeric(intVal) {
			t.Error("IsNumeric() should return true for IntegerValue")
		}
		if IsNumeric(strVal) {
			t.Error("IsNumeric() should return false for StringValue")
		}
	})

	t.Run("IsComparable", func(t *testing.T) {
		if !IsComparable(intVal) {
			t.Error("IsComparable() should return true for IntegerValue")
		}
		// All primitive values implement ComparableValue
		if !IsComparable(strVal) {
			t.Error("IsComparable() should return true for StringValue")
		}
	})

	t.Run("IsOrderable", func(t *testing.T) {
		if !IsOrderable(intVal) {
			t.Error("IsOrderable() should return true for IntegerValue")
		}
		if !IsOrderable(strVal) {
			t.Error("IsOrderable() should return true for StringValue")
		}
	})

	t.Run("IsIndexable", func(t *testing.T) {
		if IsIndexable(intVal) {
			t.Error("IsIndexable() should return false for IntegerValue")
		}
		if !IsIndexable(strVal) {
			t.Error("IsIndexable() should return true for StringValue")
		}
	})
}

// ============================================================================
// Truthiness Tests
// ============================================================================

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  bool
	}{
		{name: "nil", value: nil, want: false},
		{name: "Boolean true", value: NewBoolean(true), want: true},
		{name: "Boolean false", value: NewBoolean(false), want: false},
		{name: "Integer non-zero", value: NewInteger(42), want: true},
		{name: "Integer zero", value: NewInteger(0), want: false},
		{name: "Float non-zero", value: NewFloat(3.14), want: true},
		{name: "Float zero", value: NewFloat(0.0), want: false},
		{name: "String non-empty", value: NewString("hello"), want: true},
		{name: "String empty", value: NewString(""), want: false},
		{name: "NilValue", value: &NilValue{}, want: false},
		{name: "NullValue", value: &NullValue{}, want: false},
		{name: "UnassignedValue", value: &UnassignedValue{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTruthy(tt.value)
			if got != tt.want {
				t.Errorf("IsTruthy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsFalsy(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  bool
	}{
		{name: "nil", value: nil, want: true},
		{name: "Boolean true", value: NewBoolean(true), want: false},
		{name: "Boolean false", value: NewBoolean(false), want: true},
		{name: "Integer zero", value: NewInteger(0), want: true},
		{name: "Integer non-zero", value: NewInteger(42), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFalsy(tt.value)
			if got != tt.want {
				t.Errorf("IsFalsy() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// Comparison Tests
// ============================================================================

func TestEqual(t *testing.T) {
	tests := []struct {
		name    string
		left    Value
		right   Value
		want    bool
		wantErr bool
	}{
		{name: "equal integers", left: NewInteger(42), right: NewInteger(42), want: true, wantErr: false},
		{name: "different integers", left: NewInteger(42), right: NewInteger(43), want: false, wantErr: false},
		{name: "equal strings", left: NewString("hello"), right: NewString("hello"), want: true, wantErr: false},
		{name: "different strings", left: NewString("hello"), right: NewString("world"), want: false, wantErr: false},
		{name: "int equals float", left: NewInteger(42), right: NewFloat(42.0), want: true, wantErr: false},
		{name: "both nil", left: nil, right: nil, want: true, wantErr: false},
		{name: "left nil", left: nil, right: NewInteger(42), want: false, wantErr: false},
		{name: "right nil", left: NewInteger(42), right: nil, want: false, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Equal(tt.left, tt.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("Equal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLessThan(t *testing.T) {
	tests := []struct {
		name    string
		left    Value
		right   Value
		want    bool
		wantErr bool
	}{
		{name: "2 < 3", left: NewInteger(2), right: NewInteger(3), want: true, wantErr: false},
		{name: "3 < 2", left: NewInteger(3), right: NewInteger(2), want: false, wantErr: false},
		{name: "2.5 < 3.5", left: NewFloat(2.5), right: NewFloat(3.5), want: true, wantErr: false},
		{name: "a < b", left: NewString("a"), right: NewString("b"), want: true, wantErr: false},
		{name: "b < a", left: NewString("b"), right: NewString("a"), want: false, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LessThan(tt.left, tt.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("LessThan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGreaterThan(t *testing.T) {
	tests := []struct {
		name    string
		left    Value
		right   Value
		want    bool
		wantErr bool
	}{
		{name: "3 > 2", left: NewInteger(3), right: NewInteger(2), want: true, wantErr: false},
		{name: "2 > 3", left: NewInteger(2), right: NewInteger(3), want: false, wantErr: false},
		{name: "b > a", left: NewString("b"), right: NewString("a"), want: true, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GreaterThan(tt.left, tt.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("GreaterThan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GreaterThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// String Utilities Tests
// ============================================================================

func TestStringConcat(t *testing.T) {
	result := StringConcat(NewString("Hello"), NewString(" World"))
	expected := "Hello World"
	if result.String() != expected {
		t.Errorf("StringConcat() = %v, want %v", result.String(), expected)
	}
}

func TestStringContains(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		substr string
		want   bool
	}{
		{"contains", "Hello World", "World", true},
		{"not contains", "Hello World", "xyz", false},
		{"case sensitive", "Hello World", "world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringContains(tt.str, tt.substr)
			if got != tt.want {
				t.Errorf("StringContains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringContainsInsensitive(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		substr string
		want   bool
	}{
		{"contains exact", "Hello World", "World", true},
		{"contains different case", "Hello World", "world", true},
		{"not contains", "Hello World", "xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringContainsInsensitive(tt.str, tt.substr)
			if got != tt.want {
				t.Errorf("StringContainsInsensitive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringEqualsInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		left  string
		right string
		want  bool
	}{
		{"exact match", "Hello", "Hello", true},
		{"different case", "Hello", "hello", true},
		{"different", "Hello", "World", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringEqualsInsensitive(tt.left, tt.right)
			if got != tt.want {
				t.Errorf("StringEqualsInsensitive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// Utility Function Tests
// ============================================================================

func TestValuesEqual(t *testing.T) {
	if !ValuesEqual(NewInteger(42), NewInteger(42)) {
		t.Error("ValuesEqual() should return true for equal integers")
	}
	if ValuesEqual(NewInteger(42), NewInteger(43)) {
		t.Error("ValuesEqual() should return false for different integers")
	}
}

func TestCopyValue(t *testing.T) {
	original := NewInteger(42)
	copied := CopyValue(original)

	// Should be a copy
	if !ValuesEqual(original, copied) {
		t.Error("CopyValue() should return equal value")
	}

	// For primitives, copy might be same instance (value semantics)
	// That's OK - the important thing is it implements CopyableValue
}

func TestGetTypeName(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{"Integer", NewInteger(42), "INTEGER"},
		{"Float", NewFloat(3.14), "FLOAT"},
		{"String", NewString("hello"), "STRING"},
		{"Boolean", NewBoolean(true), "BOOLEAN"},
		{"nil", nil, "nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTypeName(tt.value)
			if got != tt.want {
				t.Errorf("GetTypeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypesMatch(t *testing.T) {
	if !TypesMatch(NewInteger(1), NewInteger(2)) {
		t.Error("TypesMatch() should return true for two integers")
	}
	if TypesMatch(NewInteger(1), NewFloat(1.0)) {
		t.Error("TypesMatch() should return false for integer and float")
	}
	if !TypesMatch(nil, nil) {
		t.Error("TypesMatch() should return true for two nils")
	}
	if TypesMatch(NewInteger(1), nil) {
		t.Error("TypesMatch() should return false for integer and nil")
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkIsInteger(b *testing.B) {
	val := NewInteger(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsInteger(val)
	}
}

func BenchmarkAsInteger(b *testing.B) {
	val := NewInteger(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AsInteger(val)
	}
}

func BenchmarkIsTruthy(b *testing.B) {
	val := NewInteger(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsTruthy(val)
	}
}

func BenchmarkEqual(b *testing.B) {
	left := NewInteger(42)
	right := NewInteger(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Equal(left, right)
	}
}

func BenchmarkLessThan(b *testing.B) {
	left := NewInteger(42)
	right := NewInteger(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = LessThan(left, right)
	}
}

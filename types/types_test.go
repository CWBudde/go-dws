package types

import (
	"testing"
)

// ============================================================================
// Basic Type Tests
// ============================================================================

func TestBasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		typ      Type
		expected string
		kind     string
	}{
		{"Integer", INTEGER, "Integer", "INTEGER"},
		{"Float", FLOAT, "Float", "FLOAT"},
		{"String", STRING, "String", "STRING"},
		{"Boolean", BOOLEAN, "Boolean", "BOOLEAN"},
		{"Nil", NIL, "Nil", "NIL"},
		{"Void", VOID, "Void", "VOID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.typ.String() != tt.expected {
				t.Errorf("String() = %v, want %v", tt.typ.String(), tt.expected)
			}
			if tt.typ.TypeKind() != tt.kind {
				t.Errorf("TypeKind() = %v, want %v", tt.typ.TypeKind(), tt.kind)
			}
		})
	}
}

func TestBasicTypeEquality(t *testing.T) {
	tests := []struct {
		name     string
		a, b     Type
		expected bool
	}{
		{"Integer equals Integer", INTEGER, INTEGER, true},
		{"Float equals Float", FLOAT, FLOAT, true},
		{"String equals String", STRING, STRING, true},
		{"Boolean equals Boolean", BOOLEAN, BOOLEAN, true},
		{"Nil equals Nil", NIL, NIL, true},
		{"Void equals Void", VOID, VOID, true},
		{"Integer not equals Float", INTEGER, FLOAT, false},
		{"String not equals Boolean", STRING, BOOLEAN, false},
		{"Integer not equals String", INTEGER, STRING, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.a.Equals(tt.b); result != tt.expected {
				t.Errorf("Equals() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTypeUtilities(t *testing.T) {
	t.Run("IsBasicType", func(t *testing.T) {
		if !IsBasicType(INTEGER) {
			t.Error("INTEGER should be a basic type")
		}
		if !IsBasicType(FLOAT) {
			t.Error("FLOAT should be a basic type")
		}
		if !IsBasicType(STRING) {
			t.Error("STRING should be a basic type")
		}
		if !IsBasicType(BOOLEAN) {
			t.Error("BOOLEAN should be a basic type")
		}
		if IsBasicType(NIL) {
			t.Error("NIL should not be a basic type")
		}
		if IsBasicType(VOID) {
			t.Error("VOID should not be a basic type")
		}
	})

	t.Run("IsNumericType", func(t *testing.T) {
		if !IsNumericType(INTEGER) {
			t.Error("INTEGER should be numeric")
		}
		if !IsNumericType(FLOAT) {
			t.Error("FLOAT should be numeric")
		}
		if IsNumericType(STRING) {
			t.Error("STRING should not be numeric")
		}
		if IsNumericType(BOOLEAN) {
			t.Error("BOOLEAN should not be numeric")
		}
	})

	t.Run("IsOrdinalType", func(t *testing.T) {
		if !IsOrdinalType(INTEGER) {
			t.Error("INTEGER should be ordinal")
		}
		if !IsOrdinalType(BOOLEAN) {
			t.Error("BOOLEAN should be ordinal")
		}
		if IsOrdinalType(FLOAT) {
			t.Error("FLOAT should not be ordinal")
		}
		if IsOrdinalType(STRING) {
			t.Error("STRING should not be ordinal")
		}
	})
}

func TestTypeFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Type
		wantErr  bool
	}{
		{"Integer", "Integer", INTEGER, false},
		{"Float", "Float", FLOAT, false},
		{"String", "String", STRING, false},
		{"Boolean", "Boolean", BOOLEAN, false},
		{"Void", "Void", VOID, false},
		{"Unknown", "Unknown", nil, true},
		{"Empty", "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TypeFromString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("TypeFromString() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

// ============================================================================
// Function Type Tests
// ============================================================================

func TestFunctionType(t *testing.T) {
	t.Run("Simple function", func(t *testing.T) {
		ft := NewFunctionType([]Type{INTEGER, STRING}, BOOLEAN)
		expected := "(Integer, String) -> Boolean"
		if ft.String() != expected {
			t.Errorf("String() = %v, want %v", ft.String(), expected)
		}
		if ft.TypeKind() != "FUNCTION" {
			t.Errorf("TypeKind() = %v, want FUNCTION", ft.TypeKind())
		}
		if ft.IsProcedure() {
			t.Error("Should not be a procedure")
		}
		if !ft.IsFunction() {
			t.Error("Should be a function")
		}
	})

	t.Run("Procedure (void return)", func(t *testing.T) {
		pt := NewProcedureType([]Type{INTEGER})
		expected := "(Integer) -> Void"
		if pt.String() != expected {
			t.Errorf("String() = %v, want %v", pt.String(), expected)
		}
		if !pt.IsProcedure() {
			t.Error("Should be a procedure")
		}
		if pt.IsFunction() {
			t.Error("Should not be a function")
		}
	})

	t.Run("No parameters", func(t *testing.T) {
		ft := NewFunctionType([]Type{}, INTEGER)
		expected := "() -> Integer"
		if ft.String() != expected {
			t.Errorf("String() = %v, want %v", ft.String(), expected)
		}
	})
}

func TestFunctionTypeEquality(t *testing.T) {
	tests := []struct {
		name     string
		a, b     Type
		expected bool
	}{
		{
			"Same function types",
			NewFunctionType([]Type{INTEGER, STRING}, BOOLEAN),
			NewFunctionType([]Type{INTEGER, STRING}, BOOLEAN),
			true,
		},
		{
			"Different parameter types",
			NewFunctionType([]Type{INTEGER}, BOOLEAN),
			NewFunctionType([]Type{FLOAT}, BOOLEAN),
			false,
		},
		{
			"Different return types",
			NewFunctionType([]Type{INTEGER}, BOOLEAN),
			NewFunctionType([]Type{INTEGER}, INTEGER),
			false,
		},
		{
			"Different parameter count",
			NewFunctionType([]Type{INTEGER}, BOOLEAN),
			NewFunctionType([]Type{INTEGER, STRING}, BOOLEAN),
			false,
		},
		{
			"Function vs non-function",
			NewFunctionType([]Type{}, INTEGER),
			INTEGER,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.a.Equals(tt.b); result != tt.expected {
				t.Errorf("Equals() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Array Type Tests
// ============================================================================

func TestArrayType(t *testing.T) {
	t.Run("Dynamic array", func(t *testing.T) {
		at := NewDynamicArrayType(INTEGER)
		expected := "array of Integer"
		if at.String() != expected {
			t.Errorf("String() = %v, want %v", at.String(), expected)
		}
		if !at.IsDynamic() {
			t.Error("Should be dynamic")
		}
		if at.IsStatic() {
			t.Error("Should not be static")
		}
		if at.Size() != -1 {
			t.Errorf("Size() = %v, want -1", at.Size())
		}
	})

	t.Run("Static array", func(t *testing.T) {
		at := NewStaticArrayType(STRING, 1, 10)
		expected := "array[1..10] of String"
		if at.String() != expected {
			t.Errorf("String() = %v, want %v", at.String(), expected)
		}
		if at.IsDynamic() {
			t.Error("Should not be dynamic")
		}
		if !at.IsStatic() {
			t.Error("Should be static")
		}
		if at.Size() != 10 {
			t.Errorf("Size() = %v, want 10", at.Size())
		}
	})
}

func TestArrayTypeEquality(t *testing.T) {
	tests := []struct {
		name     string
		a, b     Type
		expected bool
	}{
		{
			"Same dynamic arrays",
			NewDynamicArrayType(INTEGER),
			NewDynamicArrayType(INTEGER),
			true,
		},
		{
			"Different element types (dynamic)",
			NewDynamicArrayType(INTEGER),
			NewDynamicArrayType(FLOAT),
			false,
		},
		{
			"Same static arrays",
			NewStaticArrayType(STRING, 1, 10),
			NewStaticArrayType(STRING, 1, 10),
			true,
		},
		{
			"Different bounds",
			NewStaticArrayType(INTEGER, 1, 10),
			NewStaticArrayType(INTEGER, 0, 9),
			false,
		},
		{
			"Dynamic vs static",
			NewDynamicArrayType(INTEGER),
			NewStaticArrayType(INTEGER, 1, 10),
			false,
		},
		{
			"Array vs non-array",
			NewDynamicArrayType(INTEGER),
			INTEGER,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.a.Equals(tt.b); result != tt.expected {
				t.Errorf("Equals() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Record Type Tests
// ============================================================================

func TestRecordType(t *testing.T) {
	t.Run("Named record", func(t *testing.T) {
		fields := map[string]Type{
			"X": INTEGER,
			"Y": INTEGER,
		}
		rt := NewRecordType("TPoint", fields)
		if rt.String() != "TPoint" {
			t.Errorf("String() = %v, want TPoint", rt.String())
		}
		if !rt.HasField("X") {
			t.Error("Should have field X")
		}
		if !rt.HasField("Y") {
			t.Error("Should have field Y")
		}
		if rt.HasField("Z") {
			t.Error("Should not have field Z")
		}
		if rt.GetFieldType("X") != INTEGER {
			t.Error("Field X should be Integer")
		}
	})

	t.Run("Anonymous record", func(t *testing.T) {
		fields := map[string]Type{
			"Name": STRING,
			"Age":  INTEGER,
		}
		rt := NewRecordType("", fields)
		str := rt.String()
		// String should contain both fields
		if !(contains(str, "Name") && contains(str, "Age")) {
			t.Errorf("String() = %v, should contain Name and Age", str)
		}
	})
}

func TestRecordTypeEquality(t *testing.T) {
	fields1 := map[string]Type{
		"X": INTEGER,
		"Y": INTEGER,
	}
	fields2 := map[string]Type{
		"X": INTEGER,
		"Y": INTEGER,
	}
	fields3 := map[string]Type{
		"X": INTEGER,
		"Y": FLOAT,
	}

	tests := []struct {
		name     string
		a, b     Type
		expected bool
	}{
		{
			"Same named records",
			NewRecordType("TPoint", fields1),
			NewRecordType("TPoint", fields2),
			true,
		},
		{
			"Different named records",
			NewRecordType("TPoint", fields1),
			NewRecordType("TVector", fields1),
			false,
		},
		{
			"Same anonymous records",
			NewRecordType("", fields1),
			NewRecordType("", fields2),
			true,
		},
		{
			"Different field types",
			NewRecordType("", fields1),
			NewRecordType("", fields3),
			false,
		},
		{
			"Record vs non-record",
			NewRecordType("TPoint", fields1),
			INTEGER,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.a.Equals(tt.b); result != tt.expected {
				t.Errorf("Equals() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Compatibility Tests
// ============================================================================

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		name     string
		from, to Type
		expected bool
	}{
		{"Integer to Integer", INTEGER, INTEGER, true},
		{"Float to Float", FLOAT, FLOAT, true},
		{"Integer to Float", INTEGER, FLOAT, true},
		{"Float to Integer", FLOAT, INTEGER, false},
		{"String to String", STRING, STRING, true},
		{"String to Integer", STRING, INTEGER, false},
		{"Nil to Nil", NIL, NIL, true},
		{"Integer to String", INTEGER, STRING, false},
		{"Boolean to Boolean", BOOLEAN, BOOLEAN, true},
		{
			"Dynamic arrays same element",
			NewDynamicArrayType(INTEGER),
			NewDynamicArrayType(INTEGER),
			true,
		},
		{
			"Dynamic arrays different element",
			NewDynamicArrayType(INTEGER),
			NewDynamicArrayType(FLOAT),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := IsCompatible(tt.from, tt.to); result != tt.expected {
				t.Errorf("IsCompatible(%v, %v) = %v, want %v",
					tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestCanCoerce(t *testing.T) {
	tests := []struct {
		name     string
		from, to Type
		expected bool
	}{
		{"Integer to Integer (no coercion)", INTEGER, INTEGER, true},
		{"Integer to Float (coerce)", INTEGER, FLOAT, true},
		{"Float to Integer (no coerce)", FLOAT, INTEGER, false},
		{"String to Integer (no coerce)", STRING, INTEGER, false},
		{"Float to Float (no coercion)", FLOAT, FLOAT, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := CanCoerce(tt.from, tt.to); result != tt.expected {
				t.Errorf("CanCoerce(%v, %v) = %v, want %v",
					tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestNeedsCoercion(t *testing.T) {
	tests := []struct {
		name     string
		from, to Type
		expected bool
	}{
		{"Integer to Integer", INTEGER, INTEGER, false},
		{"Integer to Float", INTEGER, FLOAT, true},
		{"Float to Integer", FLOAT, INTEGER, false},
		{"Float to Float", FLOAT, FLOAT, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := NeedsCoercion(tt.from, tt.to); result != tt.expected {
				t.Errorf("NeedsCoercion(%v, %v) = %v, want %v",
					tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestPromoteTypes(t *testing.T) {
	tests := []struct {
		name        string
		left, right Type
		expected    Type
		expectsNil  bool
	}{
		{"Integer + Integer", INTEGER, INTEGER, INTEGER, false},
		{"Float + Float", FLOAT, FLOAT, FLOAT, false},
		{"Integer + Float", INTEGER, FLOAT, FLOAT, false},
		{"Float + Integer", FLOAT, INTEGER, FLOAT, false},
		{"String + String", STRING, STRING, STRING, false},
		{"Boolean + Boolean", BOOLEAN, BOOLEAN, BOOLEAN, false},
		{"Integer + String", INTEGER, STRING, nil, true},
		{"Boolean + Integer", BOOLEAN, INTEGER, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PromoteTypes(tt.left, tt.right)
			if tt.expectsNil {
				if result != nil {
					t.Errorf("PromoteTypes() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Error("PromoteTypes() = nil, want non-nil")
				} else if !result.Equals(tt.expected) {
					t.Errorf("PromoteTypes() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestIsComparableType(t *testing.T) {
	tests := []struct {
		name     string
		typ      Type
		expected bool
	}{
		{"Integer", INTEGER, true},
		{"Float", FLOAT, true},
		{"String", STRING, true},
		{"Boolean", BOOLEAN, true},
		{"Nil", NIL, true},
		{"Void", VOID, false},
		{"Function", NewFunctionType([]Type{}, INTEGER), false},
		{"Array", NewDynamicArrayType(INTEGER), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := IsComparableType(tt.typ); result != tt.expected {
				t.Errorf("IsComparableType(%v) = %v, want %v",
					tt.typ, result, tt.expected)
			}
		})
	}
}

func TestIsOrderedType(t *testing.T) {
	tests := []struct {
		name     string
		typ      Type
		expected bool
	}{
		{"Integer", INTEGER, true},
		{"Float", FLOAT, true},
		{"String", STRING, true},
		{"Boolean", BOOLEAN, false},
		{"Nil", NIL, false},
		{"Void", VOID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := IsOrderedType(tt.typ); result != tt.expected {
				t.Errorf("IsOrderedType(%v) = %v, want %v",
					tt.typ, result, tt.expected)
			}
		})
	}
}

func TestSupportsOperation(t *testing.T) {
	tests := []struct {
		name      string
		typ       Type
		operation string
		expected  bool
	}{
		{"Integer + ", INTEGER, "+", true},
		{"Integer - ", INTEGER, "-", true},
		{"Integer * ", INTEGER, "*", true},
		{"Integer / ", INTEGER, "/", true},
		{"Integer div", INTEGER, "div", true},
		{"Integer mod", INTEGER, "mod", true},
		{"Float + ", FLOAT, "+", true},
		{"Float div", FLOAT, "div", false},
		{"String + ", STRING, "+", true},
		{"String - ", STRING, "-", false},
		{"Boolean and", BOOLEAN, "and", true},
		{"Boolean or", BOOLEAN, "or", true},
		{"Boolean not", BOOLEAN, "not", true},
		{"Integer and", INTEGER, "and", false},
		{"Integer < ", INTEGER, "<", true},
		{"String < ", STRING, "<", true},
		{"Boolean < ", BOOLEAN, "<", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := SupportsOperation(tt.typ, tt.operation); result != tt.expected {
				t.Errorf("SupportsOperation(%v, %v) = %v, want %v",
					tt.typ, tt.operation, result, tt.expected)
			}
		})
	}
}

func TestIsValidType(t *testing.T) {
	tests := []struct {
		name     string
		typ      Type
		expected bool
	}{
		{"Integer", INTEGER, true},
		{"Float", FLOAT, true},
		{"String", STRING, true},
		{"Boolean", BOOLEAN, true},
		{"Nil", NIL, true},
		{"Void", VOID, true},
		{"Valid function", NewFunctionType([]Type{INTEGER}, STRING), true},
		{"Valid array", NewDynamicArrayType(INTEGER), true},
		{"Valid record", NewRecordType("", map[string]Type{"X": INTEGER}), true},
		{"Nil type", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := IsValidType(tt.typ); result != tt.expected {
				t.Errorf("IsValidType(%v) = %v, want %v",
					tt.typ, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

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
		{"Variant", VARIANT, "Variant", "VARIANT"}, // Task 9.222
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
		a        Type
		b        Type
		name     string
		expected bool
	}{
		{a: INTEGER, b: INTEGER, name: "Integer equals Integer", expected: true},
		{a: FLOAT, b: FLOAT, name: "Float equals Float", expected: true},
		{a: STRING, b: STRING, name: "String equals String", expected: true},
		{a: BOOLEAN, b: BOOLEAN, name: "Boolean equals Boolean", expected: true},
		{a: NIL, b: NIL, name: "Nil equals Nil", expected: true},
		{a: VOID, b: VOID, name: "Void equals Void", expected: true},
		{a: VARIANT, b: VARIANT, name: "Variant equals Variant", expected: true}, // Task 9.222
		{a: INTEGER, b: FLOAT, name: "Integer not equals Float", expected: false},
		{a: STRING, b: BOOLEAN, name: "String not equals Boolean", expected: false},
		{a: INTEGER, b: STRING, name: "Integer not equals String", expected: false},
		{a: VARIANT, b: INTEGER, name: "Variant not equals Integer", expected: false}, // Task 9.222
		{a: VARIANT, b: STRING, name: "Variant not equals String", expected: false},   // Task 9.222
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
		expected Type
		name     string
		input    string
		wantErr  bool
	}{
		{expected: INTEGER, name: "Integer", input: "Integer", wantErr: false},
		{expected: FLOAT, name: "Float", input: "Float", wantErr: false},
		{expected: STRING, name: "String", input: "String", wantErr: false},
		{expected: BOOLEAN, name: "Boolean", input: "Boolean", wantErr: false},
		{expected: VOID, name: "Void", input: "Void", wantErr: false},
		{expected: VARIANT, name: "Variant", input: "Variant", wantErr: false}, // Task 9.222
		{expected: nil, name: "Unknown", input: "Unknown", wantErr: true},
		{expected: nil, name: "Empty", input: "", wantErr: true},
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
// Compatibility Tests
// ============================================================================

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		from     Type
		to       Type
		name     string
		expected bool
	}{
		{from: INTEGER, to: INTEGER, name: "Integer to Integer", expected: true},
		{from: FLOAT, to: FLOAT, name: "Float to Float", expected: true},
		{from: INTEGER, to: FLOAT, name: "Integer to Float", expected: true},
		{from: FLOAT, to: INTEGER, name: "Float to Integer", expected: false},
		{from: STRING, to: STRING, name: "String to String", expected: true},
		{from: STRING, to: INTEGER, name: "String to Integer", expected: false},
		{from: NIL, to: NIL, name: "Nil to Nil", expected: true},
		{from: INTEGER, to: STRING, name: "Integer to String", expected: false},
		{from: BOOLEAN, to: BOOLEAN, name: "Boolean to Boolean", expected: true},
		{
			from:     NewDynamicArrayType(INTEGER),
			to:       NewDynamicArrayType(INTEGER),
			name:     "Dynamic arrays same element",
			expected: true,
		},
		{
			from:     NewDynamicArrayType(INTEGER),
			to:       NewDynamicArrayType(FLOAT),
			name:     "Dynamic arrays different element",
			expected: false,
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
		from     Type
		to       Type
		name     string
		expected bool
	}{
		{from: INTEGER, to: INTEGER, name: "Integer to Integer (no coercion)", expected: true},
		{from: INTEGER, to: FLOAT, name: "Integer to Float (coerce)", expected: true},
		{from: FLOAT, to: INTEGER, name: "Float to Integer (no coerce)", expected: false},
		{from: STRING, to: INTEGER, name: "String to Integer (no coerce)", expected: false},
		{from: FLOAT, to: FLOAT, name: "Float to Float (no coercion)", expected: true},
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
		from     Type
		to       Type
		name     string
		expected bool
	}{
		{from: INTEGER, to: INTEGER, name: "Integer to Integer", expected: false},
		{from: INTEGER, to: FLOAT, name: "Integer to Float", expected: true},
		{from: FLOAT, to: INTEGER, name: "Float to Integer", expected: false},
		{from: FLOAT, to: FLOAT, name: "Float to Float", expected: false},
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
		left       Type
		right      Type
		expected   Type
		name       string
		expectsNil bool
	}{
		{left: INTEGER, right: INTEGER, expected: INTEGER, name: "Integer + Integer", expectsNil: false},
		{left: FLOAT, right: FLOAT, expected: FLOAT, name: "Float + Float", expectsNil: false},
		{left: INTEGER, right: FLOAT, expected: FLOAT, name: "Integer + Float", expectsNil: false},
		{left: FLOAT, right: INTEGER, expected: FLOAT, name: "Float + Integer", expectsNil: false},
		{left: STRING, right: STRING, expected: STRING, name: "String + String", expectsNil: false},
		{left: BOOLEAN, right: BOOLEAN, expected: BOOLEAN, name: "Boolean + Boolean", expectsNil: false},
		{left: INTEGER, right: STRING, expected: nil, name: "Integer + String", expectsNil: true},
		{left: BOOLEAN, right: INTEGER, expected: nil, name: "Boolean + Integer", expectsNil: true},
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
		typ      Type
		name     string
		expected bool
	}{
		{typ: INTEGER, name: "Integer", expected: true},
		{typ: FLOAT, name: "Float", expected: true},
		{typ: STRING, name: "String", expected: true},
		{typ: BOOLEAN, name: "Boolean", expected: true},
		{typ: NIL, name: "Nil", expected: true},
		{typ: VOID, name: "Void", expected: false},
		{typ: NewFunctionType([]Type{}, INTEGER), name: "Function", expected: false},
		{typ: NewDynamicArrayType(INTEGER), name: "Array", expected: false},
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
		typ      Type
		name     string
		expected bool
	}{
		{typ: INTEGER, name: "Integer", expected: true},
		{typ: FLOAT, name: "Float", expected: true},
		{typ: STRING, name: "String", expected: true},
		{typ: BOOLEAN, name: "Boolean", expected: false},
		{typ: NIL, name: "Nil", expected: false},
		{typ: VOID, name: "Void", expected: false},
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
		typ      Type
		name     string
		expected bool
	}{
		{typ: INTEGER, name: "Integer", expected: true},
		{typ: FLOAT, name: "Float", expected: true},
		{typ: STRING, name: "String", expected: true},
		{typ: BOOLEAN, name: "Boolean", expected: true},
		{typ: NIL, name: "Nil", expected: true},
		{typ: VOID, name: "Void", expected: true},
		{typ: NewFunctionType([]Type{INTEGER}, STRING), name: "Valid function", expected: true},
		{typ: NewDynamicArrayType(INTEGER), name: "Valid array", expected: true},
		{typ: NewRecordType("", map[string]Type{"X": INTEGER}), name: "Valid record", expected: true},
		{typ: nil, name: "Nil type", expected: false},
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

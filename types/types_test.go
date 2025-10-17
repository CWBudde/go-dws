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

// ============================================================================
// ClassType Tests (Stage 7.1-7.2)
// ============================================================================

func TestNewClassType(t *testing.T) {
	class := NewClassType("TPoint", nil)

	if class.Name != "TPoint" {
		t.Errorf("Expected class name 'TPoint', got '%s'", class.Name)
	}

	if class.Parent != nil {
		t.Errorf("Expected nil parent, got %v", class.Parent)
	}

	if class.Fields == nil {
		t.Error("Expected Fields map to be initialized")
	}

	if class.Methods == nil {
		t.Error("Expected Methods map to be initialized")
	}
}

func TestClassTypeString(t *testing.T) {
	tests := []struct {
		name     string
		class    *ClassType
		expected string
	}{
		{
			name:     "root class",
			class:    NewClassType("TObject", nil),
			expected: "TObject",
		},
		{
			name:     "derived class",
			class:    NewClassType("TPerson", NewClassType("TObject", nil)),
			expected: "TPerson(TObject)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.class.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestClassTypeKind(t *testing.T) {
	class := NewClassType("TPoint", nil)
	if class.TypeKind() != "CLASS" {
		t.Errorf("Expected TypeKind 'CLASS', got '%s'", class.TypeKind())
	}
}

func TestClassTypeEquals(t *testing.T) {
	class1 := NewClassType("TPoint", nil)
	class2 := NewClassType("TPoint", nil)
	class3 := NewClassType("TPerson", nil)

	tests := []struct {
		name     string
		class1   Type
		class2   Type
		expected bool
	}{
		{
			name:     "same class name",
			class1:   class1,
			class2:   class2,
			expected: true,
		},
		{
			name:     "different class names",
			class1:   class1,
			class2:   class3,
			expected: false,
		},
		{
			name:     "class vs non-class",
			class1:   class1,
			class2:   INTEGER,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.class1.Equals(tt.class2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestClassTypeFields(t *testing.T) {
	parent := NewClassType("TObject", nil)
	parent.Fields["ID"] = INTEGER

	child := NewClassType("TPerson", parent)
	child.Fields["Name"] = STRING
	child.Fields["Age"] = INTEGER

	tests := []struct {
		name      string
		class     *ClassType
		fieldName string
		hasField  bool
		fieldType Type
	}{
		{
			name:      "own field",
			class:     child,
			fieldName: "Name",
			hasField:  true,
			fieldType: STRING,
		},
		{
			name:      "inherited field",
			class:     child,
			fieldName: "ID",
			hasField:  true,
			fieldType: INTEGER,
		},
		{
			name:      "non-existent field",
			class:     child,
			fieldName: "Email",
			hasField:  false,
			fieldType: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test HasField
			hasField := tt.class.HasField(tt.fieldName)
			if hasField != tt.hasField {
				t.Errorf("HasField: expected %v, got %v", tt.hasField, hasField)
			}

			// Test GetField
			fieldType, found := tt.class.GetField(tt.fieldName)
			if found != tt.hasField {
				t.Errorf("GetField found: expected %v, got %v", tt.hasField, found)
			}

			if tt.hasField && !fieldType.Equals(tt.fieldType) {
				t.Errorf("GetField type: expected %v, got %v", tt.fieldType, fieldType)
			}
		})
	}
}

func TestClassTypeMethods(t *testing.T) {
	parent := NewClassType("TObject", nil)
	parent.Methods["ToString"] = NewFunctionType([]Type{}, STRING)

	child := NewClassType("TPerson", parent)
	child.Methods["GetAge"] = NewFunctionType([]Type{}, INTEGER)
	child.Methods["SetName"] = NewProcedureType([]Type{STRING})

	tests := []struct {
		name       string
		class      *ClassType
		methodName string
		hasMethod  bool
		isFunction bool
	}{
		{
			name:       "own method (function)",
			class:      child,
			methodName: "GetAge",
			hasMethod:  true,
			isFunction: true,
		},
		{
			name:       "own method (procedure)",
			class:      child,
			methodName: "SetName",
			hasMethod:  true,
			isFunction: false,
		},
		{
			name:       "inherited method",
			class:      child,
			methodName: "ToString",
			hasMethod:  true,
			isFunction: true,
		},
		{
			name:       "non-existent method",
			class:      child,
			methodName: "DoSomething",
			hasMethod:  false,
			isFunction: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test HasMethod
			hasMethod := tt.class.HasMethod(tt.methodName)
			if hasMethod != tt.hasMethod {
				t.Errorf("HasMethod: expected %v, got %v", tt.hasMethod, hasMethod)
			}

			// Test GetMethod
			methodType, found := tt.class.GetMethod(tt.methodName)
			if found != tt.hasMethod {
				t.Errorf("GetMethod found: expected %v, got %v", tt.hasMethod, found)
			}

			if tt.hasMethod {
				if methodType.IsFunction() != tt.isFunction {
					t.Errorf("IsFunction: expected %v, got %v", tt.isFunction, methodType.IsFunction())
				}
			}
		})
	}
}

// ============================================================================
// InterfaceType Tests (Stage 7.3)
// ============================================================================

func TestNewInterfaceType(t *testing.T) {
	iface := NewInterfaceType("IComparable")

	if iface.Name != "IComparable" {
		t.Errorf("Expected interface name 'IComparable', got '%s'", iface.Name)
	}

	if iface.Methods == nil {
		t.Error("Expected Methods map to be initialized")
	}
}

func TestInterfaceTypeString(t *testing.T) {
	iface := NewInterfaceType("IComparable")
	if iface.String() != "IComparable" {
		t.Errorf("Expected 'IComparable', got '%s'", iface.String())
	}
}

func TestInterfaceTypeKind(t *testing.T) {
	iface := NewInterfaceType("IComparable")
	if iface.TypeKind() != "INTERFACE" {
		t.Errorf("Expected TypeKind 'INTERFACE', got '%s'", iface.TypeKind())
	}
}

func TestInterfaceTypeEquals(t *testing.T) {
	iface1 := NewInterfaceType("IComparable")
	iface2 := NewInterfaceType("IComparable")
	iface3 := NewInterfaceType("IEnumerable")

	tests := []struct {
		name     string
		iface1   Type
		iface2   Type
		expected bool
	}{
		{
			name:     "same interface name",
			iface1:   iface1,
			iface2:   iface2,
			expected: true,
		},
		{
			name:     "different interface names",
			iface1:   iface1,
			iface2:   iface3,
			expected: false,
		},
		{
			name:     "interface vs non-interface",
			iface1:   iface1,
			iface2:   STRING,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.iface1.Equals(tt.iface2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestInterfaceTypeMethods(t *testing.T) {
	iface := NewInterfaceType("IComparable")
	iface.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)
	iface.Methods["Equals"] = NewFunctionType([]Type{INTEGER}, BOOLEAN)

	tests := []struct {
		name       string
		methodName string
		hasMethod  bool
	}{
		{
			name:       "existing method",
			methodName: "CompareTo",
			hasMethod:  true,
		},
		{
			name:       "non-existent method",
			methodName: "DoSomething",
			hasMethod:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test HasMethod
			hasMethod := iface.HasMethod(tt.methodName)
			if hasMethod != tt.hasMethod {
				t.Errorf("HasMethod: expected %v, got %v", tt.hasMethod, hasMethod)
			}

			// Test GetMethod
			_, found := iface.GetMethod(tt.methodName)
			if found != tt.hasMethod {
				t.Errorf("GetMethod found: expected %v, got %v", tt.hasMethod, found)
			}
		})
	}
}

// ============================================================================
// Type Compatibility Tests (Stage 7.4)
// ============================================================================

func TestIsSubclassOf(t *testing.T) {
	// Create class hierarchy: TObject -> TPerson -> TEmployee
	tObject := NewClassType("TObject", nil)
	tPerson := NewClassType("TPerson", tObject)
	tEmployee := NewClassType("TEmployee", tPerson)
	tUnrelated := NewClassType("TPoint", nil)

	tests := []struct {
		name     string
		child    *ClassType
		parent   *ClassType
		expected bool
	}{
		{
			name:     "direct parent",
			child:    tPerson,
			parent:   tObject,
			expected: true,
		},
		{
			name:     "grandparent",
			child:    tEmployee,
			parent:   tObject,
			expected: true,
		},
		{
			name:     "immediate parent",
			child:    tEmployee,
			parent:   tPerson,
			expected: true,
		},
		{
			name:     "same class",
			child:    tPerson,
			parent:   tPerson,
			expected: true,
		},
		{
			name:     "unrelated classes",
			child:    tPerson,
			parent:   tUnrelated,
			expected: false,
		},
		{
			name:     "reverse hierarchy",
			child:    tObject,
			parent:   tPerson,
			expected: false,
		},
		{
			name:     "nil child",
			child:    nil,
			parent:   tObject,
			expected: false,
		},
		{
			name:     "nil parent",
			child:    tPerson,
			parent:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSubclassOf(tt.child, tt.parent)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsAssignableFrom(t *testing.T) {
	// Create class hierarchy
	tObject := NewClassType("TObject", nil)
	tPerson := NewClassType("TPerson", tObject)
	tEmployee := NewClassType("TEmployee", tPerson)

	// Create interface
	iComparable := NewInterfaceType("IComparable")
	iComparable.Methods["CompareTo"] = NewFunctionType([]Type{}, INTEGER)

	// Make TPerson implement IComparable
	tPerson.Methods["CompareTo"] = NewFunctionType([]Type{}, INTEGER)

	tests := []struct {
		name     string
		target   Type
		source   Type
		expected bool
	}{
		{
			name:     "exact type match",
			target:   INTEGER,
			source:   INTEGER,
			expected: true,
		},
		{
			name:     "integer to float coercion",
			target:   FLOAT,
			source:   INTEGER,
			expected: true,
		},
		{
			name:     "float to integer (no coercion)",
			target:   INTEGER,
			source:   FLOAT,
			expected: false,
		},
		{
			name:     "subclass to superclass",
			target:   tObject,
			source:   tPerson,
			expected: true,
		},
		{
			name:     "grandchild to grandparent",
			target:   tObject,
			source:   tEmployee,
			expected: true,
		},
		{
			name:     "superclass to subclass (not allowed)",
			target:   tPerson,
			source:   tObject,
			expected: false,
		},
		{
			name:     "class to implemented interface",
			target:   iComparable,
			source:   tPerson,
			expected: true,
		},
		{
			name:     "class to non-implemented interface",
			target:   iComparable,
			source:   tObject,
			expected: false,
		},
		{
			name:     "incompatible types",
			target:   STRING,
			source:   INTEGER,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAssignableFrom(tt.target, tt.source)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Interface Implementation Tests (Stage 7.5)
// ============================================================================

func TestImplementsInterface(t *testing.T) {
	// Create interface with two methods
	iComparable := NewInterfaceType("IComparable")
	iComparable.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)
	iComparable.Methods["Equals"] = NewFunctionType([]Type{INTEGER}, BOOLEAN)

	// Create interface with one method
	iSimple := NewInterfaceType("ISimple")
	iSimple.Methods["DoSomething"] = NewProcedureType([]Type{})

	// Create class that implements IComparable fully
	tFullImpl := NewClassType("TFullImpl", nil)
	tFullImpl.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)
	tFullImpl.Methods["Equals"] = NewFunctionType([]Type{INTEGER}, BOOLEAN)

	// Create class that partially implements IComparable
	tPartialImpl := NewClassType("TPartialImpl", nil)
	tPartialImpl.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)
	// Missing Equals method

	// Create class with wrong signature
	tWrongSig := NewClassType("TWrongSig", nil)
	tWrongSig.Methods["CompareTo"] = NewFunctionType([]Type{STRING}, INTEGER) // Wrong param type
	tWrongSig.Methods["Equals"] = NewFunctionType([]Type{INTEGER}, BOOLEAN)

	// Create class that implements via inheritance
	tParent := NewClassType("TParent", nil)
	tParent.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)
	tChild := NewClassType("TChild", tParent)
	tChild.Methods["Equals"] = NewFunctionType([]Type{INTEGER}, BOOLEAN)

	tests := []struct {
		name     string
		class    *ClassType
		iface    *InterfaceType
		expected bool
	}{
		{
			name:     "full implementation",
			class:    tFullImpl,
			iface:    iComparable,
			expected: true,
		},
		{
			name:     "partial implementation",
			class:    tPartialImpl,
			iface:    iComparable,
			expected: false,
		},
		{
			name:     "wrong signature",
			class:    tWrongSig,
			iface:    iComparable,
			expected: false,
		},
		{
			name:     "implementation via inheritance",
			class:    tChild,
			iface:    iComparable,
			expected: true,
		},
		{
			name:     "unrelated interface",
			class:    tFullImpl,
			iface:    iSimple,
			expected: false,
		},
		{
			name:     "nil class",
			class:    nil,
			iface:    iComparable,
			expected: false,
		},
		{
			name:     "nil interface",
			class:    tFullImpl,
			iface:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ImplementsInterface(tt.class, tt.iface)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Type Checking Utility Tests
// ============================================================================

func TestIsClassType(t *testing.T) {
	class := NewClassType("TPoint", nil)

	tests := []struct {
		name     string
		typ      Type
		expected bool
	}{
		{
			name:     "class type",
			typ:      class,
			expected: true,
		},
		{
			name:     "integer type",
			typ:      INTEGER,
			expected: false,
		},
		{
			name:     "function type",
			typ:      NewFunctionType([]Type{INTEGER}, STRING),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsClassType(tt.typ)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsInterfaceType(t *testing.T) {
	iface := NewInterfaceType("IComparable")

	tests := []struct {
		name     string
		typ      Type
		expected bool
	}{
		{
			name:     "interface type",
			typ:      iface,
			expected: true,
		},
		{
			name:     "string type",
			typ:      STRING,
			expected: false,
		},
		{
			name:     "class type",
			typ:      NewClassType("TPoint", nil),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInterfaceType(tt.typ)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Complex Hierarchy Tests
// ============================================================================

func TestComplexClassHierarchy(t *testing.T) {
	// Build a more complex hierarchy:
	// TObject
	//   ├─ TStream
	//   │    ├─ TFileStream
	//   │    └─ TMemoryStream
	//   └─ TPersistent
	//        └─ TComponent

	tObject := NewClassType("TObject", nil)
	tObject.Methods["ToString"] = NewFunctionType([]Type{}, STRING)

	tStream := NewClassType("TStream", tObject)
	tStream.Fields["Size"] = INTEGER
	tStream.Methods["Read"] = NewFunctionType([]Type{INTEGER}, INTEGER)

	tFileStream := NewClassType("TFileStream", tStream)
	tFileStream.Fields["FileName"] = STRING

	tMemoryStream := NewClassType("TMemoryStream", tStream)
	tMemoryStream.Fields["Memory"] = INTEGER

	tPersistent := NewClassType("TPersistent", tObject)
	tPersistent.Methods["Assign"] = NewProcedureType([]Type{})

	tComponent := NewClassType("TComponent", tPersistent)
	tComponent.Fields["Name"] = STRING

	// Test field inheritance through multiple levels
	t.Run("field inheritance", func(t *testing.T) {
		if !tFileStream.HasField("Size") {
			t.Error("TFileStream should have Size field from TStream")
		}
		if !tComponent.HasField("Name") {
			t.Error("TComponent should have its own Name field")
		}
	})

	// Test method inheritance through multiple levels
	t.Run("method inheritance", func(t *testing.T) {
		if !tFileStream.HasMethod("ToString") {
			t.Error("TFileStream should have ToString method from TObject")
		}
		if !tComponent.HasMethod("ToString") {
			t.Error("TComponent should have ToString method from TObject")
		}
		if !tComponent.HasMethod("Assign") {
			t.Error("TComponent should have Assign method from TPersistent")
		}
	})

	// Test subclass relationships
	t.Run("subclass relationships", func(t *testing.T) {
		if !IsSubclassOf(tFileStream, tObject) {
			t.Error("TFileStream should be subclass of TObject")
		}
		if !IsSubclassOf(tComponent, tObject) {
			t.Error("TComponent should be subclass of TObject")
		}
		if IsSubclassOf(tFileStream, tComponent) {
			t.Error("TFileStream should not be subclass of TComponent")
		}
	})
}

func TestMultipleInterfaceImplementation(t *testing.T) {
	// Create multiple interfaces
	iReadable := NewInterfaceType("IReadable")
	iReadable.Methods["Read"] = NewFunctionType([]Type{}, STRING)

	iWritable := NewInterfaceType("IWritable")
	iWritable.Methods["Write"] = NewProcedureType([]Type{STRING})

	iCloseable := NewInterfaceType("ICloseable")
	iCloseable.Methods["Close"] = NewProcedureType([]Type{})

	// Create class that implements all three
	tFile := NewClassType("TFile", nil)
	tFile.Methods["Read"] = NewFunctionType([]Type{}, STRING)
	tFile.Methods["Write"] = NewProcedureType([]Type{STRING})
	tFile.Methods["Close"] = NewProcedureType([]Type{})

	// Test each interface
	tests := []struct {
		name     string
		iface    *InterfaceType
		expected bool
	}{
		{
			name:     "implements IReadable",
			iface:    iReadable,
			expected: true,
		},
		{
			name:     "implements IWritable",
			iface:    iWritable,
			expected: true,
		},
		{
			name:     "implements ICloseable",
			iface:    iCloseable,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ImplementsInterface(tFile, tt.iface)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

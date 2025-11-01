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
		a        Type
		b        Type
		name     string
		expected bool
	}{
		{
			a:        NewFunctionType([]Type{INTEGER, STRING}, BOOLEAN),
			b:        NewFunctionType([]Type{INTEGER, STRING}, BOOLEAN),
			name:     "Same function types",
			expected: true,
		},
		{
			a:        NewFunctionType([]Type{INTEGER}, BOOLEAN),
			b:        NewFunctionType([]Type{FLOAT}, BOOLEAN),
			name:     "Different parameter types",
			expected: false,
		},
		{
			a:        NewFunctionType([]Type{INTEGER}, BOOLEAN),
			b:        NewFunctionType([]Type{INTEGER}, INTEGER),
			name:     "Different return types",
			expected: false,
		},
		{
			a:        NewFunctionType([]Type{INTEGER}, BOOLEAN),
			b:        NewFunctionType([]Type{INTEGER, STRING}, BOOLEAN),
			name:     "Different parameter count",
			expected: false,
		},
		{
			a:        NewFunctionType([]Type{}, INTEGER),
			b:        INTEGER,
			name:     "Function vs non-function",
			expected: false,
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
		a        Type
		b        Type
		name     string
		expected bool
	}{
		{
			a:        NewDynamicArrayType(INTEGER),
			b:        NewDynamicArrayType(INTEGER),
			name:     "Same dynamic arrays",
			expected: true,
		},
		{
			a:        NewDynamicArrayType(INTEGER),
			b:        NewDynamicArrayType(FLOAT),
			name:     "Different element types (dynamic)",
			expected: false,
		},
		{
			a:        NewStaticArrayType(STRING, 1, 10),
			b:        NewStaticArrayType(STRING, 1, 10),
			name:     "Same static arrays",
			expected: true,
		},
		{
			a:        NewStaticArrayType(INTEGER, 1, 10),
			b:        NewStaticArrayType(INTEGER, 0, 9),
			name:     "Different bounds",
			expected: false,
		},
		{
			a:        NewDynamicArrayType(INTEGER),
			b:        NewStaticArrayType(INTEGER, 1, 10),
			name:     "Dynamic vs static",
			expected: false,
		},
		{
			a:        NewDynamicArrayType(INTEGER),
			b:        INTEGER,
			name:     "Array vs non-array",
			expected: false,
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
		if !contains(str, "Name") || !contains(str, "Age") {
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
		a        Type
		b        Type
		name     string
		expected bool
	}{
		{
			a:        NewRecordType("TPoint", fields1),
			b:        NewRecordType("TPoint", fields2),
			name:     "Same named records",
			expected: true,
		},
		{
			a:        NewRecordType("TPoint", fields1),
			b:        NewRecordType("TVector", fields1),
			name:     "Different named records",
			expected: false,
		},
		{
			a:        NewRecordType("", fields1),
			b:        NewRecordType("", fields2),
			name:     "Same anonymous records",
			expected: true,
		},
		{
			a:        NewRecordType("", fields1),
			b:        NewRecordType("", fields3),
			name:     "Different field types",
			expected: false,
		},
		{
			a:        NewRecordType("TPoint", fields1),
			b:        INTEGER,
			name:     "Record vs non-record",
			expected: false,
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

// Task 8.55: Test RecordType with methods and properties
func TestRecordTypeWithMethods(t *testing.T) {
	t.Run("Record with methods", func(t *testing.T) {
		fields := map[string]Type{
			"X": INTEGER,
			"Y": INTEGER,
		}
		rt := NewRecordType("TPoint", fields)

		// Add methods to the record
		rt.Methods = make(map[string]*FunctionType)
		rt.Methods["GetDistance"] = NewFunctionType([]Type{}, FLOAT)
		rt.Methods["SetPosition"] = NewProcedureType([]Type{INTEGER, INTEGER})

		// Test HasMethod
		if !rt.HasMethod("GetDistance") {
			t.Error("Should have method GetDistance")
		}
		if !rt.HasMethod("SetPosition") {
			t.Error("Should have method SetPosition")
		}
		if rt.HasMethod("NonExistent") {
			t.Error("Should not have method NonExistent")
		}

		// Test GetMethod
		method := rt.GetMethod("GetDistance")
		if method == nil {
			t.Error("GetMethod should return method type")
		}
		if method != nil && !method.IsFunction() {
			t.Error("GetDistance should be a function")
		}

		method = rt.GetMethod("SetPosition")
		if method == nil {
			t.Error("GetMethod should return method type")
		}
		if method != nil && !method.IsProcedure() {
			t.Error("SetPosition should be a procedure")
		}

		// Test GetMethod for non-existent method
		method = rt.GetMethod("NonExistent")
		if method != nil {
			t.Error("GetMethod should return nil for non-existent method")
		}
	})

	t.Run("Record without methods", func(t *testing.T) {
		fields := map[string]Type{
			"X": INTEGER,
			"Y": INTEGER,
		}
		rt := NewRecordType("TPoint", fields)

		// Methods map should be initialized but empty
		if rt.Methods == nil {
			t.Error("Methods map should be initialized")
		}
		if rt.HasMethod("AnyMethod") {
			t.Error("Empty record should not have any methods")
		}
	})
}

func TestRecordTypeWithProperties(t *testing.T) {
	t.Run("Record with properties", func(t *testing.T) {
		fields := map[string]Type{
			"FX": INTEGER,
			"FY": INTEGER,
		}
		rt := NewRecordType("TPoint", fields)

		// Add properties to the record
		rt.Properties = make(map[string]*RecordPropertyInfo)
		rt.Properties["X"] = &RecordPropertyInfo{
			Name:       "X",
			Type:       INTEGER,
			ReadField:  "FX",
			WriteField: "FX",
		}
		rt.Properties["Y"] = &RecordPropertyInfo{
			Name:       "Y",
			Type:       INTEGER,
			ReadField:  "FY",
			WriteField: "FY",
		}

		// Test HasProperty
		if !rt.HasProperty("X") {
			t.Error("Should have property X")
		}
		if !rt.HasProperty("Y") {
			t.Error("Should have property Y")
		}
		if rt.HasProperty("Z") {
			t.Error("Should not have property Z")
		}

		// Test GetProperty
		prop := rt.GetProperty("X")
		if prop == nil {
			t.Error("GetProperty should return property info")
		}
		if prop != nil && prop.Type != INTEGER {
			t.Error("Property X should be Integer type")
		}

		// Test GetProperty for non-existent property
		prop = rt.GetProperty("Z")
		if prop != nil {
			t.Error("GetProperty should return nil for non-existent property")
		}
	})

	t.Run("Record without properties", func(t *testing.T) {
		fields := map[string]Type{
			"X": INTEGER,
			"Y": INTEGER,
		}
		rt := NewRecordType("TPoint", fields)

		// Properties map should be initialized but empty
		if rt.Properties == nil {
			t.Error("Properties map should be initialized")
		}
		if rt.HasProperty("AnyProperty") {
			t.Error("Empty record should not have any properties")
		}
	})
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

// ============================================================================
// TypeAlias Tests
// ============================================================================

func TestTypeAlias(t *testing.T) {
	t.Run("Create type alias", func(t *testing.T) {
		// Task 9.14: Test creating type alias
		// type TUserID = Integer;
		userID := &TypeAlias{
			Name:        "TUserID",
			AliasedType: INTEGER,
		}

		// Test String() returns alias name
		if userID.String() != "TUserID" {
			t.Errorf("String() = %v, want TUserID", userID.String())
		}

		// Test TypeKind() returns underlying type's kind
		if userID.TypeKind() != "INTEGER" {
			t.Errorf("TypeKind() = %v, want INTEGER", userID.TypeKind())
		}

		// Test that the aliased type is stored correctly
		if userID.AliasedType != INTEGER {
			t.Error("AliasedType should be INTEGER")
		}
	})

	t.Run("Alias equality with underlying type", func(t *testing.T) {
		// Task 9.14: Test alias equality with underlying type
		// type MyInt = Integer;
		// MyInt should equal Integer
		myInt := &TypeAlias{
			Name:        "MyInt",
			AliasedType: INTEGER,
		}

		if !myInt.Equals(INTEGER) {
			t.Error("Type alias should equal its underlying type")
		}

		if !INTEGER.Equals(myInt) {
			t.Error("Underlying type should equal type alias (symmetric)")
		}
	})

	t.Run("Alias inequality with different types", func(t *testing.T) {
		// Task 9.14: Test alias inequality with different types
		// type MyInt = Integer;
		// MyInt should NOT equal String
		myInt := &TypeAlias{
			Name:        "MyInt",
			AliasedType: INTEGER,
		}

		if myInt.Equals(STRING) {
			t.Error("Integer alias should not equal STRING")
		}

		if myInt.Equals(FLOAT) {
			t.Error("Integer alias should not equal FLOAT")
		}

		if myInt.Equals(BOOLEAN) {
			t.Error("Integer alias should not equal BOOLEAN")
		}

		// type MyString = String;
		// MyString should NOT equal MyInt
		myString := &TypeAlias{
			Name:        "MyString",
			AliasedType: STRING,
		}

		if myInt.Equals(myString) {
			t.Error("Integer alias should not equal String alias")
		}
	})

	t.Run("Multiple aliases to same type are equal", func(t *testing.T) {
		// type TUserID = Integer;
		// type TItemID = Integer;
		// TUserID should equal TItemID (both alias Integer)
		userID := &TypeAlias{
			Name:        "TUserID",
			AliasedType: INTEGER,
		}

		itemID := &TypeAlias{
			Name:        "TItemID",
			AliasedType: INTEGER,
		}

		if !userID.Equals(itemID) {
			t.Error("Two aliases to the same type should be equal")
		}
	})

	t.Run("Nested aliases", func(t *testing.T) {
		// Task 9.14: Test nested aliases
		// type A = Integer;
		// type B = A;
		// B should behave like Integer
		aliasA := &TypeAlias{
			Name:        "A",
			AliasedType: INTEGER,
		}

		aliasB := &TypeAlias{
			Name:        "B",
			AliasedType: aliasA,
		}

		// B.String() should return "B"
		if aliasB.String() != "B" {
			t.Errorf("String() = %v, want B", aliasB.String())
		}

		// B.TypeKind() should return "INTEGER" (resolved through chain)
		if aliasB.TypeKind() != "INTEGER" {
			t.Errorf("TypeKind() = %v, want INTEGER", aliasB.TypeKind())
		}

		// B should equal Integer
		if !aliasB.Equals(INTEGER) {
			t.Error("Nested alias should equal ultimate underlying type (Integer)")
		}

		// B should equal A
		if !aliasB.Equals(aliasA) {
			t.Error("Nested alias should equal intermediate alias")
		}

		// A should equal B (symmetric)
		if !aliasA.Equals(aliasB) {
			t.Error("Alias equality should be symmetric")
		}
	})

	t.Run("Triple nested aliases", func(t *testing.T) {
		// type A = Integer;
		// type B = A;
		// type C = B;
		aliasA := &TypeAlias{
			Name:        "A",
			AliasedType: INTEGER,
		}

		aliasB := &TypeAlias{
			Name:        "B",
			AliasedType: aliasA,
		}

		aliasC := &TypeAlias{
			Name:        "C",
			AliasedType: aliasB,
		}

		// C.TypeKind() should resolve through entire chain
		if aliasC.TypeKind() != "INTEGER" {
			t.Errorf("TypeKind() = %v, want INTEGER", aliasC.TypeKind())
		}

		// C should equal Integer
		if !aliasC.Equals(INTEGER) {
			t.Error("Triple nested alias should equal ultimate underlying type")
		}

		// C should equal A and B
		if !aliasC.Equals(aliasA) {
			t.Error("Triple nested alias should equal first alias in chain")
		}

		if !aliasC.Equals(aliasB) {
			t.Error("Triple nested alias should equal second alias in chain")
		}
	})

	t.Run("Alias to different basic types", func(t *testing.T) {
		// Test aliases to various basic types
		tests := []struct {
			name           string
			aliasName      string
			underlyingType Type
			expectedKind   string
		}{
			{"Integer alias", "TUserID", INTEGER, "INTEGER"},
			{"Float alias", "TPrice", FLOAT, "FLOAT"},
			{"String alias", "TFileName", STRING, "STRING"},
			{"Boolean alias", "TFlag", BOOLEAN, "BOOLEAN"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				alias := &TypeAlias{
					Name:        tt.aliasName,
					AliasedType: tt.underlyingType,
				}

				if alias.String() != tt.aliasName {
					t.Errorf("String() = %v, want %v", alias.String(), tt.aliasName)
				}

				if alias.TypeKind() != tt.expectedKind {
					t.Errorf("TypeKind() = %v, want %v", alias.TypeKind(), tt.expectedKind)
				}

				if !alias.Equals(tt.underlyingType) {
					t.Errorf("Alias should equal underlying type %v", tt.underlyingType)
				}
			})
		}
	})

	t.Run("Alias to complex types", func(t *testing.T) {
		// type TIntArray = array of Integer;
		arrayType := NewDynamicArrayType(INTEGER)
		arrayAlias := &TypeAlias{
			Name:        "TIntArray",
			AliasedType: arrayType,
		}

		if arrayAlias.TypeKind() != "ARRAY" {
			t.Errorf("TypeKind() = %v, want ARRAY", arrayAlias.TypeKind())
		}

		if !arrayAlias.Equals(arrayType) {
			t.Error("Array alias should equal array type")
		}

		// type TMyFunc = function(): Integer;
		funcType := NewFunctionType([]Type{}, INTEGER)
		funcAlias := &TypeAlias{
			Name:        "TMyFunc",
			AliasedType: funcType,
		}

		if funcAlias.TypeKind() != "FUNCTION" {
			t.Errorf("TypeKind() = %v, want FUNCTION", funcAlias.TypeKind())
		}

		if !funcAlias.Equals(funcType) {
			t.Error("Function alias should equal function type")
		}
	})

	t.Run("Alias not equal to non-alias", func(t *testing.T) {
		// Ensure type aliases don't equal unrelated types
		myInt := &TypeAlias{
			Name:        "MyInt",
			AliasedType: INTEGER,
		}

		// Should not equal function type
		funcType := NewFunctionType([]Type{}, INTEGER)
		if myInt.Equals(funcType) {
			t.Error("Integer alias should not equal function type")
		}

		// Should not equal array type
		arrayType := NewDynamicArrayType(INTEGER)
		if myInt.Equals(arrayType) {
			t.Error("Integer alias should not equal array type")
		}

		// Should not equal class type
		classType := NewClassType("TMyClass", nil)
		if myInt.Equals(classType) {
			t.Error("Integer alias should not equal class type")
		}
	})
}

func TestGetUnderlyingType(t *testing.T) {
	t.Run("Non-alias returns itself", func(t *testing.T) {
		// GetUnderlyingType on a basic type should return the type itself
		if GetUnderlyingType(INTEGER) != INTEGER {
			t.Error("GetUnderlyingType(INTEGER) should return INTEGER")
		}

		if GetUnderlyingType(STRING) != STRING {
			t.Error("GetUnderlyingType(STRING) should return STRING")
		}

		// GetUnderlyingType on complex types should return the type itself
		arrayType := NewDynamicArrayType(INTEGER)
		if GetUnderlyingType(arrayType) != arrayType {
			t.Error("GetUnderlyingType should return array type itself")
		}
	})

	t.Run("Single alias resolves to underlying type", func(t *testing.T) {
		// type MyInt = Integer;
		myInt := &TypeAlias{
			Name:        "MyInt",
			AliasedType: INTEGER,
		}

		underlying := GetUnderlyingType(myInt)
		if underlying != INTEGER {
			t.Errorf("GetUnderlyingType(MyInt) = %v, want INTEGER", underlying)
		}
	})

	t.Run("Nested alias resolves to ultimate type", func(t *testing.T) {
		// type A = Integer;
		// type B = A;
		// type C = B;
		aliasA := &TypeAlias{
			Name:        "A",
			AliasedType: INTEGER,
		}

		aliasB := &TypeAlias{
			Name:        "B",
			AliasedType: aliasA,
		}

		aliasC := &TypeAlias{
			Name:        "C",
			AliasedType: aliasB,
		}

		// All should resolve to INTEGER
		if GetUnderlyingType(aliasA) != INTEGER {
			t.Error("GetUnderlyingType(A) should return INTEGER")
		}

		if GetUnderlyingType(aliasB) != INTEGER {
			t.Error("GetUnderlyingType(B) should return INTEGER")
		}

		if GetUnderlyingType(aliasC) != INTEGER {
			t.Error("GetUnderlyingType(C) should return INTEGER")
		}
	})

	t.Run("Alias to complex type resolves correctly", func(t *testing.T) {
		// type TIntArray = array of Integer;
		arrayType := NewDynamicArrayType(INTEGER)
		arrayAlias := &TypeAlias{
			Name:        "TIntArray",
			AliasedType: arrayType,
		}

		underlying := GetUnderlyingType(arrayAlias)
		if underlying != arrayType {
			t.Error("GetUnderlyingType should return array type")
		}

		// Nested alias to complex type
		// type TMyArray = TIntArray;
		nestedAlias := &TypeAlias{
			Name:        "TMyArray",
			AliasedType: arrayAlias,
		}

		underlying = GetUnderlyingType(nestedAlias)
		if underlying != arrayType {
			t.Error("GetUnderlyingType should resolve through alias chain to array type")
		}
	})

	// Task 9.222: Test Variant with type aliases
	t.Run("Variant type alias", func(t *testing.T) {
		// type MyVariant = Variant;
		myVariant := &TypeAlias{
			Name:        "MyVariant",
			AliasedType: VARIANT,
		}

		// Test String() returns alias name
		if myVariant.String() != "MyVariant" {
			t.Errorf("String() = %v, want MyVariant", myVariant.String())
		}

		// Test TypeKind() returns underlying type's kind
		if myVariant.TypeKind() != "VARIANT" {
			t.Errorf("TypeKind() = %v, want VARIANT", myVariant.TypeKind())
		}

		// Test equality with underlying Variant type
		if !myVariant.Equals(VARIANT) {
			t.Error("Variant alias should equal VARIANT")
		}

		if !VARIANT.Equals(myVariant) {
			t.Error("VARIANT should equal Variant alias (symmetric)")
		}

		// Test inequality with other types
		if myVariant.Equals(INTEGER) {
			t.Error("Variant alias should not equal INTEGER")
		}

		// Test GetUnderlyingType
		underlying := GetUnderlyingType(myVariant)
		if underlying != VARIANT {
			t.Error("GetUnderlyingType should resolve to VARIANT")
		}
	})
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
		class1   Type
		class2   Type
		name     string
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
		fieldType Type
		class     *ClassType
		name      string
		fieldName string
		hasField  bool
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
		iface1   Type
		iface2   Type
		name     string
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
		child    *ClassType
		parent   *ClassType
		name     string
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
		target   Type
		source   Type
		name     string
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
		class    *ClassType
		iface    *InterfaceType
		name     string
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
// SetType Tests
// ============================================================================

func TestSetType(t *testing.T) {
	// Create enum type for testing
	colorEnum := &EnumType{
		Name:         "TColor",
		Values:       map[string]int{"Red": 0, "Green": 1, "Blue": 2},
		OrderedNames: []string{"Red", "Green", "Blue"},
	}

	t.Run("Create basic set type", func(t *testing.T) {
		// Task 8.81: Create SetType with NewSetType factory
		setType := NewSetType(colorEnum)

		// Test String() method - should be "set of TColor"
		expected := "set of TColor"
		if setType.String() != expected {
			t.Errorf("String() = %v, want %v", setType.String(), expected)
		}

		// Test TypeKind() method
		if setType.TypeKind() != "SET" {
			t.Errorf("TypeKind() = %v, want SET", setType.TypeKind())
		}

		// Test ElementType field
		if setType.ElementType != colorEnum {
			t.Error("ElementType should match the provided enum type")
		}
	})

	t.Run("Set type equality", func(t *testing.T) {
		// Task 8.81: Test Equals() method
		enum1 := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0, "Green": 1, "Blue": 2},
			OrderedNames: []string{"Red", "Green", "Blue"},
		}
		enum2 := &EnumType{
			Name:         "TSize",
			Values:       map[string]int{"Small": 0, "Medium": 1, "Large": 2},
			OrderedNames: []string{"Small", "Medium", "Large"},
		}

		set1 := NewSetType(enum1)
		set2 := NewSetType(enum1)
		set3 := NewSetType(enum2)

		// Same element type sets should be equal
		if !set1.Equals(set2) {
			t.Error("Sets with same element type should be equal")
		}

		// Different element type sets should not be equal
		if set1.Equals(set3) {
			t.Error("Sets with different element types should not be equal")
		}

		// Set should not equal other types
		if set1.Equals(INTEGER) {
			t.Error("SetType should not equal INTEGER")
		}

		// Set should not equal enum type
		if set1.Equals(enum1) {
			t.Error("SetType should not equal EnumType")
		}
	})

	t.Run("Set type with nil element type", func(t *testing.T) {
		// Task 8.83: Validation - ensure nil enum type is handled
		// For now, we allow it to be created (validation will be added in semantic analysis)
		setType := NewSetType(nil)
		if setType == nil {
			t.Error("NewSetType should not return nil even with nil element type")
		}
		if setType.ElementType != nil {
			t.Error("ElementType should be nil when created with nil")
		}
	})

	t.Run("Set type with different enum instances", func(t *testing.T) {
		// Even if values are same, different enum instances should work
		enum1 := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0},
			OrderedNames: []string{"Red"},
		}
		enum2 := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0},
			OrderedNames: []string{"Red"},
		}

		set1 := NewSetType(enum1)
		set2 := NewSetType(enum2)

		// Sets should be equal because enums have same name (nominal typing)
		if !set1.Equals(set2) {
			t.Error("Sets with same-named enums should be equal")
		}
	})

	t.Run("Set type string representation", func(t *testing.T) {
		// Test various enum types to ensure String() works correctly
		tests := []struct {
			name     string
			enum     *EnumType
			expected string
		}{
			{
				name: "simple enum",
				enum: &EnumType{
					Name:         "TStatus",
					Values:       map[string]int{"Ok": 0, "Error": 1},
					OrderedNames: []string{"Ok", "Error"},
				},
				expected: "set of TStatus",
			},
			{
				name: "weekday enum",
				enum: &EnumType{
					Name:         "TWeekday",
					Values:       map[string]int{"Mon": 0, "Tue": 1, "Wed": 2},
					OrderedNames: []string{"Mon", "Tue", "Wed"},
				},
				expected: "set of TWeekday",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				setType := NewSetType(tt.enum)
				if setType.String() != tt.expected {
					t.Errorf("String() = %v, want %v", setType.String(), tt.expected)
				}
			})
		}
	})
}

// ============================================================================
// Type Checking Utility Tests
// ============================================================================

func TestIsClassType(t *testing.T) {
	class := NewClassType("TPoint", nil)

	tests := []struct {
		typ      Type
		name     string
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
		typ      Type
		name     string
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
// Interface Inheritance Tests
// ============================================================================

// Task 7.76: Test Equals() with hierarchy support for interfaces
func TestInterfaceEqualsWithHierarchy(t *testing.T) {
	// Create interface hierarchy
	iBase := NewInterfaceType("IBase")
	iDerived := &InterfaceType{
		Name:    "IDerived",
		Parent:  iBase,
		Methods: make(map[string]*FunctionType),
	}

	tests := []struct {
		iface1   Type
		iface2   Type
		name     string
		note     string
		expected bool
	}{
		{
			name:     "exact same interface",
			iface1:   iBase,
			iface2:   iBase,
			expected: true,
			note:     "same interface instance should equal itself",
		},
		{
			name:     "same name different instance",
			iface1:   NewInterfaceType("ITest"),
			iface2:   NewInterfaceType("ITest"),
			expected: true,
			note:     "interfaces with same name should be equal (nominal typing)",
		},
		{
			name:     "different interface names",
			iface1:   iBase,
			iface2:   iDerived,
			expected: false,
			note:     "derived interface is NOT equal to base (exact match required)",
		},
		{
			name:     "interface vs non-interface",
			iface1:   iBase,
			iface2:   INTEGER,
			expected: false,
			note:     "interface should not equal non-interface type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.iface1.Equals(tt.iface2)
			if result != tt.expected {
				t.Errorf("%s: Expected %v, got %v", tt.note, tt.expected, result)
			}
		})
	}
}

// Task 7.75: Test IINTERFACE base interface constant
func TestIINTERFACEConstant(t *testing.T) {
	t.Run("IINTERFACE exists", func(t *testing.T) {
		if IINTERFACE == nil {
			t.Error("IINTERFACE should be defined")
		}
		if IINTERFACE.Name != "IInterface" {
			t.Errorf("Expected IINTERFACE name to be 'IInterface', got '%s'", IINTERFACE.Name)
		}
		if IINTERFACE.Parent != nil {
			t.Error("IINTERFACE should have no parent (root interface)")
		}
	})

	t.Run("all interfaces can inherit from IINTERFACE", func(t *testing.T) {
		iCustom := &InterfaceType{
			Name:    "ICustom",
			Parent:  IINTERFACE,
			Methods: make(map[string]*FunctionType),
		}

		if !IsSubinterfaceOf(iCustom, IINTERFACE) {
			t.Error("ICustom should be subinterface of IINTERFACE")
		}
	})
}

// Task 7.73-7.74: Test interface with Parent, IsExternal, ExternalName
func TestInterfaceTypeWithInheritance(t *testing.T) {
	// Create base interface
	iBase := NewInterfaceType("IBase")
	iBase.Methods["BaseMethod"] = NewProcedureType([]Type{})

	// Create derived interface with parent
	iDerived := &InterfaceType{
		Name:    "IDerived",
		Parent:  iBase,
		Methods: make(map[string]*FunctionType),
	}
	iDerived.Methods["DerivedMethod"] = NewProcedureType([]Type{})

	t.Run("interface has parent", func(t *testing.T) {
		if iDerived.Parent == nil {
			t.Error("IDerived should have IBase as parent")
		}
		if iDerived.Parent.Name != "IBase" {
			t.Errorf("Expected parent name 'IBase', got '%s'", iDerived.Parent.Name)
		}
	})

	t.Run("base interface has no parent", func(t *testing.T) {
		if iBase.Parent != nil {
			t.Error("IBase should have no parent")
		}
	})
}

func TestInterfaceTypeExternal(t *testing.T) {
	// Create external interface (for FFI)
	iExternal := &InterfaceType{
		Name:         "IExternal",
		Methods:      make(map[string]*FunctionType),
		IsExternal:   true,
		ExternalName: "IDispatch",
	}

	t.Run("external interface flag", func(t *testing.T) {
		if !iExternal.IsExternal {
			t.Error("Interface should be marked as external")
		}
		if iExternal.ExternalName != "IDispatch" {
			t.Errorf("Expected external name 'IDispatch', got '%s'", iExternal.ExternalName)
		}
	})

	// Create regular (non-external) interface
	iRegular := NewInterfaceType("IRegular")

	t.Run("regular interface not external", func(t *testing.T) {
		if iRegular.IsExternal {
			t.Error("Regular interface should not be marked as external")
		}
		if iRegular.ExternalName != "" {
			t.Error("Regular interface should have empty external name")
		}
	})
}

// Task 7.77: Test interface inheritance checking
func TestIsSubinterfaceOf(t *testing.T) {
	// Create interface hierarchy: IBase -> IMiddle -> IDerived
	iBase := NewInterfaceType("IBase")
	iBase.Methods["BaseMethod"] = NewProcedureType([]Type{})

	iMiddle := &InterfaceType{
		Name:    "IMiddle",
		Parent:  iBase,
		Methods: make(map[string]*FunctionType),
	}
	iMiddle.Methods["MiddleMethod"] = NewProcedureType([]Type{})

	iDerived := &InterfaceType{
		Name:    "IDerived",
		Parent:  iMiddle,
		Methods: make(map[string]*FunctionType),
	}
	iDerived.Methods["DerivedMethod"] = NewProcedureType([]Type{})

	iUnrelated := NewInterfaceType("IUnrelated")

	tests := []struct {
		child    *InterfaceType
		parent   *InterfaceType
		name     string
		expected bool
	}{
		{
			name:     "direct parent",
			child:    iDerived,
			parent:   iMiddle,
			expected: true,
		},
		{
			name:     "grandparent",
			child:    iDerived,
			parent:   iBase,
			expected: true,
		},
		{
			name:     "immediate parent",
			child:    iMiddle,
			parent:   iBase,
			expected: true,
		},
		{
			name:     "same interface",
			child:    iMiddle,
			parent:   iMiddle,
			expected: true,
		},
		{
			name:     "unrelated interfaces",
			child:    iDerived,
			parent:   iUnrelated,
			expected: false,
		},
		{
			name:     "reverse hierarchy",
			child:    iBase,
			parent:   iMiddle,
			expected: false,
		},
		{
			name:     "nil child",
			child:    nil,
			parent:   iBase,
			expected: false,
		},
		{
			name:     "nil parent",
			child:    iDerived,
			parent:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSubinterfaceOf(tt.child, tt.parent)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Task 7.77: Test circular interface inheritance detection
func TestCircularInterfaceInheritance(t *testing.T) {
	// This test will verify that we detect circular inheritance
	// For now, we'll test the concept - actual implementation may vary
	iA := NewInterfaceType("IA")
	iB := &InterfaceType{
		Name:    "IB",
		Parent:  iA,
		Methods: make(map[string]*FunctionType),
	}
	// Attempting to make iA inherit from iB would create a cycle
	// This should be detected during semantic analysis

	t.Run("no circular inheritance", func(t *testing.T) {
		// Verify simple inheritance works
		if !IsSubinterfaceOf(iB, iA) {
			t.Error("IB should be subinterface of IA")
		}
	})
}

// Task 7.78: Test interface method inheritance
func TestInterfaceMethodInheritance(t *testing.T) {
	// Create base interface with methods
	iBase := NewInterfaceType("IBase")
	iBase.Methods["BaseMethod"] = NewProcedureType([]Type{})
	iBase.Methods["GetValue"] = NewFunctionType([]Type{}, INTEGER)

	// Create derived interface
	iDerived := &InterfaceType{
		Name:    "IDerived",
		Parent:  iBase,
		Methods: make(map[string]*FunctionType),
	}
	iDerived.Methods["DerivedMethod"] = NewProcedureType([]Type{})

	t.Run("interface has all inherited methods", func(t *testing.T) {
		// Interface should have access to parent methods
		allMethods := GetAllInterfaceMethods(iDerived)

		if len(allMethods) != 3 {
			t.Errorf("Expected 3 methods (1 own + 2 inherited), got %d", len(allMethods))
		}

		if _, ok := allMethods["BaseMethod"]; !ok {
			t.Error("Should have inherited BaseMethod from parent")
		}
		if _, ok := allMethods["GetValue"]; !ok {
			t.Error("Should have inherited GetValue from parent")
		}
		if _, ok := allMethods["DerivedMethod"]; !ok {
			t.Error("Should have own DerivedMethod")
		}
	})
}

// Task 7.79: Test interface-to-interface assignment
func TestInterfaceToInterfaceAssignment(t *testing.T) {
	// Create interface hierarchy
	iBase := NewInterfaceType("IBase")
	iBase.Methods["BaseMethod"] = NewProcedureType([]Type{})

	iDerived := &InterfaceType{
		Name:    "IDerived",
		Parent:  iBase,
		Methods: make(map[string]*FunctionType),
	}
	iDerived.Methods["DerivedMethod"] = NewProcedureType([]Type{})

	iUnrelated := NewInterfaceType("IUnrelated")
	iUnrelated.Methods["UnrelatedMethod"] = NewProcedureType([]Type{})

	tests := []struct {
		target   Type
		source   Type
		name     string
		expected bool
	}{
		{
			name:     "derived to base interface (covariant)",
			target:   iBase,
			source:   iDerived,
			expected: true,
		},
		{
			name:     "same interface",
			target:   iBase,
			source:   iBase,
			expected: true,
		},
		{
			name:     "base to derived (contravariant - not allowed)",
			target:   iDerived,
			source:   iBase,
			expected: false,
		},
		{
			name:     "unrelated interfaces",
			target:   iBase,
			source:   iUnrelated,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAssignableFrom(tt.target, tt.source)
			if result != tt.expected {
				t.Errorf("IsAssignableFrom(%v, %v) = %v, want %v",
					tt.target, tt.source, result, tt.expected)
			}
		})
	}
}

// Task 7.80: Test class with multiple interface implementation
func TestClassWithMultipleInterfaces(t *testing.T) {
	// Create multiple interfaces
	iReadable := NewInterfaceType("IReadable")
	iReadable.Methods["Read"] = NewFunctionType([]Type{}, STRING)

	iWritable := NewInterfaceType("IWritable")
	iWritable.Methods["Write"] = NewProcedureType([]Type{STRING})

	// Create class implementing both
	tFile := NewClassType("TFile", nil)
	tFile.Methods["Read"] = NewFunctionType([]Type{}, STRING)
	tFile.Methods["Write"] = NewProcedureType([]Type{STRING})
	tFile.Interfaces = []*InterfaceType{iReadable, iWritable}

	t.Run("class tracks implemented interfaces", func(t *testing.T) {
		if len(tFile.Interfaces) != 2 {
			t.Errorf("Expected 2 interfaces, got %d", len(tFile.Interfaces))
		}
	})

	t.Run("class implements all tracked interfaces", func(t *testing.T) {
		for _, iface := range tFile.Interfaces {
			if !ImplementsInterface(tFile, iface) {
				t.Errorf("Class should implement interface %s", iface.Name)
			}
		}
	})
}

// Task 7.80: Test that ClassType has Interfaces field
func TestClassTypeInterfacesField(t *testing.T) {
	iComparable := NewInterfaceType("IComparable")
	iComparable.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)

	tPerson := NewClassType("TPerson", nil)
	tPerson.Methods["CompareTo"] = NewFunctionType([]Type{INTEGER}, INTEGER)
	tPerson.Interfaces = []*InterfaceType{iComparable}

	t.Run("class has interfaces field", func(t *testing.T) {
		if tPerson.Interfaces == nil {
			t.Error("ClassType should have Interfaces field initialized")
		}
		if len(tPerson.Interfaces) != 1 {
			t.Errorf("Expected 1 interface, got %d", len(tPerson.Interfaces))
		}
		if tPerson.Interfaces[0].Name != "IComparable" {
			t.Errorf("Expected interface IComparable, got %s", tPerson.Interfaces[0].Name)
		}
	})
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
		iface    *InterfaceType
		name     string
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

// ============================================================================
// Task 7.137: External Class Tests
// ============================================================================

// TestExternalClassFields tests that ClassType has IsExternal and ExternalName fields
func TestExternalClassFields(t *testing.T) {
	t.Run("regular class is not external", func(t *testing.T) {
		tRegular := NewClassType("TRegular", nil)
		if tRegular.IsExternal {
			t.Error("Regular class should not be marked as external")
		}
		if tRegular.ExternalName != "" {
			t.Errorf("Regular class should have empty ExternalName, got %q", tRegular.ExternalName)
		}
	})

	t.Run("external class without external name", func(t *testing.T) {
		tExternal := NewClassType("TExternal", nil)
		tExternal.IsExternal = true
		if !tExternal.IsExternal {
			t.Error("External class should be marked as external")
		}
		if tExternal.ExternalName != "" {
			t.Errorf("External class without name should have empty ExternalName, got %q", tExternal.ExternalName)
		}
	})

	t.Run("external class with external name", func(t *testing.T) {
		tExternal := NewClassType("TExternal", nil)
		tExternal.IsExternal = true
		tExternal.ExternalName = "MyExternalClass"
		if !tExternal.IsExternal {
			t.Error("External class should be marked as external")
		}
		if tExternal.ExternalName != "MyExternalClass" {
			t.Errorf("Expected ExternalName 'MyExternalClass', got %q", tExternal.ExternalName)
		}
	})

	t.Run("external class with parent", func(t *testing.T) {
		tParent := NewClassType("TParentExternal", nil)
		tParent.IsExternal = true

		tChild := NewClassType("TChildExternal", tParent)
		tChild.IsExternal = true
		tChild.ExternalName = "ChildExternal"

		if !tChild.IsExternal {
			t.Error("Child external class should be marked as external")
		}
		if tChild.ExternalName != "ChildExternal" {
			t.Errorf("Expected ExternalName 'ChildExternal', got %q", tChild.ExternalName)
		}
	})
}

// ============================================================================
// EnumType Tests
// ============================================================================

func TestEnumType(t *testing.T) {
	t.Run("Create basic enum", func(t *testing.T) {
		// Create a simple color enum: type TColor = (Red, Green, Blue);
		colorEnum := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0, "Green": 1, "Blue": 2},
			OrderedNames: []string{"Red", "Green", "Blue"},
		}

		// Test String() method
		if colorEnum.String() != "TColor" {
			t.Errorf("String() = %v, want TColor", colorEnum.String())
		}

		// Test TypeKind() method
		if colorEnum.TypeKind() != "ENUM" {
			t.Errorf("TypeKind() = %v, want ENUM", colorEnum.TypeKind())
		}
	})

	t.Run("GetEnumValue lookup", func(t *testing.T) {
		colorEnum := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0, "Green": 1, "Blue": 2},
			OrderedNames: []string{"Red", "Green", "Blue"},
		}

		// Test forward lookup (name -> value)
		if val := colorEnum.GetEnumValue("Red"); val != 0 {
			t.Errorf("GetEnumValue('Red') = %v, want 0", val)
		}
		if val := colorEnum.GetEnumValue("Green"); val != 1 {
			t.Errorf("GetEnumValue('Green') = %v, want 1", val)
		}
		if val := colorEnum.GetEnumValue("Blue"); val != 2 {
			t.Errorf("GetEnumValue('Blue') = %v, want 2", val)
		}

		// Test invalid name (should return -1)
		if val := colorEnum.GetEnumValue("Yellow"); val != -1 {
			t.Errorf("GetEnumValue('Yellow') = %v, want -1", val)
		}
	})

	t.Run("GetEnumName reverse lookup", func(t *testing.T) {
		colorEnum := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0, "Green": 1, "Blue": 2},
			OrderedNames: []string{"Red", "Green", "Blue"},
		}

		// Test reverse lookup (value -> name)
		if name := colorEnum.GetEnumName(0); name != "Red" {
			t.Errorf("GetEnumName(0) = %v, want Red", name)
		}
		if name := colorEnum.GetEnumName(1); name != "Green" {
			t.Errorf("GetEnumName(1) = %v, want Green", name)
		}
		if name := colorEnum.GetEnumName(2); name != "Blue" {
			t.Errorf("GetEnumName(2) = %v, want Blue", name)
		}

		// Test invalid value
		if name := colorEnum.GetEnumName(99); name != "" {
			t.Errorf("GetEnumName(99) = %v, want empty string", name)
		}
	})

	t.Run("Enum with explicit values", func(t *testing.T) {
		// type TEnum = (One = 1, Two = 5, Three = 10);
		explicitEnum := &EnumType{
			Name:         "TEnum",
			Values:       map[string]int{"One": 1, "Two": 5, "Three": 10},
			OrderedNames: []string{"One", "Two", "Three"},
		}

		if val := explicitEnum.GetEnumValue("One"); val != 1 {
			t.Errorf("GetEnumValue('One') = %v, want 1", val)
		}
		if val := explicitEnum.GetEnumValue("Two"); val != 5 {
			t.Errorf("GetEnumValue('Two') = %v, want 5", val)
		}
		if val := explicitEnum.GetEnumValue("Three"); val != 10 {
			t.Errorf("GetEnumValue('Three') = %v, want 10", val)
		}

		if name := explicitEnum.GetEnumName(5); name != "Two" {
			t.Errorf("GetEnumName(5) = %v, want Two", name)
		}
	})

	t.Run("Enum equality", func(t *testing.T) {
		enum1 := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0, "Green": 1, "Blue": 2},
			OrderedNames: []string{"Red", "Green", "Blue"},
		}
		enum2 := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0, "Green": 1, "Blue": 2},
			OrderedNames: []string{"Red", "Green", "Blue"},
		}
		enum3 := &EnumType{
			Name:         "TSize",
			Values:       map[string]int{"Small": 0, "Medium": 1, "Large": 2},
			OrderedNames: []string{"Small", "Medium", "Large"},
		}

		// Same name enums should be equal (nominal typing)
		if !enum1.Equals(enum2) {
			t.Error("Enums with same name should be equal")
		}

		// Different name enums should not be equal
		if enum1.Equals(enum3) {
			t.Error("Enums with different names should not be equal")
		}

		// Enum should not equal other types
		if enum1.Equals(INTEGER) {
			t.Error("EnumType should not equal INTEGER")
		}
	})

	t.Run("IsOrdinalType with Enum", func(t *testing.T) {
		// Task 8.31: Test that enums are ordinal types
		colorEnum := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0, "Green": 1, "Blue": 2},
			OrderedNames: []string{"Red", "Green", "Blue"},
		}
		if !IsOrdinalType(colorEnum) {
			t.Error("EnumType should be ordinal")
		}
	})
}

// ============================================================================
// PropertyInfo Tests (Stage 8, Tasks 8.26-8.29)
// ============================================================================

func TestPropertyInfo(t *testing.T) {
	t.Run("create field-backed property", func(t *testing.T) {
		// Property: property Name: String read FName write FName;
		prop := &PropertyInfo{
			Name:      "Name",
			Type:      STRING,
			ReadKind:  PropAccessField,
			ReadSpec:  "FName",
			WriteKind: PropAccessField,
			WriteSpec: "FName",
			IsIndexed: false,
			IsDefault: false,
		}

		if prop.Name != "Name" {
			t.Errorf("Expected Name='Name', got '%s'", prop.Name)
		}
		if !prop.Type.Equals(STRING) {
			t.Errorf("Expected Type=STRING, got %v", prop.Type)
		}
		if prop.ReadKind != PropAccessField {
			t.Errorf("Expected ReadKind=PropAccessField, got %v", prop.ReadKind)
		}
		if prop.ReadSpec != "FName" {
			t.Errorf("Expected ReadSpec='FName', got '%s'", prop.ReadSpec)
		}
		if prop.WriteKind != PropAccessField {
			t.Errorf("Expected WriteKind=PropAccessField, got %v", prop.WriteKind)
		}
		if prop.WriteSpec != "FName" {
			t.Errorf("Expected WriteSpec='FName', got '%s'", prop.WriteSpec)
		}
		if prop.IsIndexed {
			t.Error("Expected IsIndexed=false")
		}
		if prop.IsDefault {
			t.Error("Expected IsDefault=false")
		}
	})

	t.Run("create method-backed property", func(t *testing.T) {
		// Property: property Count: Integer read GetCount write SetCount;
		prop := &PropertyInfo{
			Name:      "Count",
			Type:      INTEGER,
			ReadKind:  PropAccessMethod,
			ReadSpec:  "GetCount",
			WriteKind: PropAccessMethod,
			WriteSpec: "SetCount",
			IsIndexed: false,
			IsDefault: false,
		}

		if prop.ReadKind != PropAccessMethod {
			t.Errorf("Expected ReadKind=PropAccessMethod, got %v", prop.ReadKind)
		}
		if prop.WriteKind != PropAccessMethod {
			t.Errorf("Expected WriteKind=PropAccessMethod, got %v", prop.WriteKind)
		}
	})

	t.Run("create read-only property", func(t *testing.T) {
		// Property: property Size: Integer read FSize;
		prop := &PropertyInfo{
			Name:      "Size",
			Type:      INTEGER,
			ReadKind:  PropAccessField,
			ReadSpec:  "FSize",
			WriteKind: PropAccessNone,
			WriteSpec: "",
			IsIndexed: false,
			IsDefault: false,
		}

		if prop.WriteKind != PropAccessNone {
			t.Errorf("Expected WriteKind=PropAccessNone, got %v", prop.WriteKind)
		}
		if prop.WriteSpec != "" {
			t.Errorf("Expected empty WriteSpec, got '%s'", prop.WriteSpec)
		}
	})

	t.Run("create expression-backed property", func(t *testing.T) {
		// Property: property Double: Integer read (FValue * 2);
		prop := &PropertyInfo{
			Name:      "Double",
			Type:      INTEGER,
			ReadKind:  PropAccessExpression,
			ReadSpec:  "(FValue * 2)",
			WriteKind: PropAccessNone,
			WriteSpec: "",
			IsIndexed: false,
			IsDefault: false,
		}

		if prop.ReadKind != PropAccessExpression {
			t.Errorf("Expected ReadKind=PropAccessExpression, got %v", prop.ReadKind)
		}
	})

	t.Run("create indexed property", func(t *testing.T) {
		// Property: property Items[index: Integer]: String read GetItem write SetItem;
		prop := &PropertyInfo{
			Name:      "Items",
			Type:      STRING,
			ReadKind:  PropAccessMethod,
			ReadSpec:  "GetItem",
			WriteKind: PropAccessMethod,
			WriteSpec: "SetItem",
			IsIndexed: true,
			IsDefault: false,
		}

		if !prop.IsIndexed {
			t.Error("Expected IsIndexed=true")
		}
	})

	t.Run("create default indexed property", func(t *testing.T) {
		// Property: property Items[index: Integer]: String read GetItem write SetItem; default;
		prop := &PropertyInfo{
			Name:      "Items",
			Type:      STRING,
			ReadKind:  PropAccessMethod,
			ReadSpec:  "GetItem",
			WriteKind: PropAccessMethod,
			WriteSpec: "SetItem",
			IsIndexed: true,
			IsDefault: true,
		}

		if !prop.IsDefault {
			t.Error("Expected IsDefault=true")
		}
	})
}

func TestClassTypeProperties(t *testing.T) {
	t.Run("HasProperty and GetProperty", func(t *testing.T) {
		// Create a class with properties
		class := NewClassType("TPerson", nil)
		class.Fields["FName"] = STRING
		class.Fields["FAge"] = INTEGER

		// Add a field-backed property
		class.Properties["Name"] = &PropertyInfo{
			Name:      "Name",
			Type:      STRING,
			ReadKind:  PropAccessField,
			ReadSpec:  "FName",
			WriteKind: PropAccessField,
			WriteSpec: "FName",
		}

		// Add a method-backed property
		class.Properties["Age"] = &PropertyInfo{
			Name:      "Age",
			Type:      INTEGER,
			ReadKind:  PropAccessMethod,
			ReadSpec:  "GetAge",
			WriteKind: PropAccessMethod,
			WriteSpec: "SetAge",
		}

		// Test HasProperty
		if !class.HasProperty("Name") {
			t.Error("Should have property Name")
		}
		if !class.HasProperty("Age") {
			t.Error("Should have property Age")
		}
		if class.HasProperty("Email") {
			t.Error("Should not have property Email")
		}

		// Test GetProperty
		nameProp, found := class.GetProperty("Name")
		if !found {
			t.Error("GetProperty should find Name")
		}
		if nameProp.Name != "Name" {
			t.Errorf("Expected property Name, got %s", nameProp.Name)
		}
		if !nameProp.Type.Equals(STRING) {
			t.Errorf("Expected property type STRING, got %v", nameProp.Type)
		}

		ageProp, found := class.GetProperty("Age")
		if !found {
			t.Error("GetProperty should find Age")
		}
		if ageProp.ReadKind != PropAccessMethod {
			t.Errorf("Expected ReadKind=PropAccessMethod, got %v", ageProp.ReadKind)
		}

		_, found = class.GetProperty("Email")
		if found {
			t.Error("GetProperty should not find Email")
		}
	})

	t.Run("property inheritance", func(t *testing.T) {
		// Create parent class with a property
		parent := NewClassType("TBase", nil)
		parent.Properties["BaseProperty"] = &PropertyInfo{
			Name:      "BaseProperty",
			Type:      INTEGER,
			ReadKind:  PropAccessField,
			ReadSpec:  "FBase",
			WriteKind: PropAccessNone,
			WriteSpec: "",
		}

		// Create child class with its own property
		child := NewClassType("TDerived", parent)
		child.Properties["ChildProperty"] = &PropertyInfo{
			Name:      "ChildProperty",
			Type:      STRING,
			ReadKind:  PropAccessField,
			ReadSpec:  "FChild",
			WriteKind: PropAccessField,
			WriteSpec: "FChild",
		}

		// Test HasProperty with inheritance
		if !child.HasProperty("ChildProperty") {
			t.Error("Should have own property ChildProperty")
		}
		if !child.HasProperty("BaseProperty") {
			t.Error("Should have inherited property BaseProperty")
		}

		// Test GetProperty with inheritance
		baseProp, found := child.GetProperty("BaseProperty")
		if !found {
			t.Error("GetProperty should find inherited BaseProperty")
		}
		if baseProp.Name != "BaseProperty" {
			t.Errorf("Expected property BaseProperty, got %s", baseProp.Name)
		}

		childProp, found := child.GetProperty("ChildProperty")
		if !found {
			t.Error("GetProperty should find own ChildProperty")
		}
		if childProp.Name != "ChildProperty" {
			t.Errorf("Expected property ChildProperty, got %s", childProp.Name)
		}
	})
}

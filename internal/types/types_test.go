package types

import (
	"testing"
)

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
// Record Type Tests
// ============================================================================

func TestRecordType(t *testing.T) {
	t.Run("Named record", func(t *testing.T) {
		fields := map[string]Type{
			"X": INTEGER,
			"Y": INTEGER,
		}
		rt := NewRecordType("TPoint", fields)
		assertStringEquals(t, rt.String(), "TPoint", "String()")
		assertHasField(t, rt, "X", true)
		assertHasField(t, rt, "Y", true)
		assertHasField(t, rt, "Z", false)
		assertTypeEquals(t, rt.GetFieldType("X"), INTEGER, "GetFieldType(X)")
	})

	t.Run("Anonymous record", func(t *testing.T) {
		fields := map[string]Type{
			"Name": STRING,
			"Age":  INTEGER,
		}
		rt := NewRecordType("", fields)
		str := rt.String()
		// String should contain both fields (keys are normalized to lowercase)
		if !contains(str, "name") || !contains(str, "age") {
			t.Errorf("String() = %v, should contain name and age", str)
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

// Test RecordType with methods and properties
func TestRecordTypeWithMethods(t *testing.T) {
	t.Run("Record with methods", func(t *testing.T) {
		fields := map[string]Type{
			"X": INTEGER,
			"Y": INTEGER,
		}
		rt := NewRecordType("TPoint", fields)

		// Add methods to the record (use lowercase keys for case-insensitive lookup)
		rt.Methods = make(map[string]*FunctionType)
		rt.Methods["getdistance"] = NewFunctionType([]Type{}, FLOAT)
		rt.Methods["setposition"] = NewProcedureType([]Type{INTEGER, INTEGER})

		// Test HasMethod
		assertHasMethod(t, rt, "GetDistance", true)
		assertHasMethod(t, rt, "SetPosition", true)
		assertHasMethod(t, rt, "NonExistent", false)

		// Test GetMethod
		method := rt.GetMethod("GetDistance")
		assertMethodType(t, method, true, "GetDistance")

		method = rt.GetMethod("SetPosition")
		assertMethodType(t, method, false, "SetPosition")

		// Test GetMethod for non-existent method
		method = rt.GetMethod("NonExistent")
		assertBoolCondition(t, method == nil, "GetMethod should return nil for non-existent method")
	})

	t.Run("Record without methods", func(t *testing.T) {
		fields := map[string]Type{
			"X": INTEGER,
			"Y": INTEGER,
		}
		rt := NewRecordType("TPoint", fields)

		// Methods map should be initialized but empty
		assertBoolCondition(t, rt.Methods != nil, "Methods map should be initialized")
		assertHasMethod(t, rt, "AnyMethod", false)
	})
}

func TestRecordTypeWithProperties(t *testing.T) {
	t.Run("Record with properties", func(t *testing.T) {
		fields := map[string]Type{
			"FX": INTEGER,
			"FY": INTEGER,
		}
		rt := NewRecordType("TPoint", fields)

		// Add properties to the record (use lowercase keys for case-insensitive lookup)
		rt.Properties = make(map[string]*RecordPropertyInfo)
		rt.Properties["x"] = &RecordPropertyInfo{
			Name:       "X",
			Type:       INTEGER,
			ReadField:  "FX",
			WriteField: "FX",
		}
		rt.Properties["y"] = &RecordPropertyInfo{
			Name:       "Y",
			Type:       INTEGER,
			ReadField:  "FY",
			WriteField: "FY",
		}

		// Test HasProperty
		assertHasProperty(t, rt, "X", true)
		assertHasProperty(t, rt, "Y", true)
		assertHasProperty(t, rt, "Z", false)

		// Test GetProperty
		prop := rt.GetProperty("X")
		assertPropertyType(t, prop, INTEGER, "Property X")

		// Test GetProperty for non-existent property
		prop = rt.GetProperty("Z")
		assertBoolCondition(t, prop == nil, "GetProperty should return nil for non-existent property")
	})

	t.Run("Record without properties", func(t *testing.T) {
		fields := map[string]Type{
			"X": INTEGER,
			"Y": INTEGER,
		}
		rt := NewRecordType("TPoint", fields)

		// Properties map should be initialized but empty
		assertBoolCondition(t, rt.Properties != nil, "Properties map should be initialized")
		assertHasProperty(t, rt, "AnyProperty", false)
	})
}

// ============================================================================
// TypeAlias Tests
// ============================================================================

func TestTypeAlias(t *testing.T) {
	t.Run("Create type alias", func(t *testing.T) {
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
		// type MyInt = Integer;
		// MyInt should equal Integer
		myInt := &TypeAlias{
			Name:        "MyInt",
			AliasedType: INTEGER,
		}

		assertTypesEqual(t, myInt, INTEGER, true, "Type alias should equal its underlying type")
		assertTypesEqual(t, INTEGER, myInt, true, "Underlying type should equal type alias (symmetric)")
	})

	t.Run("Alias inequality with different types", func(t *testing.T) {
		// type MyInt = Integer;
		// MyInt should NOT equal String
		myInt := &TypeAlias{
			Name:        "MyInt",
			AliasedType: INTEGER,
		}

		assertTypesEqual(t, myInt, STRING, false, "Integer alias should not equal STRING")
		assertTypesEqual(t, myInt, FLOAT, false, "Integer alias should not equal FLOAT")
		assertTypesEqual(t, myInt, BOOLEAN, false, "Integer alias should not equal BOOLEAN")

		// type MyString = String;
		// MyString should NOT equal MyInt
		myString := &TypeAlias{
			Name:        "MyString",
			AliasedType: STRING,
		}

		assertTypesEqual(t, myInt, myString, false, "Integer alias should not equal String alias")
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

		assertTypesEqual(t, userID, itemID, true, "Two aliases to the same type should be equal")
	})

	t.Run("Nested aliases", func(t *testing.T) {
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
		assertStringEquals(t, aliasB.String(), "B", "String()")

		// B.TypeKind() should return "INTEGER" (resolved through chain)
		assertStringEquals(t, aliasB.TypeKind(), "INTEGER", "TypeKind()")

		// B should equal Integer
		assertTypesEqual(t, aliasB, INTEGER, true, "Nested alias should equal ultimate underlying type (Integer)")

		// B should equal A
		assertTypesEqual(t, aliasB, aliasA, true, "Nested alias should equal intermediate alias")

		// A should equal B (symmetric)
		assertTypesEqual(t, aliasA, aliasB, true, "Alias equality should be symmetric")
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
		assertStringEquals(t, aliasC.TypeKind(), "INTEGER", "TypeKind()")

		// C should equal Integer
		assertTypesEqual(t, aliasC, INTEGER, true, "Triple nested alias should equal ultimate underlying type")

		// C should equal A and B
		assertTypesEqual(t, aliasC, aliasA, true, "Triple nested alias should equal first alias in chain")
		assertTypesEqual(t, aliasC, aliasB, true, "Triple nested alias should equal second alias in chain")
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
		assertTypesEqual(t, myInt, funcType, false, "Integer alias should not equal function type")

		// Should not equal array type
		arrayType := NewDynamicArrayType(INTEGER)
		assertTypesEqual(t, myInt, arrayType, false, "Integer alias should not equal array type")

		// Should not equal class type
		classType := NewClassType("TMyClass", nil)
		assertTypesEqual(t, myInt, classType, false, "Integer alias should not equal class type")
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

	// Test Variant with type aliases
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

	// Test that enums are ordinal types
	t.Run("IsOrdinalType with Enum", func(t *testing.T) {
		colorEnum := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0, "Green": 1, "Blue": 2},
			OrderedNames: []string{"Red", "Green", "Blue"},
		}
		if !IsOrdinalType(colorEnum) {
			t.Error("EnumType should be ordinal")
		}
	})

	t.Run("OrdinalBounds", func(t *testing.T) {
		enum := &EnumType{
			Name:         "TFlag",
			Values:       map[string]int{"Off": -1, "On": 3},
			OrderedNames: []string{"Off", "On"},
		}

		if low, high, ok := OrdinalBounds(BOOLEAN); !ok || low != 0 || high != 1 {
			t.Errorf("OrdinalBounds(BOOLEAN) = (%d,%d,%v), want (0,1,true)", low, high, ok)
		}

		if low, high, ok := OrdinalBounds(enum); !ok || low != -1 || high != 3 {
			t.Errorf("OrdinalBounds(enum) = (%d,%d,%v), want (-1,3,true)", low, high, ok)
		}

		subrange := &SubrangeType{BaseType: INTEGER, LowBound: 5, HighBound: 7}
		if low, high, ok := OrdinalBounds(subrange); !ok || low != 5 || high != 7 {
			t.Errorf("OrdinalBounds(subrange) = (%d,%d,%v), want (5,7,true)", low, high, ok)
		}

		if _, _, ok := OrdinalBounds(STRING); ok {
			t.Error("OrdinalBounds(STRING) should return ok=false for unbounded ordinal type")
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

// assertStringEquals checks if a string matches expected value
func assertStringEquals(t *testing.T, got, want, context string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", context, got, want)
	}
}

// assertTypeEquals checks if two types are equal
func assertTypeEquals(t *testing.T, got, want Type, context string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", context, got, want)
	}
}

// assertHasField checks if a record has a field
func assertHasField(t *testing.T, rt *RecordType, fieldName string, shouldHave bool) {
	t.Helper()
	has := rt.HasField(fieldName)
	if has != shouldHave {
		if shouldHave {
			t.Errorf("Should have field %s", fieldName)
		} else {
			t.Errorf("Should not have field %s", fieldName)
		}
	}
}

// assertHasMethod checks if a record has a method
func assertHasMethod(t *testing.T, rt *RecordType, methodName string, shouldHave bool) {
	t.Helper()
	has := rt.HasMethod(methodName)
	if has != shouldHave {
		if shouldHave {
			t.Errorf("Should have method %s", methodName)
		} else {
			t.Errorf("Should not have method %s", methodName)
		}
	}
}

// assertHasProperty checks if a record has a property
func assertHasProperty(t *testing.T, rt *RecordType, propName string, shouldHave bool) {
	t.Helper()
	has := rt.HasProperty(propName)
	if has != shouldHave {
		if shouldHave {
			t.Errorf("Should have property %s", propName)
		} else {
			t.Errorf("Should not have property %s", propName)
		}
	}
}

// assertMethodType checks if a method has expected function/procedure type
func assertMethodType(t *testing.T, method *FunctionType, isFunction bool, context string) {
	t.Helper()
	if method == nil {
		t.Errorf("%s: method should not be nil", context)
		return
	}
	if isFunction && !method.IsFunction() {
		t.Errorf("%s should be a function", context)
	}
	if !isFunction && !method.IsProcedure() {
		t.Errorf("%s should be a procedure", context)
	}
}

// assertPropertyType checks if a property has expected type
func assertPropertyType(t *testing.T, prop *RecordPropertyInfo, expectedType Type, context string) {
	t.Helper()
	if prop == nil {
		t.Errorf("%s: property should not be nil", context)
		return
	}
	if prop.Type != expectedType {
		t.Errorf("%s should be %v type, got %v", context, expectedType, prop.Type)
	}
}

// assertTypesEqual checks if two types are equal using Equals method
func assertTypesEqual(t *testing.T, a, b Type, shouldBeEqual bool, context string) {
	t.Helper()
	result := a.Equals(b)
	if result != shouldBeEqual {
		if shouldBeEqual {
			t.Errorf("%s: types should be equal", context)
		} else {
			t.Errorf("%s: types should not be equal", context)
		}
	}
}

// assertBoolCondition checks a boolean condition with custom error message
func assertBoolCondition(t *testing.T, condition bool, errorMsg string) {
	t.Helper()
	if !condition {
		t.Error(errorMsg)
	}
}

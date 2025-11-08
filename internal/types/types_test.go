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

// Test RecordType with methods and properties
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

		if !myInt.Equals(INTEGER) {
			t.Error("Type alias should equal its underlying type")
		}

		if !INTEGER.Equals(myInt) {
			t.Error("Underlying type should equal type alias (symmetric)")
		}
	})

	t.Run("Alias inequality with different types", func(t *testing.T) {
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

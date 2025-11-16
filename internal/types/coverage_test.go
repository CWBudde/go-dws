package types

import (
	"testing"
)

// ============================================================================
// Compatibility Tests - IsIdentical
// ============================================================================

func TestIsIdentical(t *testing.T) {
	tests := []struct {
		a        Type
		b        Type
		name     string
		expected bool
	}{
		{
			name:     "Same basic types",
			a:        INTEGER,
			b:        INTEGER,
			expected: true,
		},
		{
			name:     "Different basic types",
			a:        INTEGER,
			b:        FLOAT,
			expected: false,
		},
		{
			name:     "Same function types",
			a:        NewFunctionType([]Type{INTEGER}, STRING),
			b:        NewFunctionType([]Type{INTEGER}, STRING),
			expected: true,
		},
		{
			name:     "Different function types",
			a:        NewFunctionType([]Type{INTEGER}, STRING),
			b:        NewFunctionType([]Type{FLOAT}, STRING),
			expected: false,
		},
		{
			name:     "Type alias equals underlying",
			a:        &TypeAlias{Name: "MyInt", AliasedType: INTEGER},
			b:        INTEGER,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIdentical(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("IsIdentical(%v, %v) = %v, want %v",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// NilType Tests
// ============================================================================

func TestNilType(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		if NIL.String() != "Nil" {
			t.Errorf("NIL.String() = %v, want Nil", NIL.String())
		}
	})

	t.Run("TypeKind", func(t *testing.T) {
		if NIL.TypeKind() != "NIL" {
			t.Errorf("NIL.TypeKind() = %v, want NIL", NIL.TypeKind())
		}
	})

	t.Run("Equals same type", func(t *testing.T) {
		nil1 := &NilType{}
		nil2 := &NilType{}
		if !nil1.Equals(nil2) {
			t.Error("NilType should equal NilType")
		}
	})

	t.Run("Not equals different type", func(t *testing.T) {
		if NIL.Equals(INTEGER) {
			t.Error("NilType should not equal INTEGER")
		}
	})
}

// ============================================================================
// DateTimeType Tests
// ============================================================================

func TestDateTimeType(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		if DATETIME.String() != "TDateTime" {
			t.Errorf("DATETIME.String() = %v, want TDateTime", DATETIME.String())
		}
	})

	t.Run("TypeKind", func(t *testing.T) {
		if DATETIME.TypeKind() != "DATETIME" {
			t.Errorf("DATETIME.TypeKind() = %v, want DATETIME", DATETIME.TypeKind())
		}
	})

	t.Run("Equals same type", func(t *testing.T) {
		dt1 := &DateTimeType{}
		dt2 := &DateTimeType{}
		if !dt1.Equals(dt2) {
			t.Error("DateTimeType should equal DateTimeType")
		}
	})

	t.Run("Not equals different type", func(t *testing.T) {
		if DATETIME.Equals(FLOAT) {
			t.Error("DateTimeType should not equal FLOAT")
		}
	})
}

// ============================================================================
// ConstType Tests (DEPRECATED)
// ============================================================================

func TestConstType(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		if CONST.String() != "Const" {
			t.Errorf("CONST.String() = %v, want Const", CONST.String())
		}
	})

	t.Run("TypeKind", func(t *testing.T) {
		if CONST.TypeKind() != "CONST" {
			t.Errorf("CONST.TypeKind() = %v, want CONST", CONST.TypeKind())
		}
	})

	t.Run("Equals same type", func(t *testing.T) {
		c1 := &ConstType{}
		c2 := &ConstType{}
		if !c1.Equals(c2) {
			t.Error("ConstType should equal ConstType")
		}
	})

	t.Run("Not equals different type", func(t *testing.T) {
		if CONST.Equals(VARIANT) {
			t.Error("ConstType should not equal VARIANT")
		}
	})
}

// ============================================================================
// ClassMember String() Test
// ============================================================================

func TestClassMemberString(t *testing.T) {
	t.Run("SetStorageBitmask", func(t *testing.T) {
		sk := SetStorageBitmask
		if sk.String() != "bitmask" {
			t.Errorf("SetStorageBitmask.String() = %v, want bitmask", sk.String())
		}
	})

	t.Run("SetStorageMap", func(t *testing.T) {
		sk := SetStorageMap
		if sk.String() != "map" {
			t.Errorf("SetStorageMap.String() = %v, want map", sk.String())
		}
	})

	t.Run("Unknown storage kind", func(t *testing.T) {
		sk := SetStorageKind(999)
		if sk.String() != "unknown" {
			t.Errorf("Unknown SetStorageKind.String() = %v, want unknown", sk.String())
		}
	})
}

// ============================================================================
// NewEnumType Tests
// ============================================================================

func TestNewEnumType(t *testing.T) {
	t.Run("Create enum with constructor", func(t *testing.T) {
		values := map[string]int{
			"Red":   0,
			"Green": 1,
			"Blue":  2,
		}
		orderedNames := []string{"Red", "Green", "Blue"}

		enumType := NewEnumType("TColor", values, orderedNames)

		if enumType.Name != "TColor" {
			t.Errorf("Name = %v, want TColor", enumType.Name)
		}
		if len(enumType.Values) != 3 {
			t.Errorf("Values count = %v, want 3", len(enumType.Values))
		}
		if len(enumType.OrderedNames) != 3 {
			t.Errorf("OrderedNames count = %v, want 3", len(enumType.OrderedNames))
		}
		if enumType.GetEnumValue("Red") != 0 {
			t.Errorf("Red value = %v, want 0", enumType.GetEnumValue("Red"))
		}
	})

	t.Run("Create enum with explicit values", func(t *testing.T) {
		values := map[string]int{
			"One":  1,
			"Five": 5,
			"Ten":  10,
		}
		orderedNames := []string{"One", "Five", "Ten"}

		enumType := NewEnumType("TCustom", values, orderedNames)

		if enumType.GetEnumValue("Five") != 5 {
			t.Errorf("Five value = %v, want 5", enumType.GetEnumValue("Five"))
		}
		if enumType.GetEnumName(10) != "Ten" {
			t.Errorf("Value 10 name = %v, want Ten", enumType.GetEnumName(10))
		}
	})
}

// ============================================================================
// HelperType Tests
// ============================================================================

func TestHelperType(t *testing.T) {
	t.Run("String helper representation", func(t *testing.T) {
		helper := &HelperType{
			Name:           "TStringHelper",
			TargetType:     STRING,
			IsRecordHelper: false,
			Methods:        make(map[string]*FunctionType),
			Properties:     make(map[string]*PropertyInfo),
			ClassVars:      make(map[string]Type),
			ClassConsts:    make(map[string]interface{}),
			BuiltinMethods: make(map[string]string),
		}

		if helper.String() != "helper for String" {
			t.Errorf("String() = %v, want 'helper for String'", helper.String())
		}
	})

	t.Run("Record helper representation", func(t *testing.T) {
		pointType := NewRecordType("TPoint", map[string]Type{"X": INTEGER, "Y": INTEGER})
		helper := &HelperType{
			Name:           "TPointHelper",
			TargetType:     pointType,
			IsRecordHelper: true,
			Methods:        make(map[string]*FunctionType),
			Properties:     make(map[string]*PropertyInfo),
			ClassVars:      make(map[string]Type),
			ClassConsts:    make(map[string]interface{}),
			BuiltinMethods: make(map[string]string),
		}

		if helper.String() != "record helper for TPoint" {
			t.Errorf("String() = %v, want 'record helper for TPoint'", helper.String())
		}
	})

	t.Run("TypeKind", func(t *testing.T) {
		helper := &HelperType{
			Name:           "THelper",
			TargetType:     INTEGER,
			Methods:        make(map[string]*FunctionType),
			Properties:     make(map[string]*PropertyInfo),
			ClassVars:      make(map[string]Type),
			ClassConsts:    make(map[string]interface{}),
			BuiltinMethods: make(map[string]string),
		}

		if helper.TypeKind() != "HELPER" {
			t.Errorf("TypeKind() = %v, want HELPER", helper.TypeKind())
		}
	})

	t.Run("Equals same helper", func(t *testing.T) {
		helper1 := &HelperType{
			Name:           "THelper",
			TargetType:     STRING,
			Methods:        make(map[string]*FunctionType),
			Properties:     make(map[string]*PropertyInfo),
			ClassVars:      make(map[string]Type),
			ClassConsts:    make(map[string]interface{}),
			BuiltinMethods: make(map[string]string),
		}
		helper2 := &HelperType{
			Name:           "THelper",
			TargetType:     STRING,
			Methods:        make(map[string]*FunctionType),
			Properties:     make(map[string]*PropertyInfo),
			ClassVars:      make(map[string]Type),
			ClassConsts:    make(map[string]interface{}),
			BuiltinMethods: make(map[string]string),
		}

		if !helper1.Equals(helper2) {
			t.Error("Same helpers should be equal")
		}
	})

	t.Run("Not equals different name", func(t *testing.T) {
		helper1 := &HelperType{
			Name:           "THelper1",
			TargetType:     STRING,
			Methods:        make(map[string]*FunctionType),
			Properties:     make(map[string]*PropertyInfo),
			ClassVars:      make(map[string]Type),
			ClassConsts:    make(map[string]interface{}),
			BuiltinMethods: make(map[string]string),
		}
		helper2 := &HelperType{
			Name:           "THelper2",
			TargetType:     STRING,
			Methods:        make(map[string]*FunctionType),
			Properties:     make(map[string]*PropertyInfo),
			ClassVars:      make(map[string]Type),
			ClassConsts:    make(map[string]interface{}),
			BuiltinMethods: make(map[string]string),
		}

		if helper1.Equals(helper2) {
			t.Error("Helpers with different names should not be equal")
		}
	})

	t.Run("Not equals different target type", func(t *testing.T) {
		helper1 := &HelperType{
			Name:           "THelper",
			TargetType:     STRING,
			Methods:        make(map[string]*FunctionType),
			Properties:     make(map[string]*PropertyInfo),
			ClassVars:      make(map[string]Type),
			ClassConsts:    make(map[string]interface{}),
			BuiltinMethods: make(map[string]string),
		}
		helper2 := &HelperType{
			Name:           "THelper",
			TargetType:     INTEGER,
			Methods:        make(map[string]*FunctionType),
			Properties:     make(map[string]*PropertyInfo),
			ClassVars:      make(map[string]Type),
			ClassConsts:    make(map[string]interface{}),
			BuiltinMethods: make(map[string]string),
		}

		if helper1.Equals(helper2) {
			t.Error("Helpers with different target types should not be equal")
		}
	})

	t.Run("Not equals non-helper type", func(t *testing.T) {
		helper := &HelperType{
			Name:           "THelper",
			TargetType:     STRING,
			Methods:        make(map[string]*FunctionType),
			Properties:     make(map[string]*PropertyInfo),
			ClassVars:      make(map[string]Type),
			ClassConsts:    make(map[string]interface{}),
			BuiltinMethods: make(map[string]string),
		}

		if helper.Equals(STRING) {
			t.Error("HelperType should not equal STRING")
		}
	})
}

// ============================================================================
// ClassOfType Tests
// ============================================================================

func TestClassOfType(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		animalClass := NewClassType("TAnimal", nil)
		classOfAnimal := NewClassOfType(animalClass)

		if classOfAnimal.String() != "class of TAnimal" {
			t.Errorf("String() = %v, want 'class of TAnimal'", classOfAnimal.String())
		}
	})

	t.Run("String with nil class", func(t *testing.T) {
		classOf := &ClassOfType{ClassType: nil}

		if classOf.String() != "class of <unknown>" {
			t.Errorf("String() = %v, want 'class of <unknown>'", classOf.String())
		}
	})

	t.Run("TypeKind", func(t *testing.T) {
		animalClass := NewClassType("TAnimal", nil)
		classOfAnimal := NewClassOfType(animalClass)

		if classOfAnimal.TypeKind() != "CLASSOF" {
			t.Errorf("TypeKind() = %v, want CLASSOF", classOfAnimal.TypeKind())
		}
	})

	t.Run("Equals same class", func(t *testing.T) {
		animalClass := NewClassType("TAnimal", nil)
		classOf1 := NewClassOfType(animalClass)
		classOf2 := NewClassOfType(animalClass)

		if !classOf1.Equals(classOf2) {
			t.Error("Same ClassOfType should be equal")
		}
	})

	t.Run("Not equals different class", func(t *testing.T) {
		animalClass := NewClassType("TAnimal", nil)
		dogClass := NewClassType("TDog", nil)
		classOf1 := NewClassOfType(animalClass)
		classOf2 := NewClassOfType(dogClass)

		if classOf1.Equals(classOf2) {
			t.Error("Different ClassOfType should not be equal")
		}
	})

	t.Run("Not equals non-ClassOfType", func(t *testing.T) {
		animalClass := NewClassType("TAnimal", nil)
		classOfAnimal := NewClassOfType(animalClass)

		if classOfAnimal.Equals(animalClass) {
			t.Error("ClassOfType should not equal ClassType")
		}
	})

	t.Run("Equals with nil classes", func(t *testing.T) {
		classOf1 := &ClassOfType{ClassType: nil}
		classOf2 := &ClassOfType{ClassType: nil}

		if !classOf1.Equals(classOf2) {
			t.Error("ClassOfType with nil should equal ClassOfType with nil")
		}
	})

	t.Run("IsAssignableFrom exact type", func(t *testing.T) {
		animalClass := NewClassType("TAnimal", nil)
		classOfAnimal := NewClassOfType(animalClass)

		if !classOfAnimal.IsAssignableFrom(animalClass) {
			t.Error("ClassOf should accept exact class type")
		}
	})

	t.Run("IsAssignableFrom derived class", func(t *testing.T) {
		animalClass := NewClassType("TAnimal", nil)
		dogClass := NewClassType("TDog", animalClass)
		classOfAnimal := NewClassOfType(animalClass)

		if !classOfAnimal.IsAssignableFrom(dogClass) {
			t.Error("ClassOf should accept derived class")
		}
	})

	t.Run("IsAssignableFrom deeply derived class", func(t *testing.T) {
		animalClass := NewClassType("TAnimal", nil)
		dogClass := NewClassType("TDog", animalClass)
		puppyClass := NewClassType("TPuppy", dogClass)
		classOfAnimal := NewClassOfType(animalClass)

		if !classOfAnimal.IsAssignableFrom(puppyClass) {
			t.Error("ClassOf should accept deeply derived class")
		}
	})

	t.Run("IsAssignableFrom unrelated class", func(t *testing.T) {
		animalClass := NewClassType("TAnimal", nil)
		carClass := NewClassType("TCar", nil)
		classOfAnimal := NewClassOfType(animalClass)

		if classOfAnimal.IsAssignableFrom(carClass) {
			t.Error("ClassOf should not accept unrelated class")
		}
	})

	t.Run("IsAssignableFrom with nil", func(t *testing.T) {
		animalClass := NewClassType("TAnimal", nil)
		classOfAnimal := NewClassOfType(animalClass)

		if classOfAnimal.IsAssignableFrom(nil) {
			t.Error("ClassOf should not accept nil")
		}
	})
}

// ============================================================================
// InterfaceType InheritsFrom Tests
// ============================================================================

func TestInterfaceTypeInheritsFrom(t *testing.T) {
	t.Run("Direct parent", func(t *testing.T) {
		parent := NewInterfaceType("IBase")
		child := NewInterfaceType("IChild")
		child.Parent = parent

		if !child.InheritsFrom(parent) {
			t.Error("Interface should inherit from direct parent")
		}
	})

	t.Run("Grandparent", func(t *testing.T) {
		grandparent := NewInterfaceType("IGrandparent")
		parent := NewInterfaceType("IParent")
		parent.Parent = grandparent
		child := NewInterfaceType("IChild")
		child.Parent = parent

		if !child.InheritsFrom(grandparent) {
			t.Error("Interface should inherit from grandparent")
		}
	})

	t.Run("No inheritance", func(t *testing.T) {
		interface1 := NewInterfaceType("IOne")
		interface2 := NewInterfaceType("ITwo")

		if interface1.InheritsFrom(interface2) {
			t.Error("Unrelated interfaces should not inherit")
		}
	})

	t.Run("Nil interface", func(t *testing.T) {
		iface := NewInterfaceType("ITest")

		if iface.InheritsFrom(nil) {
			t.Error("Should return false for nil parent")
		}
	})

	t.Run("Nil receiver", func(t *testing.T) {
		parent := NewInterfaceType("IParent")
		var child *InterfaceType = nil

		if child.InheritsFrom(parent) {
			t.Error("Nil interface should not inherit")
		}
	})
}

// ============================================================================
// OperatorRegistry Tests
// ============================================================================

func TestOperatorRegistry(t *testing.T) {
	t.Run("Register and lookup operator", func(t *testing.T) {
		registry := NewOperatorRegistry()
		sig := &OperatorSignature{
			Operator:     "+",
			OperandTypes: []Type{INTEGER, INTEGER},
			ResultType:   INTEGER,
			Binding:      "builtin_add_int",
		}

		err := registry.Register(sig)
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}

		found, ok := registry.Lookup("+", []Type{INTEGER, INTEGER})
		if !ok {
			t.Error("Should find registered operator")
		}
		if found.ResultType != INTEGER {
			t.Errorf("Result type = %v, want INTEGER", found.ResultType)
		}
	})

	t.Run("Register duplicate operator", func(t *testing.T) {
		registry := NewOperatorRegistry()
		sig := &OperatorSignature{
			Operator:     "+",
			OperandTypes: []Type{INTEGER, INTEGER},
			ResultType:   INTEGER,
		}

		err := registry.Register(sig)
		if err != nil {
			t.Fatalf("First register failed: %v", err)
		}

		err = registry.Register(sig)
		if err != ErrOperatorDuplicate {
			t.Errorf("Expected ErrOperatorDuplicate, got %v", err)
		}
	})

	t.Run("Register nil signature", func(t *testing.T) {
		registry := NewOperatorRegistry()
		err := registry.Register(nil)
		if err == nil {
			t.Error("Should error on nil signature")
		}
	})

	t.Run("Lookup non-existent operator", func(t *testing.T) {
		registry := NewOperatorRegistry()
		_, ok := registry.Lookup("+", []Type{INTEGER, INTEGER})
		if ok {
			t.Error("Should not find non-existent operator")
		}
	})

	t.Run("Register multiple overloads", func(t *testing.T) {
		registry := NewOperatorRegistry()

		intSig := &OperatorSignature{
			Operator:     "+",
			OperandTypes: []Type{INTEGER, INTEGER},
			ResultType:   INTEGER,
		}
		floatSig := &OperatorSignature{
			Operator:     "+",
			OperandTypes: []Type{FLOAT, FLOAT},
			ResultType:   FLOAT,
		}

		if err := registry.Register(intSig); err != nil {
			t.Fatalf("Register int failed: %v", err)
		}
		if err := registry.Register(floatSig); err != nil {
			t.Fatalf("Register float failed: %v", err)
		}

		intFound, ok := registry.Lookup("+", []Type{INTEGER, INTEGER})
		if !ok || intFound.ResultType != INTEGER {
			t.Error("Should find integer overload")
		}

		floatFound, ok := registry.Lookup("+", []Type{FLOAT, FLOAT})
		if !ok || floatFound.ResultType != FLOAT {
			t.Error("Should find float overload")
		}
	})
}

// ============================================================================
// ConversionRegistry Tests
// ============================================================================

func TestConversionRegistry(t *testing.T) {
	t.Run("Register and find implicit conversion", func(t *testing.T) {
		registry := NewConversionRegistry()
		sig := &ConversionSignature{
			From:    INTEGER,
			To:      FLOAT,
			Kind:    ConversionImplicit,
			Binding: "int_to_float",
		}

		err := registry.Register(sig)
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}

		found, ok := registry.FindImplicit(INTEGER, FLOAT)
		if !ok {
			t.Error("Should find implicit conversion")
		}
		if found.Binding != "int_to_float" {
			t.Errorf("Binding = %v, want int_to_float", found.Binding)
		}
	})

	t.Run("Register and find explicit conversion", func(t *testing.T) {
		registry := NewConversionRegistry()
		sig := &ConversionSignature{
			From:    FLOAT,
			To:      INTEGER,
			Kind:    ConversionExplicit,
			Binding: "float_to_int",
		}

		err := registry.Register(sig)
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}

		found, ok := registry.FindExplicit(FLOAT, INTEGER)
		if !ok {
			t.Error("Should find explicit conversion")
		}
		if found.Binding != "float_to_int" {
			t.Errorf("Binding = %v, want float_to_int", found.Binding)
		}
	})

	t.Run("Register duplicate implicit conversion", func(t *testing.T) {
		registry := NewConversionRegistry()
		sig := &ConversionSignature{
			From: INTEGER,
			To:   FLOAT,
			Kind: ConversionImplicit,
		}

		err := registry.Register(sig)
		if err != nil {
			t.Fatalf("First register failed: %v", err)
		}

		err = registry.Register(sig)
		if err != ErrConversionDuplicate {
			t.Errorf("Expected ErrConversionDuplicate, got %v", err)
		}
	})

	t.Run("Register duplicate explicit conversion", func(t *testing.T) {
		registry := NewConversionRegistry()
		sig := &ConversionSignature{
			From: FLOAT,
			To:   INTEGER,
			Kind: ConversionExplicit,
		}

		err := registry.Register(sig)
		if err != nil {
			t.Fatalf("First register failed: %v", err)
		}

		err = registry.Register(sig)
		if err != ErrConversionDuplicate {
			t.Errorf("Expected ErrConversionDuplicate, got %v", err)
		}
	})

	t.Run("Register nil signature", func(t *testing.T) {
		registry := NewConversionRegistry()
		err := registry.Register(nil)
		if err == nil {
			t.Error("Should error on nil signature")
		}
	})

	t.Run("Find on nil registry", func(t *testing.T) {
		var registry *ConversionRegistry = nil
		_, ok := registry.FindImplicit(INTEGER, FLOAT)
		if ok {
			t.Error("Should return false for nil registry")
		}

		_, ok = registry.FindExplicit(FLOAT, INTEGER)
		if ok {
			t.Error("Should return false for nil registry")
		}
	})

	t.Run("Find non-existent conversion", func(t *testing.T) {
		registry := NewConversionRegistry()
		_, ok := registry.FindImplicit(STRING, INTEGER)
		if ok {
			t.Error("Should not find non-existent conversion")
		}
	})
}

// ============================================================================
// FunctionTypeWithMetadata Tests
// ============================================================================

func TestNewFunctionTypeWithMetadata(t *testing.T) {
	t.Run("Create function with metadata", func(t *testing.T) {
		params := []Type{INTEGER, INTEGER}
		names := []string{"x", "y"}
		defaults := []interface{}{nil, nil}
		lazy := []bool{false, false}
		varParams := []bool{false, false}
		constParams := []bool{false, false}

		ft := NewFunctionTypeWithMetadata(
			params,
			names,
			defaults,
			lazy,
			varParams,
			constParams,
			INTEGER,
		)

		if len(ft.Parameters) != 2 {
			t.Errorf("Parameters count = %v, want 2", len(ft.Parameters))
		}
		if len(ft.ParamNames) != 2 {
			t.Errorf("ParamNames count = %v, want 2", len(ft.ParamNames))
		}
		if ft.ParamNames[0] != "x" {
			t.Errorf("First param name = %v, want x", ft.ParamNames[0])
		}
		if ft.DefaultValues[0] != nil {
			t.Error("Default value should be nil")
		}
	})

	t.Run("Create function with default values", func(t *testing.T) {
		params := []Type{INTEGER, STRING}
		names := []string{"count", "message"}
		defaults := []interface{}{10, "hello"}
		lazy := []bool{false, false}
		varParams := []bool{false, false}
		constParams := []bool{false, true}

		ft := NewFunctionTypeWithMetadata(
			params,
			names,
			defaults,
			lazy,
			varParams,
			constParams,
			STRING,
		)

		if ft.DefaultValues[0] != 10 {
			t.Errorf("Default value = %v, want 10", ft.DefaultValues[0])
		}
		if ft.DefaultValues[1] != "hello" {
			t.Errorf("Default value = %v, want hello", ft.DefaultValues[1])
		}
		if !ft.ConstParams[1] {
			t.Error("Second param should be const")
		}
	})
}

// ============================================================================
// ClassType Advanced Method Tests
// ============================================================================

func TestClassTypeMethodOverloads(t *testing.T) {
	t.Run("Add and get method overloads", func(t *testing.T) {
		classType := NewClassType("TMyClass", nil)

		method1 := &MethodInfo{
			Signature: NewFunctionType([]Type{INTEGER}, STRING),
		}
		method2 := &MethodInfo{
			Signature: NewFunctionType([]Type{STRING}, STRING),
		}

		classType.AddMethodOverload("Process", method1)
		classType.AddMethodOverload("Process", method2)

		overloads := classType.GetMethodOverloads("Process")
		if len(overloads) != 2 {
			t.Errorf("Overloads count = %v, want 2", len(overloads))
		}
	})

	t.Run("GetMethodOverloads for non-existent method", func(t *testing.T) {
		classType := NewClassType("TMyClass", nil)
		overloads := classType.GetMethodOverloads("NonExistent")
		if overloads != nil {
			t.Errorf("Expected nil for non-existent method, got %v", overloads)
		}
	})
}

func TestClassTypeConstructorOverloads(t *testing.T) {
	t.Run("Add and get constructor overloads", func(t *testing.T) {
		classType := NewClassType("TMyClass", nil)

		ctor1 := &MethodInfo{
			Signature: NewProcedureType([]Type{}),
		}
		ctor2 := &MethodInfo{
			Signature: NewProcedureType([]Type{STRING}),
		}

		classType.AddConstructorOverload("Create", ctor1)
		classType.AddConstructorOverload("Create", ctor2)

		overloads := classType.GetConstructorOverloads("Create")
		if len(overloads) != 2 {
			t.Errorf("Constructor overloads count = %v, want 2", len(overloads))
		}
	})

	t.Run("Constructor name case-insensitive", func(t *testing.T) {
		classType := NewClassType("TMyClass", nil)

		ctor := &MethodInfo{
			Signature: NewProcedureType([]Type{}),
		}

		classType.AddConstructorOverload("CREATE", ctor)

		overloads := classType.GetConstructorOverloads("create")
		if len(overloads) != 1 {
			t.Errorf("Should find constructor case-insensitively, got %v", len(overloads))
		}
	})
}

func TestClassTypeOperators(t *testing.T) {
	t.Run("Register and lookup operator", func(t *testing.T) {
		classType := NewClassType("TMyClass", nil)

		sig := &OperatorSignature{
			Operator:     "+",
			OperandTypes: []Type{classType, classType},
			ResultType:   classType,
		}

		err := classType.RegisterOperator(sig)
		if err != nil {
			t.Fatalf("RegisterOperator failed: %v", err)
		}

		found, ok := classType.LookupOperator("+", []Type{classType, classType})
		if !ok {
			t.Error("Should find registered operator")
		}
		if found.ResultType != classType {
			t.Error("Result type mismatch")
		}
	})

	t.Run("Lookup operator in parent", func(t *testing.T) {
		parentClass := NewClassType("TParent", nil)
		childClass := NewClassType("TChild", parentClass)

		sig := &OperatorSignature{
			Operator:     "+",
			OperandTypes: []Type{parentClass, parentClass},
			ResultType:   parentClass,
		}

		err := parentClass.RegisterOperator(sig)
		if err != nil {
			t.Fatalf("RegisterOperator failed: %v", err)
		}

		found, ok := childClass.LookupOperator("+", []Type{parentClass, parentClass})
		if !ok {
			t.Error("Should find operator in parent")
		}
		if found == nil {
			t.Error("Found operator should not be nil")
		}
	})

	t.Run("Lookup on nil class", func(t *testing.T) {
		var classType *ClassType = nil
		_, ok := classType.LookupOperator("+", []Type{INTEGER, INTEGER})
		if ok {
			t.Error("Should return false for nil class")
		}
	})
}

func TestClassTypeHasConstructor(t *testing.T) {
	t.Run("Has direct constructor", func(t *testing.T) {
		classType := NewClassType("TMyClass", nil)
		classType.Constructors["create"] = NewProcedureType([]Type{})

		if !classType.HasConstructor("Create") {
			t.Error("Should find constructor case-insensitively")
		}
	})

	t.Run("Has constructor in parent", func(t *testing.T) {
		parentClass := NewClassType("TParent", nil)
		parentClass.Constructors["create"] = NewProcedureType([]Type{})

		childClass := NewClassType("TChild", parentClass)

		if !childClass.HasConstructor("Create") {
			t.Error("Should find constructor in parent")
		}
	})

	t.Run("No constructor", func(t *testing.T) {
		classType := NewClassType("TMyClass", nil)

		if classType.HasConstructor("Create") {
			t.Error("Should not find non-existent constructor")
		}
	})

	t.Run("Nil class", func(t *testing.T) {
		var classType *ClassType = nil
		if classType.HasConstructor("Create") {
			t.Error("Nil class should not have constructor")
		}
	})
}

func TestClassTypeConstants(t *testing.T) {
	t.Run("Has direct constant", func(t *testing.T) {
		classType := NewClassType("TMyClass", nil)
		classType.Constants["MaxSize"] = 100

		if !classType.HasConstant("MaxSize") {
			t.Error("Should find constant")
		}

		value, ok := classType.GetConstant("MaxSize")
		if !ok {
			t.Error("Should get constant")
		}
		if value != 100 {
			t.Errorf("Constant value = %v, want 100", value)
		}
	})

	t.Run("Has constant in parent", func(t *testing.T) {
		parentClass := NewClassType("TParent", nil)
		parentClass.Constants["MaxSize"] = 100

		childClass := NewClassType("TChild", parentClass)

		if !childClass.HasConstant("MaxSize") {
			t.Error("Should find constant in parent")
		}

		value, ok := childClass.GetConstant("MaxSize")
		if !ok {
			t.Error("Should get constant from parent")
		}
		if value != 100 {
			t.Errorf("Constant value = %v, want 100", value)
		}
	})

	t.Run("No constant", func(t *testing.T) {
		classType := NewClassType("TMyClass", nil)

		if classType.HasConstant("MaxSize") {
			t.Error("Should not find non-existent constant")
		}

		_, ok := classType.GetConstant("MaxSize")
		if ok {
			t.Error("Should not get non-existent constant")
		}
	})

	t.Run("Nil class", func(t *testing.T) {
		var classType *ClassType = nil

		if classType.HasConstant("MaxSize") {
			t.Error("Nil class should not have constant")
		}

		_, ok := classType.GetConstant("MaxSize")
		if ok {
			t.Error("Nil class should not get constant")
		}
	})
}

func TestClassTypeImplementsInterface(t *testing.T) {
	t.Run("Direct implementation", func(t *testing.T) {
		iface := NewInterfaceType("IMyInterface")
		classType := NewClassType("TMyClass", nil)
		classType.Interfaces = []*InterfaceType{iface}

		if !classType.ImplementsInterface(iface) {
			t.Error("Should implement interface directly")
		}
	})

	t.Run("Parent implements interface", func(t *testing.T) {
		iface := NewInterfaceType("IMyInterface")
		parentClass := NewClassType("TParent", nil)
		parentClass.Interfaces = []*InterfaceType{iface}

		childClass := NewClassType("TChild", parentClass)

		if !childClass.ImplementsInterface(iface) {
			t.Error("Should inherit interface from parent")
		}
	})

	t.Run("Implements parent interface", func(t *testing.T) {
		parentIface := NewInterfaceType("IParent")
		childIface := NewInterfaceType("IChild")
		childIface.Parent = parentIface

		classType := NewClassType("TMyClass", nil)
		classType.Interfaces = []*InterfaceType{childIface}

		if !classType.ImplementsInterface(parentIface) {
			t.Error("Should implement parent interface")
		}
	})

	t.Run("Does not implement interface", func(t *testing.T) {
		iface := NewInterfaceType("IMyInterface")
		classType := NewClassType("TMyClass", nil)

		if classType.ImplementsInterface(iface) {
			t.Error("Should not implement unrelated interface")
		}
	})

	t.Run("Nil class or interface", func(t *testing.T) {
		iface := NewInterfaceType("IMyInterface")
		var classType *ClassType = nil

		if classType.ImplementsInterface(iface) {
			t.Error("Nil class should not implement interface")
		}

		classType = NewClassType("TMyClass", nil)
		if classType.ImplementsInterface(nil) {
			t.Error("Should not implement nil interface")
		}
	})
}

package types

import (
	"testing"
)

func TestFunctionTypeString(t *testing.T) {
	tests := []struct {
		name     string
		funcType *FunctionType
		expected string
	}{
		{
			name:     "simple function",
			funcType: NewFunctionType([]Type{INTEGER, STRING}, BOOLEAN),
			expected: "(Integer, String) -> Boolean",
		},
		{
			name:     "no parameters",
			funcType: NewFunctionType([]Type{}, INTEGER),
			expected: "() -> Integer",
		},
		{
			name:     "procedure (void return)",
			funcType: NewProcedureType([]Type{INTEGER}),
			expected: "(Integer) -> Void",
		},
		{
			name: "variadic function",
			funcType: NewVariadicFunctionType(
				[]Type{STRING, NewDynamicArrayType(INTEGER)},
				INTEGER,
				BOOLEAN,
			),
			expected: "(String, ...array of Integer) -> Boolean",
		},
		{
			name: "variadic with no fixed params",
			funcType: NewVariadicFunctionType(
				[]Type{NewDynamicArrayType(STRING)},
				STRING,
				VOID,
			),
			expected: "(...array of String) -> Void",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.funcType.String()
			if result != tt.expected {
				t.Errorf("String() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestFunctionTypeEquals(t *testing.T) {
	tests := []struct {
		ft2      Type
		ft1      *FunctionType
		name     string
		expected bool
	}{
		{
			name:     "identical simple functions",
			ft1:      NewFunctionType([]Type{INTEGER, STRING}, BOOLEAN),
			ft2:      NewFunctionType([]Type{INTEGER, STRING}, BOOLEAN),
			expected: true,
		},
		{
			name:     "different parameter types",
			ft1:      NewFunctionType([]Type{INTEGER, STRING}, BOOLEAN),
			ft2:      NewFunctionType([]Type{INTEGER, INTEGER}, BOOLEAN),
			expected: false,
		},
		{
			name:     "different return types",
			ft1:      NewFunctionType([]Type{INTEGER}, BOOLEAN),
			ft2:      NewFunctionType([]Type{INTEGER}, STRING),
			expected: false,
		},
		{
			name:     "different parameter count",
			ft1:      NewFunctionType([]Type{INTEGER, STRING}, BOOLEAN),
			ft2:      NewFunctionType([]Type{INTEGER}, BOOLEAN),
			expected: false,
		},
		{
			name:     "non-function type",
			ft1:      NewFunctionType([]Type{INTEGER}, BOOLEAN),
			ft2:      INTEGER,
			expected: false,
		},
		{
			name:     "identical variadic functions",
			ft1:      NewVariadicFunctionType([]Type{STRING, NewDynamicArrayType(INTEGER)}, INTEGER, BOOLEAN),
			ft2:      NewVariadicFunctionType([]Type{STRING, NewDynamicArrayType(INTEGER)}, INTEGER, BOOLEAN),
			expected: true,
		},
		{
			name:     "variadic vs non-variadic",
			ft1:      NewVariadicFunctionType([]Type{STRING, NewDynamicArrayType(INTEGER)}, INTEGER, BOOLEAN),
			ft2:      NewFunctionType([]Type{STRING, NewDynamicArrayType(INTEGER)}, BOOLEAN),
			expected: false,
		},
		{
			name:     "different variadic types",
			ft1:      NewVariadicFunctionType([]Type{STRING, NewDynamicArrayType(INTEGER)}, INTEGER, BOOLEAN),
			ft2:      NewVariadicFunctionType([]Type{STRING, NewDynamicArrayType(STRING)}, STRING, BOOLEAN),
			expected: false,
		},
		{
			name: "variadic with nil VariadicType (invalid but defensive)",
			ft1: &FunctionType{
				Parameters:   []Type{NewDynamicArrayType(INTEGER)},
				ReturnType:   VOID,
				IsVariadic:   true,
				VariadicType: nil,
			},
			ft2: &FunctionType{
				Parameters:   []Type{NewDynamicArrayType(INTEGER)},
				ReturnType:   VOID,
				IsVariadic:   true,
				VariadicType: INTEGER,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ft1.Equals(tt.ft2)
			if result != tt.expected {
				t.Errorf("Equals() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFunctionTypeIsProcedure(t *testing.T) {
	tests := []struct {
		funcType *FunctionType
		name     string
		expected bool
	}{
		{
			name:     "procedure (void return)",
			funcType: NewProcedureType([]Type{INTEGER}),
			expected: true,
		},
		{
			name:     "function (non-void return)",
			funcType: NewFunctionType([]Type{INTEGER}, STRING),
			expected: false,
		},
		{
			name:     "variadic procedure",
			funcType: NewVariadicFunctionType([]Type{NewDynamicArrayType(INTEGER)}, INTEGER, VOID),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.funcType.IsProcedure()
			if result != tt.expected {
				t.Errorf("IsProcedure() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFunctionTypeIsFunction(t *testing.T) {
	tests := []struct {
		funcType *FunctionType
		name     string
		expected bool
	}{
		{
			name:     "function (non-void return)",
			funcType: NewFunctionType([]Type{INTEGER}, STRING),
			expected: true,
		},
		{
			name:     "procedure (void return)",
			funcType: NewProcedureType([]Type{INTEGER}),
			expected: false,
		},
		{
			name:     "variadic function",
			funcType: NewVariadicFunctionType([]Type{NewDynamicArrayType(INTEGER)}, INTEGER, STRING),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.funcType.IsFunction()
			if result != tt.expected {
				t.Errorf("IsFunction() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNewVariadicFunctionType(t *testing.T) {
	// Test basic creation
	ft := NewVariadicFunctionType(
		[]Type{STRING, NewDynamicArrayType(INTEGER)},
		INTEGER,
		BOOLEAN,
	)

	if !ft.IsVariadic {
		t.Error("Expected IsVariadic to be true")
	}

	if ft.VariadicType == nil {
		t.Fatal("Expected VariadicType to be set")
	}

	if !ft.VariadicType.Equals(INTEGER) {
		t.Errorf("Expected VariadicType to be Integer, got %v", ft.VariadicType)
	}

	if len(ft.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(ft.Parameters))
	}

	if !ft.ReturnType.Equals(BOOLEAN) {
		t.Errorf("Expected return type Boolean, got %v", ft.ReturnType)
	}

	// Metadata arrays should be initialized
	if len(ft.ParamNames) != 2 {
		t.Errorf("Expected ParamNames length 2, got %d", len(ft.ParamNames))
	}
	if len(ft.LazyParams) != 2 {
		t.Errorf("Expected LazyParams length 2, got %d", len(ft.LazyParams))
	}
	if len(ft.VarParams) != 2 {
		t.Errorf("Expected VarParams length 2, got %d", len(ft.VarParams))
	}
	if len(ft.ConstParams) != 2 {
		t.Errorf("Expected ConstParams length 2, got %d", len(ft.ConstParams))
	}
}

func TestNewVariadicFunctionTypeWithMetadata(t *testing.T) {
	params := []Type{STRING, NewDynamicArrayType(INTEGER)}
	names := []string{"prefix", "values"}
	defaults := []interface{}{nil, nil}
	lazy := []bool{false, false}
	varParams := []bool{false, false}
	constParams := []bool{true, true}

	ft := NewVariadicFunctionTypeWithMetadata(
		params, names, defaults, lazy, varParams, constParams,
		INTEGER,
		BOOLEAN,
	)

	if !ft.IsVariadic {
		t.Error("Expected IsVariadic to be true")
	}

	if ft.VariadicType == nil {
		t.Fatal("Expected VariadicType to be set")
	}

	if !ft.VariadicType.Equals(INTEGER) {
		t.Errorf("Expected VariadicType to be Integer, got %v", ft.VariadicType)
	}

	// Check metadata arrays
	if len(ft.ParamNames) != 2 || ft.ParamNames[0] != "prefix" || ft.ParamNames[1] != "values" {
		t.Errorf("ParamNames not set correctly: %v", ft.ParamNames)
	}

	if !ft.ConstParams[0] || !ft.ConstParams[1] {
		t.Error("Expected both parameters to be const")
	}
}

func TestFunctionTypeTypeKind(t *testing.T) {
	ft := NewFunctionType([]Type{INTEGER}, BOOLEAN)
	if ft.TypeKind() != "FUNCTION" {
		t.Errorf("Expected TypeKind() to return 'FUNCTION', got %q", ft.TypeKind())
	}

	// Variadic function should also return "FUNCTION"
	variadicFt := NewVariadicFunctionType([]Type{NewDynamicArrayType(STRING)}, STRING, VOID)
	if variadicFt.TypeKind() != "FUNCTION" {
		t.Errorf("Expected TypeKind() to return 'FUNCTION' for variadic, got %q", variadicFt.TypeKind())
	}
}

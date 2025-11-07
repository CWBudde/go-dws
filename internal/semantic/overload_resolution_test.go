package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

// TestSignaturesEqual tests signature comparison for overload detection
func TestSignaturesEqual(t *testing.T) {
	tests := []struct {
		name     string
		sig1     *types.FunctionType
		sig2     *types.FunctionType
		expected bool
	}{
		{
			name: "identical signatures",
			sig1: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER, types.STRING},
				ReturnType: types.FLOAT,
			},
			sig2: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER, types.STRING},
				ReturnType: types.FLOAT,
			},
			expected: true,
		},
		{
			name: "same parameters, different return type",
			sig1: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER},
				ReturnType: types.FLOAT,
			},
			sig2: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER},
				ReturnType: types.STRING,
			},
			expected: true, // Return type doesn't matter for signature equality
		},
		{
			name: "different parameter count",
			sig1: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER},
				ReturnType: types.VOID,
			},
			sig2: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER, types.STRING},
				ReturnType: types.VOID,
			},
			expected: false,
		},
		{
			name: "different parameter types",
			sig1: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER, types.STRING},
				ReturnType: types.VOID,
			},
			sig2: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER, types.FLOAT},
				ReturnType: types.VOID,
			},
			expected: false,
		},
		{
			name: "same types, different var modifier",
			sig1: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER},
				VarParams:  []bool{false},
				ReturnType: types.VOID,
			},
			sig2: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER},
				VarParams:  []bool{true}, // var parameter
				ReturnType: types.VOID,
			},
			expected: false,
		},
		{
			name: "same types, different const modifier",
			sig1: &types.FunctionType{
				Parameters:  []types.Type{types.STRING},
				ConstParams: []bool{false},
				ReturnType:  types.VOID,
			},
			sig2: &types.FunctionType{
				Parameters:  []types.Type{types.STRING},
				ConstParams: []bool{true}, // const parameter
				ReturnType:  types.VOID,
			},
			expected: false,
		},
		{
			name: "same types, different lazy modifier",
			sig1: &types.FunctionType{
				Parameters: []types.Type{types.BOOLEAN},
				LazyParams: []bool{false},
				ReturnType: types.VOID,
			},
			sig2: &types.FunctionType{
				Parameters: []types.Type{types.BOOLEAN},
				LazyParams: []bool{true}, // lazy parameter
				ReturnType: types.VOID,
			},
			expected: false,
		},
		{
			name: "both variadic with same element type",
			sig1: &types.FunctionType{
				Parameters:   []types.Type{types.NewDynamicArrayType(types.INTEGER)},
				IsVariadic:   true,
				VariadicType: types.INTEGER,
				ReturnType:   types.VOID,
			},
			sig2: &types.FunctionType{
				Parameters:   []types.Type{types.NewDynamicArrayType(types.INTEGER)},
				IsVariadic:   true,
				VariadicType: types.INTEGER,
				ReturnType:   types.VOID,
			},
			expected: true,
		},
		{
			name: "variadic vs non-variadic",
			sig1: &types.FunctionType{
				Parameters:   []types.Type{types.NewDynamicArrayType(types.INTEGER)},
				IsVariadic:   true,
				VariadicType: types.INTEGER,
				ReturnType:   types.VOID,
			},
			sig2: &types.FunctionType{
				Parameters: []types.Type{types.NewDynamicArrayType(types.INTEGER)},
				IsVariadic: false,
				ReturnType: types.VOID,
			},
			expected: false,
		},
		{
			name: "empty parameter lists",
			sig1: &types.FunctionType{
				Parameters: []types.Type{},
				ReturnType: types.INTEGER,
			},
			sig2: &types.FunctionType{
				Parameters: []types.Type{},
				ReturnType: types.STRING,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SignaturesEqual(tt.sig1, tt.sig2)
			if result != tt.expected {
				t.Errorf("SignaturesEqual() = %v, want %v", result, tt.expected)
				t.Errorf("  sig1: %s", tt.sig1.String())
				t.Errorf("  sig2: %s", tt.sig2.String())
			}
		})
	}
}

// TestSignatureDistance tests distance calculation for overload resolution
func TestSignatureDistance(t *testing.T) {
	tests := []struct {
		name      string
		argTypes  []types.Type
		signature *types.FunctionType
		expected  int // -1 means incompatible
	}{
		{
			name:     "exact match - single parameter",
			argTypes: []types.Type{types.INTEGER},
			signature: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER},
				ReturnType: types.VOID,
			},
			expected: 0,
		},
		{
			name:     "exact match - multiple parameters",
			argTypes: []types.Type{types.INTEGER, types.STRING, types.FLOAT},
			signature: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER, types.STRING, types.FLOAT},
				ReturnType: types.VOID,
			},
			expected: 0,
		},
		{
			name:     "implicit conversion - Integer to Float",
			argTypes: []types.Type{types.INTEGER},
			signature: &types.FunctionType{
				Parameters: []types.Type{types.FLOAT},
				ReturnType: types.VOID,
			},
			expected: 1,
		},
		{
			name:     "multiple conversions",
			argTypes: []types.Type{types.INTEGER, types.INTEGER},
			signature: &types.FunctionType{
				Parameters: []types.Type{types.FLOAT, types.FLOAT},
				ReturnType: types.VOID,
			},
			expected: 2, // Two Integer->Float conversions
		},
		{
			name:     "mixed exact and conversion",
			argTypes: []types.Type{types.INTEGER, types.STRING},
			signature: &types.FunctionType{
				Parameters: []types.Type{types.FLOAT, types.STRING},
				ReturnType: types.VOID,
			},
			expected: 1, // One conversion (Integer->Float), one exact (String)
		},
		{
			name:     "incompatible types",
			argTypes: []types.Type{types.STRING},
			signature: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER},
				ReturnType: types.VOID,
			},
			expected: -1,
		},
		{
			name:     "too few arguments",
			argTypes: []types.Type{types.INTEGER},
			signature: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER, types.STRING},
				ReturnType: types.VOID,
			},
			expected: -1,
		},
		{
			name:     "too many arguments",
			argTypes: []types.Type{types.INTEGER, types.STRING, types.FLOAT},
			signature: &types.FunctionType{
				Parameters: []types.Type{types.INTEGER, types.STRING},
				ReturnType: types.VOID,
			},
			expected: -1,
		},
		{
			name:     "empty arguments and parameters",
			argTypes: []types.Type{},
			signature: &types.FunctionType{
				Parameters: []types.Type{},
				ReturnType: types.VOID,
			},
			expected: 0,
		},
		{
			name:     "variadic with exact match",
			argTypes: []types.Type{types.INTEGER, types.INTEGER, types.INTEGER},
			signature: &types.FunctionType{
				Parameters:   []types.Type{types.INTEGER, types.NewDynamicArrayType(types.INTEGER)},
				IsVariadic:   true,
				VariadicType: types.INTEGER,
				ReturnType:   types.VOID,
			},
			expected: 0, // First arg exact, rest match variadic type
		},
		{
			name:     "variadic with no variadic args",
			argTypes: []types.Type{types.INTEGER},
			signature: &types.FunctionType{
				Parameters:   []types.Type{types.INTEGER, types.NewDynamicArrayType(types.INTEGER)},
				IsVariadic:   true,
				VariadicType: types.INTEGER,
				ReturnType:   types.VOID,
			},
			expected: 0, // Just the required parameter
		},
		{
			name:     "variadic with type conversion",
			argTypes: []types.Type{types.STRING, types.INTEGER, types.INTEGER},
			signature: &types.FunctionType{
				Parameters:   []types.Type{types.STRING, types.NewDynamicArrayType(types.FLOAT)},
				IsVariadic:   true,
				VariadicType: types.FLOAT,
				ReturnType:   types.VOID,
			},
			expected: 2, // Two Integer->Float conversions for variadic args
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SignatureDistance(tt.argTypes, tt.signature)
			if result != tt.expected {
				t.Errorf("SignatureDistance() = %d, want %d", result, tt.expected)
				t.Errorf("  args: %v", formatArgTypes(tt.argTypes))
				t.Errorf("  signature: %s", tt.signature.String())
			}
		})
	}
}

// TestResolveOverload tests overload resolution algorithm
func TestResolveOverload(t *testing.T) {
	// Create some test function types
	funcIntToInt := &types.FunctionType{
		Parameters: []types.Type{types.INTEGER},
		ReturnType: types.INTEGER,
	}
	funcFloatToInt := &types.FunctionType{
		Parameters: []types.Type{types.FLOAT},
		ReturnType: types.INTEGER,
	}
	funcIntIntToInt := &types.FunctionType{
		Parameters: []types.Type{types.INTEGER, types.INTEGER},
		ReturnType: types.INTEGER,
	}
	funcStringToString := &types.FunctionType{
		Parameters: []types.Type{types.STRING},
		ReturnType: types.STRING,
	}

	tests := []struct {
		name        string
		candidates  []*Symbol
		argTypes    []types.Type
		expectError bool
		expectIndex int // Which candidate should be selected (if no error)
	}{
		{
			name: "single candidate - exact match",
			candidates: []*Symbol{
				{Name: "Test", Type: funcIntToInt},
			},
			argTypes:    []types.Type{types.INTEGER},
			expectError: false,
			expectIndex: 0,
		},
		{
			name: "single candidate - with conversion",
			candidates: []*Symbol{
				{Name: "Test", Type: funcFloatToInt},
			},
			argTypes:    []types.Type{types.INTEGER},
			expectError: false,
			expectIndex: 0,
		},
		{
			name: "single candidate - incompatible",
			candidates: []*Symbol{
				{Name: "Test", Type: funcIntToInt},
			},
			argTypes:    []types.Type{types.STRING},
			expectError: true,
		},
		{
			name: "two candidates - exact match wins",
			candidates: []*Symbol{
				{Name: "Test", Type: funcIntToInt},   // Exact match (distance 0)
				{Name: "Test", Type: funcFloatToInt}, // Requires conversion (distance 1)
			},
			argTypes:    []types.Type{types.INTEGER},
			expectError: false,
			expectIndex: 0, // First one is exact match
		},
		{
			name: "two candidates - only one compatible",
			candidates: []*Symbol{
				{Name: "Test", Type: funcIntToInt},
				{Name: "Test", Type: funcStringToString},
			},
			argTypes:    []types.Type{types.INTEGER},
			expectError: false,
			expectIndex: 0,
		},
		{
			name: "two candidates - ambiguous (both exact)",
			candidates: []*Symbol{
				{Name: "Test", Type: funcIntToInt},
				{Name: "Test", Type: &types.FunctionType{
					Parameters: []types.Type{types.INTEGER},
					ReturnType: types.STRING, // Different return type, same signature
				}},
			},
			argTypes:    []types.Type{types.INTEGER},
			expectError: true, // Ambiguous
		},
		{
			name: "different parameter counts",
			candidates: []*Symbol{
				{Name: "Test", Type: funcIntToInt},
				{Name: "Test", Type: funcIntIntToInt},
			},
			argTypes:    []types.Type{types.INTEGER, types.INTEGER},
			expectError: false,
			expectIndex: 1, // Second one matches
		},
		{
			name: "no compatible candidates",
			candidates: []*Symbol{
				{Name: "Test", Type: funcIntToInt},
				{Name: "Test", Type: funcFloatToInt},
			},
			argTypes:    []types.Type{types.STRING},
			expectError: true,
		},
		{
			name:        "no candidates",
			candidates:  []*Symbol{},
			argTypes:    []types.Type{types.INTEGER},
			expectError: true,
		},
		{
			name: "empty arguments - exact match",
			candidates: []*Symbol{
				{Name: "Test", Type: &types.FunctionType{
					Parameters: []types.Type{},
					ReturnType: types.VOID,
				}},
			},
			argTypes:    []types.Type{},
			expectError: false,
			expectIndex: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveOverload(tt.candidates, tt.argTypes)

			if tt.expectError {
				if err == nil {
					t.Errorf("ResolveOverload() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ResolveOverload() unexpected error: %v", err)
				} else if result != tt.candidates[tt.expectIndex] {
					t.Errorf("ResolveOverload() returned wrong candidate")
					t.Errorf("  expected: %s (%p)", tt.candidates[tt.expectIndex].Name, tt.candidates[tt.expectIndex])
					t.Errorf("  got: %s (%p)", result.Name, result)
				}
			}
		})
	}
}

// TestTypeDistance tests individual type distance calculations
func TestTypeDistance(t *testing.T) {
	tests := []struct {
		name     string
		from     types.Type
		to       types.Type
		expected int
	}{
		{
			name:     "exact match - Integer",
			from:     types.INTEGER,
			to:       types.INTEGER,
			expected: 0,
		},
		{
			name:     "exact match - String",
			from:     types.STRING,
			to:       types.STRING,
			expected: 0,
		},
		{
			name:     "Integer to Float",
			from:     types.INTEGER,
			to:       types.FLOAT,
			expected: 1,
		},
		{
			name:     "Float to Integer (not allowed)",
			from:     types.FLOAT,
			to:       types.INTEGER,
			expected: -1,
		},
		{
			name:     "String to Integer (not allowed)",
			from:     types.STRING,
			to:       types.INTEGER,
			expected: -1,
		},
		{
			name:     "Integer to String (not allowed)",
			from:     types.INTEGER,
			to:       types.STRING,
			expected: -1,
		},
		{
			name:     "Boolean to String (not allowed)",
			from:     types.BOOLEAN,
			to:       types.STRING,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := typeDistance(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("typeDistance(%s, %s) = %d, want %d",
					tt.from.String(), tt.to.String(), result, tt.expected)
			}
		})
	}
}

// TestOverloadResolutionWithModifiers tests that parameter modifiers are considered
func TestOverloadResolutionWithModifiers(t *testing.T) {
	// Two overloads: one with var parameter, one without
	funcWithVar := &types.FunctionType{
		Parameters: []types.Type{types.INTEGER},
		VarParams:  []bool{true},
		ReturnType: types.VOID,
	}
	funcWithoutVar := &types.FunctionType{
		Parameters: []types.Type{types.INTEGER},
		VarParams:  []bool{false},
		ReturnType: types.VOID,
	}

	// These should be considered different signatures
	if SignaturesEqual(funcWithVar, funcWithoutVar) {
		t.Error("SignaturesEqual() should return false for different var modifiers")
	}

	// Both should be valid overloads
	candidates := []*Symbol{
		{Name: "Test", Type: funcWithVar},
		{Name: "Test", Type: funcWithoutVar},
	}

	// Calling with INTEGER should match both (ambiguous if no other info)
	_, err := ResolveOverload(candidates, []types.Type{types.INTEGER})
	if err == nil {
		t.Error("Expected ambiguity error for identical parameter types with different modifiers")
	}
}

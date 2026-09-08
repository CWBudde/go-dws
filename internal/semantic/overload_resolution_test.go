package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

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

// TestResolveOverloadTieBreaks verifies deterministic tie-breaking rules.
func TestResolveOverloadTieBreaks(t *testing.T) {
	t.Run("non-variadic beats variadic on equal distance", func(t *testing.T) {
		exact := types.NewFunctionType([]types.Type{types.INTEGER, types.STRING}, types.VOID)
		variadic := types.NewVariadicFunctionType(
			[]types.Type{types.INTEGER, types.STRING, types.NewDynamicArrayType(types.STRING)},
			types.STRING, types.VOID)

		candidates := []*Symbol{
			{Name: "F", Type: variadic},
			{Name: "F", Type: exact},
		}
		selected, err := ResolveOverload(candidates, []types.Type{types.INTEGER, types.STRING})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if selected.Type != exact {
			t.Errorf("expected the non-variadic overload to win, got %v", selected.Type)
		}
	})

	t.Run("nil argument prefers class over dynamic array", func(t *testing.T) {
		base := types.NewClassType("TBase", nil)
		classSig := types.NewFunctionType([]types.Type{base}, types.VOID)
		arraySig := types.NewFunctionType([]types.Type{types.NewDynamicArrayType(base)}, types.VOID)

		candidates := []*Symbol{
			{Name: "F", Type: arraySig},
			{Name: "F", Type: classSig},
		}
		selected, err := ResolveOverload(candidates, []types.Type{types.NIL})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if selected.Type != classSig {
			t.Errorf("expected the class-typed overload to win for nil, got %v", selected.Type)
		}
	})

	t.Run("array argument prefers array-of-variant over variant boxing", func(t *testing.T) {
		variantParam := types.NewFunctionType([]types.Type{types.VARIANT}, types.VOID)
		arrayParam := types.NewFunctionType([]types.Type{types.NewDynamicArrayType(types.VARIANT)}, types.VOID)

		candidates := []*Symbol{
			{Name: "F", Type: variantParam},
			{Name: "F", Type: arrayParam},
		}
		selected, err := ResolveOverload(candidates, []types.Type{types.NewDynamicArrayType(types.INTEGER)})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if selected.Type != arrayParam {
			t.Errorf("expected the array-typed overload to win, got %v", selected.Type)
		}
	})
}

package evaluator

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestEvalEnumBinaryOp tests binary operations on enum values.
func TestEvalEnumBinaryOp(t *testing.T) {
	tests := []struct {
		expected any
		left     *runtime.EnumValue
		right    *runtime.EnumValue
		name     string
		op       string
	}{
		// Comparison operators
		{
			name: "enum equality - equal values",
			op:   "=",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			expected: true,
		},
		{
			name: "enum equality - different values",
			op:   "=",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Green",
				OrdinalValue: 1,
			},
			expected: false,
		},
		{
			name: "enum inequality - different values",
			op:   "<>",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Green",
				OrdinalValue: 1,
			},
			expected: true,
		},
		{
			name: "enum inequality - equal values",
			op:   "<>",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Blue",
				OrdinalValue: 2,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Blue",
				OrdinalValue: 2,
			},
			expected: false,
		},
		{
			name: "enum less than - true",
			op:   "<",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Green",
				OrdinalValue: 1,
			},
			expected: true,
		},
		{
			name: "enum less than - false",
			op:   "<",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Green",
				OrdinalValue: 1,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			expected: false,
		},
		{
			name: "enum greater than - true",
			op:   ">",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Blue",
				OrdinalValue: 2,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			expected: true,
		},
		{
			name: "enum greater than - false",
			op:   ">",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Blue",
				OrdinalValue: 2,
			},
			expected: false,
		},
		{
			name: "enum less than or equal - true (less)",
			op:   "<=",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Green",
				OrdinalValue: 1,
			},
			expected: true,
		},
		{
			name: "enum less than or equal - true (equal)",
			op:   "<=",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			expected: true,
		},
		{
			name: "enum less than or equal - false",
			op:   "<=",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Green",
				OrdinalValue: 1,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			expected: false,
		},
		{
			name: "enum greater than or equal - true (greater)",
			op:   ">=",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Blue",
				OrdinalValue: 2,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			expected: true,
		},
		{
			name: "enum greater than or equal - true (equal)",
			op:   ">=",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			expected: true,
		},
		{
			name: "enum greater than or equal - false",
			op:   ">=",
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Blue",
				OrdinalValue: 2,
			},
			expected: false,
		},
		// Bitwise operations (for flag enums)
		{
			name: "enum bitwise and",
			op:   "and",
			left: &runtime.EnumValue{
				TypeName:     "TFlags",
				ValueName:    "",
				OrdinalValue: 0b1100, // 12
			},
			right: &runtime.EnumValue{
				TypeName:     "TFlags",
				ValueName:    "",
				OrdinalValue: 0b1010, // 10
			},
			expected: 0b1000, // 8
		},
		{
			name: "enum bitwise or",
			op:   "or",
			left: &runtime.EnumValue{
				TypeName:     "TFlags",
				ValueName:    "",
				OrdinalValue: 0b1100, // 12
			},
			right: &runtime.EnumValue{
				TypeName:     "TFlags",
				ValueName:    "",
				OrdinalValue: 0b1010, // 10
			},
			expected: 0b1110, // 14
		},
		{
			name: "enum bitwise xor",
			op:   "xor",
			left: &runtime.EnumValue{
				TypeName:     "TFlags",
				ValueName:    "",
				OrdinalValue: 0b1100, // 12
			},
			right: &runtime.EnumValue{
				TypeName:     "TFlags",
				ValueName:    "",
				OrdinalValue: 0b1010, // 10
			},
			expected: 0b0110, // 6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal evaluator for testing
			typeSystem := interptypes.NewTypeSystem()
			var output bytes.Buffer
			config := DefaultConfig()
			unitRegistry := units.NewUnitRegistry(nil)
			refCountMgr := runtime.NewRefCountManager()
			eval := NewEvaluator(typeSystem, &output, config, unitRegistry, nil, refCountMgr)

			// Create a dummy node for error reporting
			node := &ast.BinaryExpression{
				Left:     &ast.Identifier{Value: "left"},
				Operator: tt.op,
				Right:    &ast.Identifier{Value: "right"},
			}

			// Call evalEnumBinaryOp
			result := eval.evalEnumBinaryOp(tt.op, tt.left, tt.right, node)

			// Check for errors
			if errVal, ok := result.(*runtime.ErrorValue); ok {
				t.Fatalf("unexpected error: %s", errVal.Message)
			}

			// Verify result based on operator type
			if tt.op == "=" || tt.op == "<>" || tt.op == "<" || tt.op == ">" || tt.op == "<=" || tt.op == ">=" {
				// Comparison operators return boolean
				boolResult, ok := result.(*runtime.BooleanValue)
				if !ok {
					t.Fatalf("expected BooleanValue, got %T", result)
				}
				if boolResult.Value != tt.expected.(bool) {
					t.Errorf("expected %v, got %v", tt.expected, boolResult.Value)
				}
			} else {
				// Bitwise operators return enum
				enumResult, ok := result.(*runtime.EnumValue)
				if !ok {
					t.Fatalf("expected EnumValue, got %T", result)
				}
				if enumResult.OrdinalValue != tt.expected.(int) {
					t.Errorf("expected ordinal %v, got %v", tt.expected, enumResult.OrdinalValue)
				}
				if enumResult.TypeName != tt.left.TypeName {
					t.Errorf("expected type name %s, got %s", tt.left.TypeName, enumResult.TypeName)
				}
			}
		})
	}
}

// TestEvalEnumBinaryOpErrors tests error cases for enum binary operations.
func TestEvalEnumBinaryOpErrors(t *testing.T) {
	tests := []struct {
		left        Value
		right       Value
		name        string
		op          string
		expectError bool
	}{
		{
			left: &runtime.IntegerValue{Value: 42},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			name:        "left operand not enum",
			op:          "=",
			expectError: true,
		},
		{
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			right:       &runtime.IntegerValue{Value: 42},
			name:        "right operand not enum",
			op:          "=",
			expectError: true,
		},
		{
			left: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Red",
				OrdinalValue: 0,
			},
			right: &runtime.EnumValue{
				TypeName:     "TColor",
				ValueName:    "Green",
				OrdinalValue: 1,
			},
			name:        "unknown operator",
			op:          "+",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal evaluator for testing
			typeSystem := interptypes.NewTypeSystem()
			var output bytes.Buffer
			config := DefaultConfig()
			unitRegistry := units.NewUnitRegistry(nil)
			refCountMgr := runtime.NewRefCountManager()
			eval := NewEvaluator(typeSystem, &output, config, unitRegistry, nil, refCountMgr)

			// Create a dummy node for error reporting
			node := &ast.BinaryExpression{
				Left:     &ast.Identifier{Value: "left"},
				Operator: tt.op,
				Right:    &ast.Identifier{Value: "right"},
			}

			// Call evalEnumBinaryOp
			result := eval.evalEnumBinaryOp(tt.op, tt.left, tt.right, node)

			// Check if error was returned as expected
			if tt.expectError {
				if _, ok := result.(*runtime.ErrorValue); !ok {
					t.Errorf("expected error, got %T: %v", result, result)
				}
			} else {
				if errVal, ok := result.(*runtime.ErrorValue); ok {
					t.Errorf("unexpected error: %s", errVal.Message)
				}
			}
		})
	}
}

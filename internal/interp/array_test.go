package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// ArrayValue Tests
// ============================================================================

// TestArrayValue_Creation tests creating an ArrayValue for both static and dynamic arrays.
func TestArrayValue_Creation(t *testing.T) {
	t.Run("dynamic array creation", func(t *testing.T) {
		// Create a dynamic array type: array of Integer
		elementType := types.INTEGER
		arrayType := types.NewDynamicArrayType(elementType)

		// Create an empty dynamic array
		arr := NewArrayValue(arrayType)

		// Verify Type() returns "ARRAY"
		if arr.Type() != "ARRAY" {
			t.Errorf("expected Type() = 'ARRAY', got '%s'", arr.Type())
		}

		// Verify it's empty
		if len(arr.Elements) != 0 {
			t.Errorf("expected empty array, got %d elements", len(arr.Elements))
		}

		// Verify String() for empty array
		if arr.String() != "[]" {
			t.Errorf("expected String() = '[]', got '%s'", arr.String())
		}
	})

	t.Run("static array creation", func(t *testing.T) {
		// Create a static array type: array[1..5] of Integer
		elementType := types.INTEGER
		lowBound := 1
		highBound := 5
		arrayType := types.NewStaticArrayType(elementType, lowBound, highBound)

		// Create a static array (should be pre-allocated with 5 elements)
		arr := NewArrayValue(arrayType)

		// Verify Type() returns "ARRAY"
		if arr.Type() != "ARRAY" {
			t.Errorf("expected Type() = 'ARRAY', got '%s'", arr.Type())
		}

		// Verify it has 5 elements (initialized to nil/zero values)
		expectedSize := highBound - lowBound + 1
		if len(arr.Elements) != expectedSize {
			t.Errorf("expected %d elements, got %d", expectedSize, len(arr.Elements))
		}
	})
}

// TestArrayValue_WithElements tests ArrayValue with pre-set elements.
func TestArrayValue_WithElements(t *testing.T) {
	// Create a dynamic array type: array of Integer
	elementType := types.INTEGER
	arrayType := types.NewDynamicArrayType(elementType)

	// Create an array with some elements
	arr := &ArrayValue{
		ArrayType: arrayType,
		Elements: []Value{
			&IntegerValue{Value: 10},
			&IntegerValue{Value: 20},
			&IntegerValue{Value: 30},
		},
	}

	// Verify Type()
	if arr.Type() != "ARRAY" {
		t.Errorf("expected Type() = 'ARRAY', got '%s'", arr.Type())
	}

	// Verify String() shows elements
	str := arr.String()
	expected := "[10, 20, 30]"
	if str != expected {
		t.Errorf("expected String() = '%s', got '%s'", expected, str)
	}

	// Verify element count
	if len(arr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Elements))
	}
}

// ============================================================================
// Array Indexing Tests
// ============================================================================

// TestArrayDeclaration_Basic tests that array type declarations work.
func TestArrayDeclaration_Basic(t *testing.T) {
	input := `
		type TIntArray = array[0..2] of Integer;
	`

	result := testEval(input)
	// Type declarations return nil
	if _, ok := result.(*NilValue); !ok {
		t.Errorf("expected NilValue from type declaration, got %T: %v", result, result)
	}
}

// TestArrayReturnType_High tests that functions returning arrays
// properly initialize the Result variable so .High works.
// This is Task 9.218 - fixing primes.pas High helper error.
func TestArrayReturnType_High(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "empty array .High",
			input: `
				type TIntArray = array of Integer;

				function GetEmptyArray(): TIntArray;
				begin
					// Result is auto-initialized to empty array
				end;

				var arr := GetEmptyArray();
				PrintLn(arr.High);
			`,
			expected: "-1\n", // Empty array has High = -1
		},
		{
			name: "array with elements .High",
			input: `
				type TIntArray = array of Integer;

				function GetArray(): TIntArray;
				begin
					Result.Add(10);
					Result.Add(20);
					Result.Add(30);
				end;

				var arr := GetArray();
				PrintLn(arr.High);
			`,
			expected: "2\n", // 3 elements, High = 2
		},
		{
			name: "array .Low",
			input: `
				type TIntArray = array of Integer;

				function GetArray(): TIntArray;
				begin
					Result.Add(100);
				end;

				var arr := GetArray();
				PrintLn(arr.Low);
			`,
			expected: "0\n", // Dynamic arrays always have Low = 0
		},
		{
			name: "use .High in for loop like primes.pas",
			input: `
				type TIntArray = array of Integer;

				function GetNumbers(): TIntArray;
				var i: Integer;
				begin
					for i := 1 to 5 do
						Result.Add(i * 10);
				end;

				var r := GetNumbers();
				var i: Integer;
				var sum := 0;
				for i := 0 to r.High do
					sum := sum + r[i];
				PrintLn(sum);
			`,
			expected: "150\n", // 10 + 20 + 30 + 40 + 50 = 150
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			interp := New(&out)
			result := interpret(interp, tt.input)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

package runtime

import "testing"

// TestArrayHelperIndexOf_ClampsNegativeStart verifies that a fromIndex below
// the array's low bound is clamped to the start of the array instead of
// reporting "not found" (see fixture ArrayPass/indexof_from_static:
// a.IndexOf(1, -5) = 1).
func TestArrayHelperIndexOf_ClampsNegativeStart(t *testing.T) {
	arr := &ArrayValue{
		Elements: []Value{
			&IntegerValue{Value: 0},
			&IntegerValue{Value: 1},
			&IntegerValue{Value: 0},
			&IntegerValue{Value: 1},
		},
	}

	tests := []struct {
		name     string
		needle   int64
		start    int
		expected int64
	}{
		{"negative start clamps to zero", 1, -5, 1},
		{"start zero", 1, 0, 1},
		{"start past first match", 1, 2, 3},
		{"start beyond end", 1, 5, -1},
		{"not found", 7, -3, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrayHelperIndexOf(arr, &IntegerValue{Value: tt.needle}, tt.start)
			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("ArrayHelperIndexOf returned %T, want IntegerValue", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("ArrayHelperIndexOf(%d, %d) = %d, want %d", tt.needle, tt.start, intVal.Value, tt.expected)
			}
		})
	}
}

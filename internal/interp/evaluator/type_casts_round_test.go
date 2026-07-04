package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// TestCastToInteger_FloatRounding verifies that Integer(<float>) rounds with
// banker's rounding (half to even), matching DWScript/Delphi Round semantics
// (see fixture casts_base_types: 2 + Integer(1.5) = 4).
func TestCastToInteger_FloatRounding(t *testing.T) {
	e := &Evaluator{}

	tests := []struct {
		in       float64
		expected int64
	}{
		{1.5, 2},
		{2.5, 2},
		{0.5, 0},
		{-0.5, 0},
		{-1.5, -2},
		{1.25, 1},
		{1.75, 2},
		{3.0, 3},
	}

	for _, tt := range tests {
		result := e.castToInteger(&runtime.FloatValue{Value: tt.in})
		intVal, ok := result.(*runtime.IntegerValue)
		if !ok {
			t.Fatalf("castToInteger(%v) returned %T, want IntegerValue", tt.in, result)
		}
		if intVal.Value != tt.expected {
			t.Errorf("castToInteger(%v) = %d, want %d", tt.in, intVal.Value, tt.expected)
		}
	}
}

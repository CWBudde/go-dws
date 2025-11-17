package builtins

import (
	"math"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Math Conversion Functions Tests
// =============================================================================

func TestDegToRad(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
	}{
		{
			name:      "0 degrees",
			args:      []Value{&runtime.FloatValue{Value: 0}},
			expected:  0.0,
			tolerance: 1e-9,
		},
		{
			name:      "180 degrees",
			args:      []Value{&runtime.FloatValue{Value: 180}},
			expected:  math.Pi,
			tolerance: 1e-9,
		},
		{
			name:      "90 degrees",
			args:      []Value{&runtime.FloatValue{Value: 90}},
			expected:  math.Pi / 2,
			tolerance: 1e-9,
		},
		{
			name:      "360 degrees",
			args:      []Value{&runtime.FloatValue{Value: 360}},
			expected:  2 * math.Pi,
			tolerance: 1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DegToRad(ctx, tt.args)
			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("DegToRad() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestRadToDeg(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
	}{
		{
			name:      "0 radians",
			args:      []Value{&runtime.FloatValue{Value: 0}},
			expected:  0.0,
			tolerance: 1e-9,
		},
		{
			name:      "π radians",
			args:      []Value{&runtime.FloatValue{Value: math.Pi}},
			expected:  180.0,
			tolerance: 1e-9,
		},
		{
			name:      "π/2 radians",
			args:      []Value{&runtime.FloatValue{Value: math.Pi / 2}},
			expected:  90.0,
			tolerance: 1e-9,
		},
		{
			name:      "2π radians",
			args:      []Value{&runtime.FloatValue{Value: 2 * math.Pi}},
			expected:  360.0,
			tolerance: 1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RadToDeg(ctx, tt.args)
			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("RadToDeg() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestRound(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
	}{
		{
			name:     "round 3.4",
			args:     []Value{&runtime.FloatValue{Value: 3.4}},
			expected: &runtime.IntegerValue{Value: 3},
		},
		{
			name:     "round 3.5",
			args:     []Value{&runtime.FloatValue{Value: 3.5}},
			expected: &runtime.IntegerValue{Value: 4},
		},
		{
			name:     "round 3.9",
			args:     []Value{&runtime.FloatValue{Value: 3.9}},
			expected: &runtime.IntegerValue{Value: 4},
		},
		{
			name:     "round -3.5",
			args:     []Value{&runtime.FloatValue{Value: -3.5}},
			expected: &runtime.IntegerValue{Value: -4},
		},
		{
			name:     "round 0.0",
			args:     []Value{&runtime.FloatValue{Value: 0.0}},
			expected: &runtime.IntegerValue{Value: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Round(ctx, tt.args)

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Round() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTrunc(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
	}{
		{
			name:     "trunc 3.9",
			args:     []Value{&runtime.FloatValue{Value: 3.9}},
			expected: &runtime.IntegerValue{Value: 3},
		},
		{
			name:     "trunc -3.9",
			args:     []Value{&runtime.FloatValue{Value: -3.9}},
			expected: &runtime.IntegerValue{Value: -3},
		},
		{
			name:     "trunc 0.0",
			args:     []Value{&runtime.FloatValue{Value: 0.0}},
			expected: &runtime.IntegerValue{Value: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Trunc(ctx, tt.args)

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Trunc() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCeil(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
	}{
		{
			name:     "ceil 3.1",
			args:     []Value{&runtime.FloatValue{Value: 3.1}},
			expected: &runtime.IntegerValue{Value: 4},
		},
		{
			name:     "ceil 3.0",
			args:     []Value{&runtime.FloatValue{Value: 3.0}},
			expected: &runtime.IntegerValue{Value: 3},
		},
		{
			name:     "ceil -3.1",
			args:     []Value{&runtime.FloatValue{Value: -3.1}},
			expected: &runtime.IntegerValue{Value: -3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Ceil(ctx, tt.args)

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Ceil() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFloor(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
	}{
		{
			name:     "floor 3.9",
			args:     []Value{&runtime.FloatValue{Value: 3.9}},
			expected: &runtime.IntegerValue{Value: 3},
		},
		{
			name:     "floor 3.0",
			args:     []Value{&runtime.FloatValue{Value: 3.0}},
			expected: &runtime.IntegerValue{Value: 3},
		},
		{
			name:     "floor -3.1",
			args:     []Value{&runtime.FloatValue{Value: -3.1}},
			expected: &runtime.IntegerValue{Value: -4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Floor(ctx, tt.args)

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Floor() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClampInt(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
	}{
		{
			name: "within range",
			args: []Value{
				&runtime.IntegerValue{Value: 5},
				&runtime.IntegerValue{Value: 0},
				&runtime.IntegerValue{Value: 10},
			},
			expected: &runtime.IntegerValue{Value: 5},
		},
		{
			name: "below minimum",
			args: []Value{
				&runtime.IntegerValue{Value: -5},
				&runtime.IntegerValue{Value: 0},
				&runtime.IntegerValue{Value: 10},
			},
			expected: &runtime.IntegerValue{Value: 0},
		},
		{
			name: "above maximum",
			args: []Value{
				&runtime.IntegerValue{Value: 15},
				&runtime.IntegerValue{Value: 0},
				&runtime.IntegerValue{Value: 10},
			},
			expected: &runtime.IntegerValue{Value: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClampInt(ctx, tt.args)

			if !valuesEqual(result, tt.expected) {
				t.Errorf("ClampInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
	}{
		{
			name: "within range",
			args: []Value{
				&runtime.FloatValue{Value: 5.5},
				&runtime.FloatValue{Value: 0.0},
				&runtime.FloatValue{Value: 10.0},
			},
			expected:  5.5,
			tolerance: 1e-9,
		},
		{
			name: "below minimum",
			args: []Value{
				&runtime.FloatValue{Value: -5.0},
				&runtime.FloatValue{Value: 0.0},
				&runtime.FloatValue{Value: 10.0},
			},
			expected:  0.0,
			tolerance: 1e-9,
		},
		{
			name: "above maximum",
			args: []Value{
				&runtime.FloatValue{Value: 15.0},
				&runtime.FloatValue{Value: 0.0},
				&runtime.FloatValue{Value: 10.0},
			},
			expected:  10.0,
			tolerance: 1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clamp(ctx, tt.args)
			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("Clamp() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestFrac(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
	}{
		{
			name:      "frac 3.14",
			args:      []Value{&runtime.FloatValue{Value: 3.14}},
			expected:  0.14,
			tolerance: 1e-9,
		},
		{
			name:      "frac -3.14",
			args:      []Value{&runtime.FloatValue{Value: -3.14}},
			expected:  -0.14,
			tolerance: 1e-9,
		},
		{
			name:      "frac 5.0",
			args:      []Value{&runtime.FloatValue{Value: 5.0}},
			expected:  0.0,
			tolerance: 1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Frac(ctx, tt.args)
			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("Frac() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestInt(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
	}{
		{
			name:      "int of 3.9",
			args:      []Value{&runtime.FloatValue{Value: 3.9}},
			expected:  3.0,
			tolerance: 1e-9,
		},
		{
			name:      "int of -3.9",
			args:      []Value{&runtime.FloatValue{Value: -3.9}},
			expected:  -3.0,
			tolerance: 1e-9,
		},
		{
			name:      "int of 0.0",
			args:      []Value{&runtime.FloatValue{Value: 0.0}},
			expected:  0.0,
			tolerance: 1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Int(ctx, tt.args)
			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("Int() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestIntPower(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
	}{
		{
			name: "2^3",
			args: []Value{
				&runtime.IntegerValue{Value: 2},
				&runtime.IntegerValue{Value: 3},
			},
			expected:  8.0,
			tolerance: 1e-9,
		},
		{
			name: "10^0",
			args: []Value{
				&runtime.IntegerValue{Value: 10},
				&runtime.IntegerValue{Value: 0},
			},
			expected:  1.0,
			tolerance: 1e-9,
		},
		{
			name: "5^2",
			args: []Value{
				&runtime.IntegerValue{Value: 5},
				&runtime.IntegerValue{Value: 2},
			},
			expected:  25.0,
			tolerance: 1e-9,
		},
		{
			name: "2^10",
			args: []Value{
				&runtime.IntegerValue{Value: 2},
				&runtime.IntegerValue{Value: 10},
			},
			expected:  1024.0,
			tolerance: 1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntPower(ctx, tt.args)
			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("IntPower() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

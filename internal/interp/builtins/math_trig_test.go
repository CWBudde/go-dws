package builtins

import (
	"math"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Trigonometric Functions Tests
// =============================================================================

func TestSin(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
	}{
		{
			name:      "sin(0)",
			args:      []Value{&runtime.FloatValue{Value: 0}},
			expected:  0.0,
			tolerance: 1e-9,
		},
		{
			name:      "sin(π/2)",
			args:      []Value{&runtime.FloatValue{Value: math.Pi / 2}},
			expected:  1.0,
			tolerance: 1e-9,
		},
		{
			name:      "sin(π)",
			args:      []Value{&runtime.FloatValue{Value: math.Pi}},
			expected:  0.0,
			tolerance: 1e-9,
		},
		{
			name:      "works with integer",
			args:      []Value{&runtime.IntegerValue{Value: 0}},
			expected:  0.0,
			tolerance: 1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sin(ctx, tt.args)
			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("Sin() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestCos(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
	}{
		{
			name:      "cos(0)",
			args:      []Value{&runtime.FloatValue{Value: 0}},
			expected:  1.0,
			tolerance: 1e-9,
		},
		{
			name:      "cos(π/2)",
			args:      []Value{&runtime.FloatValue{Value: math.Pi / 2}},
			expected:  0.0,
			tolerance: 1e-9,
		},
		{
			name:      "cos(π)",
			args:      []Value{&runtime.FloatValue{Value: math.Pi}},
			expected:  -1.0,
			tolerance: 1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Cos(ctx, tt.args)
			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("Cos() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestTan(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
	}{
		{
			name:      "tan(0)",
			args:      []Value{&runtime.FloatValue{Value: 0}},
			expected:  0.0,
			tolerance: 1e-9,
		},
		{
			name:      "tan(π/4)",
			args:      []Value{&runtime.FloatValue{Value: math.Pi / 4}},
			expected:  1.0,
			tolerance: 1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Tan(ctx, tt.args)
			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("Tan() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestArcSin(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
		isError   bool
	}{
		{
			name:      "arcsin(0)",
			args:      []Value{&runtime.FloatValue{Value: 0}},
			expected:  0.0,
			tolerance: 1e-9,
		},
		{
			name:      "arcsin(1)",
			args:      []Value{&runtime.FloatValue{Value: 1}},
			expected:  math.Pi / 2,
			tolerance: 1e-9,
		},
		{
			name:      "arcsin(-1)",
			args:      []Value{&runtime.FloatValue{Value: -1}},
			expected:  -math.Pi / 2,
			tolerance: 1e-9,
		},
		{
			name:    "out of range",
			args:    []Value{&runtime.FloatValue{Value: 2}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArcSin(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("ArcSin() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestArcCos(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
		isError   bool
	}{
		{
			name:      "arccos(1)",
			args:      []Value{&runtime.FloatValue{Value: 1}},
			expected:  0.0,
			tolerance: 1e-9,
		},
		{
			name:      "arccos(0)",
			args:      []Value{&runtime.FloatValue{Value: 0}},
			expected:  math.Pi / 2,
			tolerance: 1e-9,
		},
		{
			name:      "arccos(-1)",
			args:      []Value{&runtime.FloatValue{Value: -1}},
			expected:  math.Pi,
			tolerance: 1e-9,
		},
		{
			name:    "out of range",
			args:    []Value{&runtime.FloatValue{Value: 2}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArcCos(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("ArcCos() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestArcTan(t *testing.T) {
	ctx := newMockContext()

	result := ArcTan(ctx, []Value{&runtime.FloatValue{Value: 0}})
	floatResult, ok := result.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("expected FloatValue, got %T", result)
	}
	if math.Abs(floatResult.Value) > 1e-9 {
		t.Errorf("ArcTan(0) = %v, want 0.0", floatResult.Value)
	}

	result = ArcTan(ctx, []Value{&runtime.FloatValue{Value: 1}})
	floatResult, ok = result.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("expected FloatValue, got %T", result)
	}
	if math.Abs(floatResult.Value-math.Pi/4) > 1e-9 {
		t.Errorf("ArcTan(1) = %v, want π/4", floatResult.Value)
	}
}

func TestArcTan2(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
	}{
		{
			name: "arctan2(0, 1)",
			args: []Value{
				&runtime.FloatValue{Value: 0},
				&runtime.FloatValue{Value: 1},
			},
			expected:  0.0,
			tolerance: 1e-9,
		},
		{
			name: "arctan2(1, 0)",
			args: []Value{
				&runtime.FloatValue{Value: 1},
				&runtime.FloatValue{Value: 0},
			},
			expected:  math.Pi / 2,
			tolerance: 1e-9,
		},
		{
			name: "arctan2(1, 1)",
			args: []Value{
				&runtime.FloatValue{Value: 1},
				&runtime.FloatValue{Value: 1},
			},
			expected:  math.Pi / 4,
			tolerance: 1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArcTan2(ctx, tt.args)
			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("ArcTan2() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestCoTan(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
		isError   bool
	}{
		{
			name:      "cotan(π/4)",
			args:      []Value{&runtime.FloatValue{Value: math.Pi / 4}},
			expected:  1.0,
			tolerance: 1e-9,
		},
		{
			name:    "cotan(0) - error",
			args:    []Value{&runtime.FloatValue{Value: 0}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CoTan(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("CoTan() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestHypot(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
	}{
		{
			name: "hypot(3, 4)",
			args: []Value{
				&runtime.FloatValue{Value: 3},
				&runtime.FloatValue{Value: 4},
			},
			expected:  5.0,
			tolerance: 1e-9,
		},
		{
			name: "hypot(0, 0)",
			args: []Value{
				&runtime.FloatValue{Value: 0},
				&runtime.FloatValue{Value: 0},
			},
			expected:  0.0,
			tolerance: 1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hypot(ctx, tt.args)
			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("Hypot() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestSinh(t *testing.T) {
	ctx := newMockContext()

	result := Sinh(ctx, []Value{&runtime.FloatValue{Value: 0}})
	floatResult, ok := result.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("expected FloatValue, got %T", result)
	}
	if math.Abs(floatResult.Value) > 1e-9 {
		t.Errorf("Sinh(0) = %v, want 0.0", floatResult.Value)
	}
}

func TestCosh(t *testing.T) {
	ctx := newMockContext()

	result := Cosh(ctx, []Value{&runtime.FloatValue{Value: 0}})
	floatResult, ok := result.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("expected FloatValue, got %T", result)
	}
	if math.Abs(floatResult.Value-1.0) > 1e-9 {
		t.Errorf("Cosh(0) = %v, want 1.0", floatResult.Value)
	}
}

func TestTanh(t *testing.T) {
	ctx := newMockContext()

	result := Tanh(ctx, []Value{&runtime.FloatValue{Value: 0}})
	floatResult, ok := result.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("expected FloatValue, got %T", result)
	}
	if math.Abs(floatResult.Value) > 1e-9 {
		t.Errorf("Tanh(0) = %v, want 0.0", floatResult.Value)
	}
}

func TestArcSinh(t *testing.T) {
	ctx := newMockContext()

	result := ArcSinh(ctx, []Value{&runtime.FloatValue{Value: 0}})
	floatResult, ok := result.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("expected FloatValue, got %T", result)
	}
	if math.Abs(floatResult.Value) > 1e-9 {
		t.Errorf("ArcSinh(0) = %v, want 0.0", floatResult.Value)
	}
}

func TestArcCosh(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
		isError   bool
	}{
		{
			name:      "arccosh(1)",
			args:      []Value{&runtime.FloatValue{Value: 1}},
			expected:  0.0,
			tolerance: 1e-9,
		},
		{
			name:    "arccosh(0) - error",
			args:    []Value{&runtime.FloatValue{Value: 0}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArcCosh(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("ArcCosh() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

func TestArcTanh(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64
		tolerance float64
		isError   bool
	}{
		{
			name:      "arctanh(0)",
			args:      []Value{&runtime.FloatValue{Value: 0}},
			expected:  0.0,
			tolerance: 1e-9,
		},
		{
			name:    "arctanh(2) - error",
			args:    []Value{&runtime.FloatValue{Value: 2}},
			isError: true,
		},
		{
			name:    "arctanh(-2) - error",
			args:    []Value{&runtime.FloatValue{Value: -2}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArcTanh(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			floatResult, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if math.Abs(floatResult.Value-tt.expected) > tt.tolerance {
				t.Errorf("ArcTanh() = %v, want %v", floatResult.Value, tt.expected)
			}
		})
	}
}

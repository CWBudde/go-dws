package builtins

import (
	"math"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Basic Math Functions Tests
// =============================================================================

func TestAbs(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name:     "positive integer",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: &runtime.IntegerValue{Value: 42},
		},
		{
			name:     "negative integer",
			args:     []Value{&runtime.IntegerValue{Value: -42}},
			expected: &runtime.IntegerValue{Value: 42},
		},
		{
			name:     "zero integer",
			args:     []Value{&runtime.IntegerValue{Value: 0}},
			expected: &runtime.IntegerValue{Value: 0},
		},
		{
			name:     "positive float",
			args:     []Value{&runtime.FloatValue{Value: 3.14}},
			expected: &runtime.FloatValue{Value: 3.14},
		},
		{
			name:     "negative float",
			args:     []Value{&runtime.FloatValue{Value: -3.14}},
			expected: &runtime.FloatValue{Value: 3.14},
		},
		{
			name:    "no arguments",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "too many arguments",
			args:    []Value{&runtime.IntegerValue{Value: 1}, &runtime.IntegerValue{Value: 2}},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "hello"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Abs(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Abs() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMin(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name: "two integers - first smaller",
			args: []Value{
				&runtime.IntegerValue{Value: 10},
				&runtime.IntegerValue{Value: 20},
			},
			expected: &runtime.IntegerValue{Value: 10},
		},
		{
			name: "two integers - second smaller",
			args: []Value{
				&runtime.IntegerValue{Value: 30},
				&runtime.IntegerValue{Value: 15},
			},
			expected: &runtime.IntegerValue{Value: 15},
		},
		{
			name: "two floats - first smaller",
			args: []Value{
				&runtime.FloatValue{Value: 1.5},
				&runtime.FloatValue{Value: 2.5},
			},
			expected: &runtime.FloatValue{Value: 1.5},
		},
		{
			name: "integer and float - integer smaller",
			args: []Value{
				&runtime.IntegerValue{Value: 5},
				&runtime.FloatValue{Value: 10.0},
			},
			expected: &runtime.FloatValue{Value: 5.0},
		},
		{
			name: "float and integer - integer smaller",
			args: []Value{
				&runtime.FloatValue{Value: 10.5},
				&runtime.IntegerValue{Value: 5},
			},
			expected: &runtime.FloatValue{Value: 5.0},
		},
		{
			name:    "no arguments",
			args:    []Value{},
			isError: true,
		},
		{
			name: "one argument",
			args: []Value{
				&runtime.IntegerValue{Value: 5},
			},
			isError: true,
		},
		{
			name: "wrong type",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.IntegerValue{Value: 5},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Min(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Min() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMax(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name: "two integers - first larger",
			args: []Value{
				&runtime.IntegerValue{Value: 20},
				&runtime.IntegerValue{Value: 10},
			},
			expected: &runtime.IntegerValue{Value: 20},
		},
		{
			name: "two integers - second larger",
			args: []Value{
				&runtime.IntegerValue{Value: 15},
				&runtime.IntegerValue{Value: 30},
			},
			expected: &runtime.IntegerValue{Value: 30},
		},
		{
			name: "two floats - first larger",
			args: []Value{
				&runtime.FloatValue{Value: 2.5},
				&runtime.FloatValue{Value: 1.5},
			},
			expected: &runtime.FloatValue{Value: 2.5},
		},
		{
			name: "integer and float - float larger",
			args: []Value{
				&runtime.IntegerValue{Value: 5},
				&runtime.FloatValue{Value: 10.0},
			},
			expected: &runtime.FloatValue{Value: 10.0},
		},
		{
			name: "float and integer - float larger",
			args: []Value{
				&runtime.FloatValue{Value: 10.5},
				&runtime.IntegerValue{Value: 5},
			},
			expected: &runtime.FloatValue{Value: 10.5},
		},
		{
			name:    "no arguments",
			args:    []Value{},
			isError: true,
		},
		{
			name: "wrong type",
			args: []Value{
				&runtime.IntegerValue{Value: 5},
				&runtime.StringValue{Value: "hello"},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Max(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Max() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSqr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name:     "positive integer",
			args:     []Value{&runtime.IntegerValue{Value: 5}},
			expected: &runtime.IntegerValue{Value: 25},
		},
		{
			name:     "negative integer",
			args:     []Value{&runtime.IntegerValue{Value: -4}},
			expected: &runtime.IntegerValue{Value: 16},
		},
		{
			name:     "zero",
			args:     []Value{&runtime.IntegerValue{Value: 0}},
			expected: &runtime.IntegerValue{Value: 0},
		},
		{
			name:     "float",
			args:     []Value{&runtime.FloatValue{Value: 2.5}},
			expected: &runtime.FloatValue{Value: 6.25},
		},
		{
			name:    "no arguments",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "hello"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sqr(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Sqr() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPower(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name: "integer base and exponent",
			args: []Value{
				&runtime.IntegerValue{Value: 2},
				&runtime.IntegerValue{Value: 3},
			},
			expected: &runtime.FloatValue{Value: 8.0},
		},
		{
			name: "float base and integer exponent",
			args: []Value{
				&runtime.FloatValue{Value: 2.0},
				&runtime.IntegerValue{Value: 3},
			},
			expected: &runtime.FloatValue{Value: 8.0},
		},
		{
			name: "zero to the power of zero",
			args: []Value{
				&runtime.IntegerValue{Value: 0},
				&runtime.IntegerValue{Value: 0},
			},
			expected: &runtime.FloatValue{Value: 1.0},
		},
		{
			name: "negative exponent",
			args: []Value{
				&runtime.IntegerValue{Value: 2},
				&runtime.IntegerValue{Value: -2},
			},
			expected: &runtime.FloatValue{Value: 0.25},
		},
		{
			name:    "no arguments",
			args:    []Value{},
			isError: true,
		},
		{
			name: "wrong base type",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.IntegerValue{Value: 2},
			},
			isError: true,
		},
		{
			name: "wrong exponent type",
			args: []Value{
				&runtime.IntegerValue{Value: 2},
				&runtime.StringValue{Value: "hello"},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Power(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Power() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSqrt(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name:     "positive integer",
			args:     []Value{&runtime.IntegerValue{Value: 16}},
			expected: &runtime.FloatValue{Value: 4.0},
		},
		{
			name:     "zero",
			args:     []Value{&runtime.IntegerValue{Value: 0}},
			expected: &runtime.FloatValue{Value: 0.0},
		},
		{
			name:     "positive float",
			args:     []Value{&runtime.FloatValue{Value: 2.25}},
			expected: &runtime.FloatValue{Value: 1.5},
		},
		{
			name:    "negative integer",
			args:    []Value{&runtime.IntegerValue{Value: -4}},
			isError: true,
		},
		{
			name:    "negative float",
			args:    []Value{&runtime.FloatValue{Value: -2.5}},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "hello"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sqrt(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Sqrt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExp(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name:     "zero",
			args:     []Value{&runtime.IntegerValue{Value: 0}},
			expected: &runtime.FloatValue{Value: 1.0},
		},
		{
			name:     "positive integer",
			args:     []Value{&runtime.IntegerValue{Value: 1}},
			expected: &runtime.FloatValue{Value: math.E},
		},
		{
			name:     "float value",
			args:     []Value{&runtime.FloatValue{Value: 2.0}},
			expected: &runtime.FloatValue{Value: math.Exp(2.0)},
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "hello"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Exp(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Exp() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLn(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name:     "positive integer",
			args:     []Value{&runtime.IntegerValue{Value: 1}},
			expected: &runtime.FloatValue{Value: 0.0},
		},
		{
			name:     "e",
			args:     []Value{&runtime.FloatValue{Value: math.E}},
			expected: &runtime.FloatValue{Value: 1.0},
		},
		{
			name:    "zero",
			args:    []Value{&runtime.IntegerValue{Value: 0}},
			isError: true,
		},
		{
			name:    "negative",
			args:    []Value{&runtime.IntegerValue{Value: -5}},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.StringValue{Value: "hello"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Ln(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Ln() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLog2(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name:     "power of 2",
			args:     []Value{&runtime.IntegerValue{Value: 8}},
			expected: &runtime.FloatValue{Value: 3.0},
		},
		{
			name:     "one",
			args:     []Value{&runtime.IntegerValue{Value: 1}},
			expected: &runtime.FloatValue{Value: 0.0},
		},
		{
			name:    "zero",
			args:    []Value{&runtime.IntegerValue{Value: 0}},
			isError: true,
		},
		{
			name:    "negative",
			args:    []Value{&runtime.IntegerValue{Value: -2}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Log2(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Log2() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLog10(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name:     "power of 10",
			args:     []Value{&runtime.IntegerValue{Value: 100}},
			expected: &runtime.FloatValue{Value: 2.0},
		},
		{
			name:     "one",
			args:     []Value{&runtime.IntegerValue{Value: 1}},
			expected: &runtime.FloatValue{Value: 0.0},
		},
		{
			name:    "zero",
			args:    []Value{&runtime.IntegerValue{Value: 0}},
			isError: true,
		},
		{
			name:    "negative",
			args:    []Value{&runtime.FloatValue{Value: -10.0}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Log10(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Log10() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLogN(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name: "log base 2 of 8",
			args: []Value{
				&runtime.IntegerValue{Value: 8},
				&runtime.IntegerValue{Value: 2},
			},
			expected: &runtime.FloatValue{Value: 3.0},
		},
		{
			name: "log base 10 of 100",
			args: []Value{
				&runtime.IntegerValue{Value: 100},
				&runtime.IntegerValue{Value: 10},
			},
			expected: &runtime.FloatValue{Value: 2.0},
		},
		{
			name: "negative value",
			args: []Value{
				&runtime.IntegerValue{Value: -10},
				&runtime.IntegerValue{Value: 2},
			},
			isError: true,
		},
		{
			name: "negative base",
			args: []Value{
				&runtime.IntegerValue{Value: 10},
				&runtime.IntegerValue{Value: -2},
			},
			isError: true,
		},
		{
			name: "base 1",
			args: []Value{
				&runtime.IntegerValue{Value: 10},
				&runtime.IntegerValue{Value: 1},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LogN(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("LogN() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUnsigned32(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name:     "positive value",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: &runtime.IntegerValue{Value: 42},
		},
		{
			name:     "negative value wraps around",
			args:     []Value{&runtime.IntegerValue{Value: -1}},
			expected: &runtime.IntegerValue{Value: 4294967295},
		},
		{
			name:     "large value truncates",
			args:     []Value{&runtime.IntegerValue{Value: 0x1FFFFFFFF}},
			expected: &runtime.IntegerValue{Value: 0xFFFFFFFF},
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.FloatValue{Value: 42.0}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Unsigned32(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Unsigned32() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMaxInt(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name:     "no arguments - returns constant",
			args:     []Value{},
			expected: &runtime.IntegerValue{Value: math.MaxInt64},
		},
		{
			name: "two integers - first larger",
			args: []Value{
				&runtime.IntegerValue{Value: 100},
				&runtime.IntegerValue{Value: 50},
			},
			expected: &runtime.IntegerValue{Value: 100},
		},
		{
			name: "two integers - second larger",
			args: []Value{
				&runtime.IntegerValue{Value: 30},
				&runtime.IntegerValue{Value: 60},
			},
			expected: &runtime.IntegerValue{Value: 60},
		},
		{
			name: "one argument",
			args: []Value{
				&runtime.IntegerValue{Value: 42},
			},
			isError: true,
		},
		{
			name: "wrong type",
			args: []Value{
				&runtime.IntegerValue{Value: 42},
				&runtime.FloatValue{Value: 10.0},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaxInt(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("MaxInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMinInt(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name:     "no arguments - returns constant",
			args:     []Value{},
			expected: &runtime.IntegerValue{Value: math.MinInt64},
		},
		{
			name: "two integers - first smaller",
			args: []Value{
				&runtime.IntegerValue{Value: 50},
				&runtime.IntegerValue{Value: 100},
			},
			expected: &runtime.IntegerValue{Value: 50},
		},
		{
			name: "two integers - second smaller",
			args: []Value{
				&runtime.IntegerValue{Value: 60},
				&runtime.IntegerValue{Value: 30},
			},
			expected: &runtime.IntegerValue{Value: 30},
		},
		{
			name: "one argument",
			args: []Value{
				&runtime.IntegerValue{Value: 42},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MinInt(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("MinInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsNaN(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
	}{
		{
			name:     "NaN value",
			args:     []Value{&runtime.FloatValue{Value: math.NaN()}},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name:     "normal float",
			args:     []Value{&runtime.FloatValue{Value: 42.0}},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name:     "infinity",
			args:     []Value{&runtime.FloatValue{Value: math.Inf(1)}},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name:     "integer",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: &runtime.BooleanValue{Value: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNaN(ctx, tt.args)

			if !valuesEqual(result, tt.expected) {
				t.Errorf("IsNaN() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPi(t *testing.T) {
	ctx := newMockContext()

	result := Pi(ctx, []Value{})
	expected := &runtime.FloatValue{Value: math.Pi}

	if !valuesEqual(result, expected) {
		t.Errorf("Pi() = %v, want %v", result, expected)
	}

	// Test with arguments (should error)
	result = Pi(ctx, []Value{&runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("expected error with arguments, got %v", result)
	}
}

func TestSign(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
	}{
		{
			name:     "positive integer",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: &runtime.IntegerValue{Value: 1},
		},
		{
			name:     "negative integer",
			args:     []Value{&runtime.IntegerValue{Value: -42}},
			expected: &runtime.IntegerValue{Value: -1},
		},
		{
			name:     "zero",
			args:     []Value{&runtime.IntegerValue{Value: 0}},
			expected: &runtime.IntegerValue{Value: 0},
		},
		{
			name:     "positive float",
			args:     []Value{&runtime.FloatValue{Value: 3.14}},
			expected: &runtime.IntegerValue{Value: 1},
		},
		{
			name:     "negative float",
			args:     []Value{&runtime.FloatValue{Value: -3.14}},
			expected: &runtime.IntegerValue{Value: -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sign(ctx, tt.args)

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Sign() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestOdd(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
		isError  bool
	}{
		{
			name:     "odd number",
			args:     []Value{&runtime.IntegerValue{Value: 5}},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name:     "even number",
			args:     []Value{&runtime.IntegerValue{Value: 4}},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name:     "negative odd",
			args:     []Value{&runtime.IntegerValue{Value: -3}},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name:     "zero",
			args:     []Value{&runtime.IntegerValue{Value: 0}},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name:    "float not allowed",
			args:    []Value{&runtime.FloatValue{Value: 5.0}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Odd(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Odd() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestInfinity(t *testing.T) {
	ctx := newMockContext()

	result := Infinity(ctx, []Value{})
	expected := &runtime.FloatValue{Value: math.Inf(1)}

	if !valuesEqual(result, expected) {
		t.Errorf("Infinity() = %v, want %v", result, expected)
	}

	// Test with arguments (should error)
	result = Infinity(ctx, []Value{&runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("expected error with arguments, got %v", result)
	}
}

func TestNaN(t *testing.T) {
	ctx := newMockContext()

	result := NaN(ctx, []Value{})

	floatResult, ok := result.(*runtime.FloatValue)
	if !ok {
		t.Fatalf("expected FloatValue, got %T", result)
	}

	if !math.IsNaN(floatResult.Value) {
		t.Errorf("NaN() should return NaN, got %v", floatResult.Value)
	}

	// Test with arguments (should error)
	result = NaN(ctx, []Value{&runtime.IntegerValue{Value: 1}})
	if result.Type() != "ERROR" {
		t.Errorf("expected error with arguments, got %v", result)
	}
}

func TestIsFinite(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
	}{
		{
			name:     "normal float",
			args:     []Value{&runtime.FloatValue{Value: 42.0}},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name:     "integer",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name:     "positive infinity",
			args:     []Value{&runtime.FloatValue{Value: math.Inf(1)}},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name:     "negative infinity",
			args:     []Value{&runtime.FloatValue{Value: math.Inf(-1)}},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name:     "NaN",
			args:     []Value{&runtime.FloatValue{Value: math.NaN()}},
			expected: &runtime.BooleanValue{Value: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFinite(ctx, tt.args)

			if !valuesEqual(result, tt.expected) {
				t.Errorf("IsFinite() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsInfinite(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		expected Value
		name     string
		args     []Value
	}{
		{
			name:     "positive infinity",
			args:     []Value{&runtime.FloatValue{Value: math.Inf(1)}},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name:     "negative infinity",
			args:     []Value{&runtime.FloatValue{Value: math.Inf(-1)}},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name:     "normal float",
			args:     []Value{&runtime.FloatValue{Value: 42.0}},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name:     "NaN",
			args:     []Value{&runtime.FloatValue{Value: math.NaN()}},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name:     "integer",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: &runtime.BooleanValue{Value: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInfinite(ctx, tt.args)

			if !valuesEqual(result, tt.expected) {
				t.Errorf("IsInfinite() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Helper Functions for Float Comparison
// =============================================================================

func floatValuesEqual(a, b Value) bool {
	av, aok := a.(*runtime.FloatValue)
	bv, bok := b.(*runtime.FloatValue)
	if !aok || !bok {
		return valuesEqual(a, b)
	}

	// Handle NaN comparison
	if math.IsNaN(av.Value) && math.IsNaN(bv.Value) {
		return true
	}
	// Handle infinity comparison
	if math.IsInf(av.Value, 1) && math.IsInf(bv.Value, 1) {
		return true
	}
	if math.IsInf(av.Value, -1) && math.IsInf(bv.Value, -1) {
		return true
	}
	// Regular float comparison with tolerance
	return math.Abs(av.Value-bv.Value) < 1e-9
}

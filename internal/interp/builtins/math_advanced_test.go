package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Advanced Math Functions Tests
// =============================================================================

func TestFactorial(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
		isError  bool
	}{
		{
			name:     "factorial of 0",
			args:     []Value{&runtime.IntegerValue{Value: 0}},
			expected: &runtime.IntegerValue{Value: 1},
		},
		{
			name:     "factorial of 1",
			args:     []Value{&runtime.IntegerValue{Value: 1}},
			expected: &runtime.IntegerValue{Value: 1},
		},
		{
			name:     "factorial of 5",
			args:     []Value{&runtime.IntegerValue{Value: 5}},
			expected: &runtime.IntegerValue{Value: 120},
		},
		{
			name:     "factorial of 10",
			args:     []Value{&runtime.IntegerValue{Value: 10}},
			expected: &runtime.IntegerValue{Value: 3628800},
		},
		{
			name:     "factorial of 20",
			args:     []Value{&runtime.IntegerValue{Value: 20}},
			expected: &runtime.IntegerValue{Value: 2432902008176640000},
		},
		{
			name:    "negative number",
			args:    []Value{&runtime.IntegerValue{Value: -5}},
			isError: true,
		},
		{
			name:    "overflow (21!)",
			args:    []Value{&runtime.IntegerValue{Value: 21}},
			isError: true,
		},
		{
			name:    "wrong type",
			args:    []Value{&runtime.FloatValue{Value: 5.0}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Factorial(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Factorial() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGcd(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
		isError  bool
	}{
		{
			name: "gcd(12, 18)",
			args: []Value{
				&runtime.IntegerValue{Value: 12},
				&runtime.IntegerValue{Value: 18},
			},
			expected: &runtime.IntegerValue{Value: 6},
		},
		{
			name: "gcd(100, 50)",
			args: []Value{
				&runtime.IntegerValue{Value: 100},
				&runtime.IntegerValue{Value: 50},
			},
			expected: &runtime.IntegerValue{Value: 50},
		},
		{
			name: "gcd with negative numbers",
			args: []Value{
				&runtime.IntegerValue{Value: -12},
				&runtime.IntegerValue{Value: 18},
			},
			expected: &runtime.IntegerValue{Value: 6},
		},
		{
			name: "coprime numbers",
			args: []Value{
				&runtime.IntegerValue{Value: 17},
				&runtime.IntegerValue{Value: 19},
			},
			expected: &runtime.IntegerValue{Value: 1},
		},
		{
			name: "one is zero",
			args: []Value{
				&runtime.IntegerValue{Value: 42},
				&runtime.IntegerValue{Value: 0},
			},
			expected: &runtime.IntegerValue{Value: 42},
		},
		{
			name: "wrong type",
			args: []Value{
				&runtime.FloatValue{Value: 12.0},
				&runtime.IntegerValue{Value: 18},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Gcd(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Gcd() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLcm(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
		isError  bool
	}{
		{
			name: "lcm(12, 18)",
			args: []Value{
				&runtime.IntegerValue{Value: 12},
				&runtime.IntegerValue{Value: 18},
			},
			expected: &runtime.IntegerValue{Value: 36},
		},
		{
			name: "lcm(4, 6)",
			args: []Value{
				&runtime.IntegerValue{Value: 4},
				&runtime.IntegerValue{Value: 6},
			},
			expected: &runtime.IntegerValue{Value: 12},
		},
		{
			name: "lcm with negative",
			args: []Value{
				&runtime.IntegerValue{Value: -12},
				&runtime.IntegerValue{Value: 18},
			},
			expected: &runtime.IntegerValue{Value: 36},
		},
		{
			name: "one is zero",
			args: []Value{
				&runtime.IntegerValue{Value: 0},
				&runtime.IntegerValue{Value: 5},
			},
			expected: &runtime.IntegerValue{Value: 0},
		},
		{
			name: "wrong type",
			args: []Value{
				&runtime.FloatValue{Value: 12.0},
				&runtime.IntegerValue{Value: 18},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Lcm(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("Lcm() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsPrime(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
	}{
		{
			name:     "2 is prime",
			args:     []Value{&runtime.IntegerValue{Value: 2}},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name:     "3 is prime",
			args:     []Value{&runtime.IntegerValue{Value: 3}},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name:     "17 is prime",
			args:     []Value{&runtime.IntegerValue{Value: 17}},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name:     "97 is prime",
			args:     []Value{&runtime.IntegerValue{Value: 97}},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name:     "1 is not prime",
			args:     []Value{&runtime.IntegerValue{Value: 1}},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name:     "4 is not prime",
			args:     []Value{&runtime.IntegerValue{Value: 4}},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name:     "15 is not prime",
			args:     []Value{&runtime.IntegerValue{Value: 15}},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name:     "100 is not prime",
			args:     []Value{&runtime.IntegerValue{Value: 100}},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name:     "negative is not prime",
			args:     []Value{&runtime.IntegerValue{Value: -5}},
			expected: &runtime.BooleanValue{Value: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPrime(ctx, tt.args)

			if !valuesEqual(result, tt.expected) {
				t.Errorf("IsPrime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLeastFactor(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
	}{
		{
			name:     "least factor of 1",
			args:     []Value{&runtime.IntegerValue{Value: 1}},
			expected: &runtime.IntegerValue{Value: 1},
		},
		{
			name:     "least factor of 2",
			args:     []Value{&runtime.IntegerValue{Value: 2}},
			expected: &runtime.IntegerValue{Value: 2},
		},
		{
			name:     "least factor of 15",
			args:     []Value{&runtime.IntegerValue{Value: 15}},
			expected: &runtime.IntegerValue{Value: 3},
		},
		{
			name:     "least factor of 100",
			args:     []Value{&runtime.IntegerValue{Value: 100}},
			expected: &runtime.IntegerValue{Value: 2},
		},
		{
			name:     "least factor of 77",
			args:     []Value{&runtime.IntegerValue{Value: 77}},
			expected: &runtime.IntegerValue{Value: 7},
		},
		{
			name:     "prime number returns itself",
			args:     []Value{&runtime.IntegerValue{Value: 17}},
			expected: &runtime.IntegerValue{Value: 17},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LeastFactor(ctx, tt.args)

			if !valuesEqual(result, tt.expected) {
				t.Errorf("LeastFactor() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPopCount(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
	}{
		{
			name:     "popcount of 0",
			args:     []Value{&runtime.IntegerValue{Value: 0}},
			expected: &runtime.IntegerValue{Value: 0},
		},
		{
			name:     "popcount of 1 (binary: 1)",
			args:     []Value{&runtime.IntegerValue{Value: 1}},
			expected: &runtime.IntegerValue{Value: 1},
		},
		{
			name:     "popcount of 7 (binary: 111)",
			args:     []Value{&runtime.IntegerValue{Value: 7}},
			expected: &runtime.IntegerValue{Value: 3},
		},
		{
			name:     "popcount of 15 (binary: 1111)",
			args:     []Value{&runtime.IntegerValue{Value: 15}},
			expected: &runtime.IntegerValue{Value: 4},
		},
		{
			name:     "popcount of 255 (binary: 11111111)",
			args:     []Value{&runtime.IntegerValue{Value: 255}},
			expected: &runtime.IntegerValue{Value: 8},
		},
		{
			name:     "popcount of -1 (all bits set)",
			args:     []Value{&runtime.IntegerValue{Value: -1}},
			expected: &runtime.IntegerValue{Value: 64},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PopCount(ctx, tt.args)

			if !valuesEqual(result, tt.expected) {
				t.Errorf("PopCount() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTestBit(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
		isError  bool
	}{
		{
			name: "bit 0 of 1",
			args: []Value{
				&runtime.IntegerValue{Value: 1},
				&runtime.IntegerValue{Value: 0},
			},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name: "bit 1 of 1",
			args: []Value{
				&runtime.IntegerValue{Value: 1},
				&runtime.IntegerValue{Value: 1},
			},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name: "bit 0 of 8 (binary: 1000)",
			args: []Value{
				&runtime.IntegerValue{Value: 8},
				&runtime.IntegerValue{Value: 0},
			},
			expected: &runtime.BooleanValue{Value: false},
		},
		{
			name: "bit 3 of 8 (binary: 1000)",
			args: []Value{
				&runtime.IntegerValue{Value: 8},
				&runtime.IntegerValue{Value: 3},
			},
			expected: &runtime.BooleanValue{Value: true},
		},
		{
			name: "bit out of range negative",
			args: []Value{
				&runtime.IntegerValue{Value: 8},
				&runtime.IntegerValue{Value: -1},
			},
			isError: true,
		},
		{
			name: "bit out of range too high",
			args: []Value{
				&runtime.IntegerValue{Value: 8},
				&runtime.IntegerValue{Value: 64},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TestBit(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("TestBit() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHaversine(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expected  float64 // Using float64 for easier comparison
		tolerance float64
		isError   bool
	}{
		{
			name: "same location",
			args: []Value{
				&runtime.FloatValue{Value: 40.7128},
				&runtime.FloatValue{Value: -74.0060},
				&runtime.FloatValue{Value: 40.7128},
				&runtime.FloatValue{Value: -74.0060},
			},
			expected:  0.0,
			tolerance: 0.01,
		},
		{
			name: "New York to London",
			args: []Value{
				&runtime.FloatValue{Value: 40.7128},  // NY lat
				&runtime.FloatValue{Value: -74.0060}, // NY lon
				&runtime.FloatValue{Value: 51.5074},  // London lat
				&runtime.FloatValue{Value: -0.1278},  // London lon
			},
			expected:  5570.0, // approximately 5570 km
			tolerance: 10.0,
		},
		{
			name: "works with integers",
			args: []Value{
				&runtime.IntegerValue{Value: 0},
				&runtime.IntegerValue{Value: 0},
				&runtime.IntegerValue{Value: 0},
				&runtime.IntegerValue{Value: 0},
			},
			expected:  0.0,
			tolerance: 0.01,
		},
		{
			name: "wrong argument count",
			args: []Value{
				&runtime.FloatValue{Value: 40.7128},
				&runtime.FloatValue{Value: -74.0060},
			},
			isError: true,
		},
		{
			name: "wrong type",
			args: []Value{
				&runtime.StringValue{Value: "not a number"},
				&runtime.FloatValue{Value: -74.0060},
				&runtime.FloatValue{Value: 51.5074},
				&runtime.FloatValue{Value: -0.1278},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Haversine(ctx, tt.args)

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

			diff := floatResult.Value - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tolerance {
				t.Errorf("Haversine() = %v, want %v (±%v)", floatResult.Value, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestCompareNum(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected Value
		isError  bool
	}{
		{
			name: "equal integers",
			args: []Value{
				&runtime.IntegerValue{Value: 5},
				&runtime.IntegerValue{Value: 5},
			},
			expected: &runtime.IntegerValue{Value: 0},
		},
		{
			name: "first integer less than second",
			args: []Value{
				&runtime.IntegerValue{Value: 3},
				&runtime.IntegerValue{Value: 7},
			},
			expected: &runtime.IntegerValue{Value: -1},
		},
		{
			name: "first integer greater than second",
			args: []Value{
				&runtime.IntegerValue{Value: 10},
				&runtime.IntegerValue{Value: 5},
			},
			expected: &runtime.IntegerValue{Value: 1},
		},
		{
			name: "equal floats",
			args: []Value{
				&runtime.FloatValue{Value: 3.14},
				&runtime.FloatValue{Value: 3.14},
			},
			expected: &runtime.IntegerValue{Value: 0},
		},
		{
			name: "first float less than second",
			args: []Value{
				&runtime.FloatValue{Value: 2.5},
				&runtime.FloatValue{Value: 3.5},
			},
			expected: &runtime.IntegerValue{Value: -1},
		},
		{
			name: "mixed int and float",
			args: []Value{
				&runtime.IntegerValue{Value: 5},
				&runtime.FloatValue{Value: 5.0},
			},
			expected: &runtime.IntegerValue{Value: 0},
		},
		{
			name: "wrong argument count",
			args: []Value{
				&runtime.IntegerValue{Value: 5},
			},
			isError: true,
		},
		{
			name: "wrong type",
			args: []Value{
				&runtime.StringValue{Value: "5"},
				&runtime.IntegerValue{Value: 5},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareNum(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("CompareNum() = %v, want %v", result, tt.expected)
			}
		})
	}
}

package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

func TestIntToStr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "basic integer",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: "42",
		},
		{
			name:     "negative integer",
			args:     []Value{&runtime.IntegerValue{Value: -123}},
			expected: "-123",
		},
		{
			name:     "zero",
			args:     []Value{&runtime.IntegerValue{Value: 0}},
			expected: "0",
		},
		{
			name:     "base 16",
			args:     []Value{&runtime.IntegerValue{Value: 255}, &runtime.IntegerValue{Value: 16}},
			expected: "ff",
		},
		{
			name:     "base 2",
			args:     []Value{&runtime.IntegerValue{Value: 5}, &runtime.IntegerValue{Value: 2}},
			expected: "101",
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
		{
			name:    "invalid base",
			args:    []Value{&runtime.IntegerValue{Value: 10}, &runtime.IntegerValue{Value: 1}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntToStr(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, strVal.Value)
			}
		})
	}
}

func TestIntToBin(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "basic conversion",
			args:     []Value{&runtime.IntegerValue{Value: 5}, &runtime.IntegerValue{Value: 4}},
			expected: "0101",
		},
		{
			name:     "zero",
			args:     []Value{&runtime.IntegerValue{Value: 0}, &runtime.IntegerValue{Value: 1}},
			expected: "0",
		},
		{
			name:    "wrong argument count",
			args:    []Value{&runtime.IntegerValue{Value: 5}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntToBin(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, strVal.Value)
			}
		})
	}
}

func TestStrToInt(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "basic integer",
			args:     []Value{&runtime.StringValue{Value: "42"}},
			expected: 42,
		},
		{
			name:     "negative integer",
			args:     []Value{&runtime.StringValue{Value: "-123"}},
			expected: -123,
		},
		{
			name:     "with whitespace",
			args:     []Value{&runtime.StringValue{Value: "  42  "}},
			expected: 42,
		},
		{
			name:     "hex with 0x prefix",
			args:     []Value{&runtime.StringValue{Value: "0xFF"}, &runtime.IntegerValue{Value: 16}},
			expected: 255,
		},
		{
			name:    "invalid string",
			args:    []Value{&runtime.StringValue{Value: "abc"}},
			isError: true,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToInt(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, intVal.Value)
			}
		})
	}
}

func TestStrToFloat(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected float64
		isError  bool
	}{
		{
			name:     "basic float",
			args:     []Value{&runtime.StringValue{Value: "3.14"}},
			expected: 3.14,
		},
		{
			name:     "negative float",
			args:     []Value{&runtime.StringValue{Value: "-123.45"}},
			expected: -123.45,
		},
		{
			name:     "with whitespace",
			args:     []Value{&runtime.StringValue{Value: "  2.5  "}},
			expected: 2.5,
		},
		{
			name:    "invalid string",
			args:    []Value{&runtime.StringValue{Value: "abc"}},
			isError: true,
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToFloat(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}

			if floatVal.Value != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, floatVal.Value)
			}
		})
	}
}

func TestFloatToStr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "basic float",
			args:     []Value{&runtime.FloatValue{Value: 3.14}},
			expected: "3.14",
		},
		{
			name:     "integer as float",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: "42",
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FloatToStr(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, strVal.Value)
			}
		})
	}
}

func TestBoolToStr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name:     "true",
			args:     []Value{&runtime.BooleanValue{Value: true}},
			expected: "True",
		},
		{
			name:     "false",
			args:     []Value{&runtime.BooleanValue{Value: false}},
			expected: "False",
		},
		{
			name:    "wrong argument count",
			args:    []Value{},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BoolToStr(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, strVal.Value)
			}
		})
	}
}

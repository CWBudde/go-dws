package runtime

import (
	"testing"
)

// ============================================================================
// ToInteger Tests
// ============================================================================

func TestToInteger(t *testing.T) {
	tests := []struct {
		input   Value
		name    string
		want    int64
		wantErr bool
	}{
		{name: "IntegerValue direct", input: NewInteger(42), want: 42, wantErr: false},
		{name: "FloatValue truncate", input: NewFloat(3.14), want: 3, wantErr: false},
		{name: "FloatValue truncate negative", input: NewFloat(-3.14), want: -3, wantErr: false},
		{name: "StringValue valid", input: NewString("42"), want: 42, wantErr: false},
		{name: "StringValue negative", input: NewString("-42"), want: -42, wantErr: false},
		{name: "StringValue invalid", input: NewString("not a number"), want: 0, wantErr: true},
		{name: "BooleanValue true", input: NewBoolean(true), want: 1, wantErr: false},
		{name: "BooleanValue false", input: NewBoolean(false), want: 0, wantErr: false},
		{name: "nil value", input: nil, want: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInteger(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToInteger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ToInteger() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// ToFloat Tests
// ============================================================================

func TestToFloat(t *testing.T) {
	tests := []struct {
		input   Value
		name    string
		want    float64
		wantErr bool
	}{
		{name: "FloatValue direct", input: NewFloat(3.14), want: 3.14, wantErr: false},
		{name: "IntegerValue convert", input: NewInteger(42), want: 42.0, wantErr: false},
		{name: "StringValue valid", input: NewString("3.14"), want: 3.14, wantErr: false},
		{name: "StringValue invalid", input: NewString("not a number"), want: 0.0, wantErr: true},
		{name: "BooleanValue true", input: NewBoolean(true), want: 1.0, wantErr: false},
		{name: "BooleanValue false", input: NewBoolean(false), want: 0.0, wantErr: false},
		{name: "nil value", input: nil, want: 0.0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToFloat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ToFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// ToString Tests
// ============================================================================

func TestToString(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  string
	}{
		{"StringValue", NewString("hello"), "hello"},
		{"IntegerValue", NewInteger(42), "42"},
		{"FloatValue", NewFloat(3.14), "3.14"},
		{"BooleanValue true", NewBoolean(true), "True"},
		{"BooleanValue false", NewBoolean(false), "False"},
		{"nil value", nil, "nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToString(tt.input)
			if got != tt.want {
				t.Errorf("ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// ToBoolean Tests
// ============================================================================

func TestToBoolean(t *testing.T) {
	tests := []struct {
		input   Value
		name    string
		want    bool
		wantErr bool
	}{
		{name: "BooleanValue true", input: NewBoolean(true), want: true, wantErr: false},
		{name: "BooleanValue false", input: NewBoolean(false), want: false, wantErr: false},
		{name: "IntegerValue non-zero", input: NewInteger(42), want: true, wantErr: false},
		{name: "IntegerValue zero", input: NewInteger(0), want: false, wantErr: false},
		{name: "FloatValue non-zero", input: NewFloat(3.14), want: true, wantErr: false},
		{name: "FloatValue zero", input: NewFloat(0.0), want: false, wantErr: false},
		{name: "StringValue true", input: NewString("true"), want: true, wantErr: false},
		{name: "StringValue false", input: NewString("false"), want: false, wantErr: false},
		{name: "StringValue yes", input: NewString("yes"), want: true, wantErr: false},
		{name: "StringValue no", input: NewString("no"), want: false, wantErr: false},
		{name: "StringValue 1", input: NewString("1"), want: true, wantErr: false},
		{name: "StringValue 0", input: NewString("0"), want: false, wantErr: false},
		{name: "StringValue empty", input: NewString(""), want: false, wantErr: false},
		{name: "StringValue invalid", input: NewString("maybe"), want: false, wantErr: true},
		{name: "nil value", input: nil, want: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToBoolean(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToBoolean() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ToBoolean() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// Arithmetic Operation Tests
// ============================================================================

func TestAddNumeric(t *testing.T) {
	tests := []struct {
		left    Value
		right   Value
		want    Value
		name    string
		wantErr bool
	}{
		{name: "int + int", left: NewInteger(2), right: NewInteger(3), want: NewInteger(5), wantErr: false},
		{name: "float + float", left: NewFloat(2.5), right: NewFloat(3.5), want: NewFloat(6.0), wantErr: false},
		{name: "int + float", left: NewInteger(2), right: NewFloat(3.5), want: NewFloat(5.5), wantErr: false},
		{name: "negative numbers", left: NewInteger(-2), right: NewInteger(3), want: NewInteger(1), wantErr: false},
		{name: "non-numeric left", left: NewString("hello"), right: NewInteger(3), want: nil, wantErr: true},
		{name: "non-numeric right", left: NewInteger(2), right: NewString("world"), want: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddNumeric(tt.left, tt.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !ValuesEqual(got, tt.want) {
					t.Errorf("AddNumeric() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestSubtractNumeric(t *testing.T) {
	tests := []struct {
		left    Value
		right   Value
		want    Value
		name    string
		wantErr bool
	}{
		{name: "int - int", left: NewInteger(5), right: NewInteger(3), want: NewInteger(2), wantErr: false},
		{name: "float - float", left: NewFloat(5.5), right: NewFloat(2.5), want: NewFloat(3.0), wantErr: false},
		{name: "negative result", left: NewInteger(3), right: NewInteger(5), want: NewInteger(-2), wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SubtractNumeric(tt.left, tt.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("SubtractNumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !ValuesEqual(got, tt.want) {
					t.Errorf("SubtractNumeric() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMultiplyNumeric(t *testing.T) {
	tests := []struct {
		left    Value
		right   Value
		want    Value
		name    string
		wantErr bool
	}{
		{name: "int * int", left: NewInteger(2), right: NewInteger(3), want: NewInteger(6), wantErr: false},
		{name: "float * float", left: NewFloat(2.5), right: NewFloat(4.0), want: NewFloat(10.0), wantErr: false},
		{name: "zero multiplication", left: NewInteger(5), right: NewInteger(0), want: NewInteger(0), wantErr: false},
		{name: "negative multiplication", left: NewInteger(-2), right: NewInteger(3), want: NewInteger(-6), wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MultiplyNumeric(tt.left, tt.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultiplyNumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !ValuesEqual(got, tt.want) {
					t.Errorf("MultiplyNumeric() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDivideNumeric(t *testing.T) {
	tests := []struct {
		left    Value
		right   Value
		want    Value
		name    string
		wantErr bool
	}{
		{name: "int / int", left: NewInteger(6), right: NewInteger(2), want: NewFloat(3.0), wantErr: false},
		{name: "float / float", left: NewFloat(10.0), right: NewFloat(2.5), want: NewFloat(4.0), wantErr: false},
		{name: "division by zero", left: NewInteger(5), right: NewInteger(0), want: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DivideNumeric(tt.left, tt.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("DivideNumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !ValuesEqual(got, tt.want) {
					t.Errorf("DivideNumeric() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestIntDivideNumeric(t *testing.T) {
	tests := []struct {
		left    Value
		right   Value
		want    Value
		name    string
		wantErr bool
	}{
		{name: "6 div 2", left: NewInteger(6), right: NewInteger(2), want: NewInteger(3), wantErr: false},
		{name: "7 div 2", left: NewInteger(7), right: NewInteger(2), want: NewInteger(3), wantErr: false},
		{name: "division by zero", left: NewInteger(5), right: NewInteger(0), want: nil, wantErr: true},
		{name: "negative division", left: NewInteger(-7), right: NewInteger(2), want: NewInteger(-3), wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IntDivideNumeric(tt.left, tt.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("IntDivideNumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !ValuesEqual(got, tt.want) {
					t.Errorf("IntDivideNumeric() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestModNumeric(t *testing.T) {
	tests := []struct {
		left    Value
		right   Value
		want    Value
		name    string
		wantErr bool
	}{
		{name: "7 mod 3", left: NewInteger(7), right: NewInteger(3), want: NewInteger(1), wantErr: false},
		{name: "10 mod 5", left: NewInteger(10), right: NewInteger(5), want: NewInteger(0), wantErr: false},
		{name: "modulo by zero", left: NewInteger(5), right: NewInteger(0), want: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ModNumeric(tt.left, tt.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModNumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !ValuesEqual(got, tt.want) {
					t.Errorf("ModNumeric() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkToInteger(b *testing.B) {
	val := NewInteger(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ToInteger(val)
	}
}

func BenchmarkToFloat(b *testing.B) {
	val := NewInteger(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ToFloat(val)
	}
}

func BenchmarkAddNumericIntegers(b *testing.B) {
	left := NewInteger(42)
	right := NewInteger(58)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AddNumeric(left, right)
	}
}

func BenchmarkAddNumericFloats(b *testing.B) {
	left := NewFloat(42.5)
	right := NewFloat(58.3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AddNumeric(left, right)
	}
}

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
		{"IntegerValue direct", NewInteger(42), 42, false},
		{"FloatValue truncate", NewFloat(3.14), 3, false},
		{"FloatValue truncate negative", NewFloat(-3.14), -3, false},
		{"StringValue valid", NewString("42"), 42, false},
		{"StringValue negative", NewString("-42"), -42, false},
		{"StringValue invalid", NewString("not a number"), 0, true},
		{"BooleanValue true", NewBoolean(true), 1, false},
		{"BooleanValue false", NewBoolean(false), 0, false},
		{"nil value", nil, 0, true},
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
		{"FloatValue direct", NewFloat(3.14), 3.14, false},
		{"IntegerValue convert", NewInteger(42), 42.0, false},
		{"StringValue valid", NewString("3.14"), 3.14, false},
		{"StringValue invalid", NewString("not a number"), 0.0, true},
		{"BooleanValue true", NewBoolean(true), 1.0, false},
		{"BooleanValue false", NewBoolean(false), 0.0, false},
		{"nil value", nil, 0.0, true},
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
		{"BooleanValue true", NewBoolean(true), true, false},
		{"BooleanValue false", NewBoolean(false), false, false},
		{"IntegerValue non-zero", NewInteger(42), true, false},
		{"IntegerValue zero", NewInteger(0), false, false},
		{"FloatValue non-zero", NewFloat(3.14), true, false},
		{"FloatValue zero", NewFloat(0.0), false, false},
		{"StringValue true", NewString("true"), true, false},
		{"StringValue false", NewString("false"), false, false},
		{"StringValue yes", NewString("yes"), true, false},
		{"StringValue no", NewString("no"), false, false},
		{"StringValue 1", NewString("1"), true, false},
		{"StringValue 0", NewString("0"), false, false},
		{"StringValue empty", NewString(""), false, false},
		{"StringValue invalid", NewString("maybe"), false, true},
		{"nil value", nil, false, true},
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
		{"int + int", NewInteger(2), NewInteger(3), NewInteger(5), false},
		{"float + float", NewFloat(2.5), NewFloat(3.5), NewFloat(6.0), false},
		{"int + float", NewInteger(2), NewFloat(3.5), NewFloat(5.5), false},
		{"negative numbers", NewInteger(-2), NewInteger(3), NewInteger(1), false},
		{"non-numeric left", NewString("hello"), NewInteger(3), nil, true},
		{"non-numeric right", NewInteger(2), NewString("world"), nil, true},
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
		{"int - int", NewInteger(5), NewInteger(3), NewInteger(2), false},
		{"float - float", NewFloat(5.5), NewFloat(2.5), NewFloat(3.0), false},
		{"negative result", NewInteger(3), NewInteger(5), NewInteger(-2), false},
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
		{"int * int", NewInteger(2), NewInteger(3), NewInteger(6), false},
		{"float * float", NewFloat(2.5), NewFloat(4.0), NewFloat(10.0), false},
		{"zero multiplication", NewInteger(5), NewInteger(0), NewInteger(0), false},
		{"negative multiplication", NewInteger(-2), NewInteger(3), NewInteger(-6), false},
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
		{"int / int", NewInteger(6), NewInteger(2), NewFloat(3.0), false},
		{"float / float", NewFloat(10.0), NewFloat(2.5), NewFloat(4.0), false},
		{"division by zero", NewInteger(5), NewInteger(0), nil, true},
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
		{"6 div 2", NewInteger(6), NewInteger(2), NewInteger(3), false},
		{"7 div 2", NewInteger(7), NewInteger(2), NewInteger(3), false},
		{"division by zero", NewInteger(5), NewInteger(0), nil, true},
		{"negative division", NewInteger(-7), NewInteger(2), NewInteger(-3), false},
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
		{"7 mod 3", NewInteger(7), NewInteger(3), NewInteger(1), false},
		{"10 mod 5", NewInteger(10), NewInteger(5), NewInteger(0), false},
		{"modulo by zero", NewInteger(5), NewInteger(0), nil, true},
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

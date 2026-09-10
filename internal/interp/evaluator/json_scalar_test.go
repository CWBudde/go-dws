package evaluator

import (
	"math"
	"testing"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

func TestJSONScalarFloat(t *testing.T) {
	tests := []struct {
		name  string
		value *jsonvalue.Value
		want  float64
		ok    bool
	}{
		{"int64", jsonvalue.NewInt64(4611686018427387905), 4611686018427387905, true},
		{"number", jsonvalue.NewNumber(1.25), 1.25, true},
		{"true", jsonvalue.NewBoolean(true), 1, true},
		{"false", jsonvalue.NewBoolean(false), 0, true},
		{"numeric string", jsonvalue.NewString("1.25"), 1.25, true},
		{"empty string", jsonvalue.NewString(""), 0, true},
		{"non-numeric string", jsonvalue.NewString("abc"), 0, true},
		{"object is not a value", jsonvalue.NewObject(), 0, false},
		{"array is not a value", jsonvalue.NewArray(), 0, false},
		{"null is not a value", jsonvalue.NewNull(), 0, false},
		{"undefined is not a value", jsonvalue.NewUndefined(), 0, false},
		{"nil is not a value", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := jsonScalarFloat(tt.value)
			if ok != tt.ok {
				t.Fatalf("jsonScalarFloat() ok = %v, want %v", ok, tt.ok)
			}
			if ok && math.Abs(got-tt.want) > 1e-9*math.Max(1, math.Abs(tt.want)) {
				t.Errorf("jsonScalarFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONScalarInteger(t *testing.T) {
	tests := []struct {
		name  string
		value *jsonvalue.Value
		want  int64
		ok    bool
	}{
		{"int64 keeps full precision", jsonvalue.NewInt64(4611686018427387905), 4611686018427387905, true},
		{"number truncates", jsonvalue.NewNumber(1.25), 1, true},
		{"numeric string truncates", jsonvalue.NewString("1.25"), 1, true},
		{"empty string", jsonvalue.NewString(""), 0, true},
		{"boolean", jsonvalue.NewBoolean(true), 1, true},
		{"object is not a value", jsonvalue.NewObject(), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := jsonScalarInteger(tt.value)
			if ok != tt.ok {
				t.Fatalf("jsonScalarInteger() ok = %v, want %v", ok, tt.ok)
			}
			if ok && got != tt.want {
				t.Errorf("jsonScalarInteger() = %d, want %d", got, tt.want)
			}
		})
	}
}

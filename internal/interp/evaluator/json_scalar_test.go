package evaluator

import (
	"math"
	"testing"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

func TestJSONScalarFloat(t *testing.T) {
	tests := []struct {
		value *jsonvalue.Value
		name  string
		want  float64
		ok    bool
	}{
		{name: "int64", value: jsonvalue.NewInt64(4611686018427387905), want: 4611686018427387905, ok: true},
		{name: "number", value: jsonvalue.NewNumber(1.25), want: 1.25, ok: true},
		{name: "true", value: jsonvalue.NewBoolean(true), want: 1, ok: true},
		{name: "false", value: jsonvalue.NewBoolean(false), want: 0, ok: true},
		{name: "numeric string", value: jsonvalue.NewString("1.25"), want: 1.25, ok: true},
		{name: "empty string", value: jsonvalue.NewString(""), want: 0, ok: true},
		{name: "non-numeric string", value: jsonvalue.NewString("abc"), want: 0, ok: true},
		{name: "object is not a value", value: jsonvalue.NewObject(), want: 0, ok: false},
		{name: "array is not a value", value: jsonvalue.NewArray(), want: 0, ok: false},
		{name: "null is not a value", value: jsonvalue.NewNull(), want: 0, ok: false},
		{name: "undefined is not a value", value: jsonvalue.NewUndefined(), want: 0, ok: false},
		{name: "nil is not a value", value: nil, want: 0, ok: false},
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
		value *jsonvalue.Value
		name  string
		want  int64
		ok    bool
	}{
		{name: "int64 keeps full precision", value: jsonvalue.NewInt64(4611686018427387905), want: 4611686018427387905, ok: true},
		{name: "number truncates", value: jsonvalue.NewNumber(1.25), want: 1, ok: true},
		{name: "numeric string truncates", value: jsonvalue.NewString("1.25"), want: 1, ok: true},
		{name: "empty string", value: jsonvalue.NewString(""), want: 0, ok: true},
		{name: "boolean", value: jsonvalue.NewBoolean(true), want: 1, ok: true},
		{name: "object is not a value", value: jsonvalue.NewObject(), want: 0, ok: false},
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

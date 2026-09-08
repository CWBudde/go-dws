package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

// TestParameterlessBuiltinAsBareIdentifier verifies that a builtin taking no
// arguments carries its result type when used as a bare identifier, so it can
// take part in ordinary expressions instead of typing as VOID.
func TestParameterlessBuiltinAsBareIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Random in arithmetic", `var x: Float := Random * 0;`},
		{"Pi in arithmetic", `var x: Float := Pi * 2;`},
		{"Now compared", `var b: Boolean := Now > 0;`},
		{"RandSeed is Integer", `var i: Integer := RandSeed;`},
		{"inferred from Random", `var x := Random; var y := x * 0;`},
		{"method call on result", `var s: String := (Random * 0).ToString;`},
		{"GetStackTrace is String", `var s: String := GetStackTrace;`},
		{"UnixTime is Integer", `var i: Integer := UnixTime;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

// TestParameterlessBuiltinType checks which registry signatures qualify as
// implicit parameterless calls. Optional, variadic, and procedure signatures
// must not, since a bare name there does not unambiguously mean a call.
func TestParameterlessBuiltinType(t *testing.T) {
	analyzer := NewAnalyzer()

	tests := []struct {
		expected types.Type
		name     string
		builtin  string
		ok       bool
	}{
		{types.FLOAT, "parameterless function", "Random", true},
		{types.INTEGER, "parameterless Integer function", "RandSeed", true},
		{types.FLOAT, "case insensitive", "rAnDoM", true},
		{types.INTEGER, "not in the isBuiltinFunction list", "UnixTime", true},
		{nil, "parameterless procedure", "Randomize", false},
		{nil, "required parameters", "Sqrt", false},
		{nil, "optional parameters", "Trim", false},
		{nil, "variadic", "PrintLn", false},
		{nil, "unknown name", "NoSuchBuiltin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultType, ok := analyzer.parameterlessBuiltinType(tt.builtin)
			if ok != tt.ok {
				t.Fatalf("parameterlessBuiltinType(%q) ok = %v, want %v", tt.builtin, ok, tt.ok)
			}
			if resultType != tt.expected {
				t.Errorf("parameterlessBuiltinType(%q) = %v, want %v", tt.builtin, resultType, tt.expected)
			}
		})
	}
}

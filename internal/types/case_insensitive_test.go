package types

import (
	"testing"
)

// TestTypeFromString_CaseInsensitive verifies that TypeFromString correctly handles
// type names in any case (lowercase, UPPERCASE, MiXeD case).
// DWScript is case-insensitive, so all variations should work.
func TestTypeFromString_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Type
		wantErr  bool
	}{
		// PascalCase (original/traditional format)
		{"Integer PascalCase", "Integer", INTEGER, false},
		{"Float PascalCase", "Float", FLOAT, false},
		{"String PascalCase", "String", STRING, false},
		{"Boolean PascalCase", "Boolean", BOOLEAN, false},
		{"TDateTime PascalCase", "TDateTime", DATETIME, false},
		{"Void PascalCase", "Void", VOID, false},
		{"Variant PascalCase", "Variant", VARIANT, false},

		// lowercase (common in some Pascal code)
		{"Integer lowercase", "integer", INTEGER, false},
		{"Float lowercase", "float", FLOAT, false},
		{"String lowercase", "string", STRING, false},
		{"Boolean lowercase", "boolean", BOOLEAN, false},
		{"TDateTime lowercase", "tdatetime", DATETIME, false},
		{"Void lowercase", "void", VOID, false},
		{"Variant lowercase", "variant", VARIANT, false},

		// UPPERCASE (sometimes used for emphasis)
		{"Integer UPPERCASE", "INTEGER", INTEGER, false},
		{"Float UPPERCASE", "FLOAT", FLOAT, false},
		{"String UPPERCASE", "STRING", STRING, false},
		{"Boolean UPPERCASE", "BOOLEAN", BOOLEAN, false},
		{"TDateTime UPPERCASE", "TDATETIME", DATETIME, false},
		{"Void UPPERCASE", "VOID", VOID, false},
		{"Variant UPPERCASE", "VARIANT", VARIANT, false},

		// MiXeD CaSe (edge cases)
		{"Integer mixed 1", "InTeGeR", INTEGER, false},
		{"Integer mixed 2", "iNtEgEr", INTEGER, false},
		{"Float mixed 1", "FlOaT", FLOAT, false},
		{"Float mixed 2", "fLoAt", FLOAT, false},
		{"String mixed 1", "StRiNg", STRING, false},
		{"String mixed 2", "sTrInG", STRING, false},
		{"Boolean mixed 1", "BoOlEaN", BOOLEAN, false},
		{"Boolean mixed 2", "bOoLeAn", BOOLEAN, false},
		{"TDateTime mixed 1", "TdAtEtImE", DATETIME, false},
		{"TDateTime mixed 2", "tDaTeTiMe", DATETIME, false},
		{"Void mixed", "VoId", VOID, false},
		{"Variant mixed", "VaRiAnT", VARIANT, false},

		// Unknown type (should still error regardless of case)
		{"Unknown lowercase", "unknown", nil, true},
		{"Unknown UPPERCASE", "UNKNOWN", nil, true},
		{"Unknown mixed", "UnKnOwN", nil, true},
		{"Unknown PascalCase", "UnknownType", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TypeFromString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("TypeFromString(%q) expected error but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("TypeFromString(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("TypeFromString(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// TestTypeFromString_AllBuiltinTypes ensures all built-in types work in lowercase,
// which is the most common case-insensitivity issue.
func TestTypeFromString_AllBuiltinTypes(t *testing.T) {
	builtins := []struct {
		lowercase string
		expected  Type
	}{
		{"integer", INTEGER},
		{"float", FLOAT},
		{"string", STRING},
		{"boolean", BOOLEAN},
		{"tdatetime", DATETIME},
		{"void", VOID},
		{"variant", VARIANT},
	}

	for _, b := range builtins {
		t.Run(b.lowercase, func(t *testing.T) {
			result, err := TypeFromString(b.lowercase)
			if err != nil {
				t.Errorf("TypeFromString(%q) failed: %v", b.lowercase, err)
			}
			if result != b.expected {
				t.Errorf("TypeFromString(%q) = %v, want %v", b.lowercase, result, b.expected)
			}
		})
	}
}

// TestTypeFromString_PreserveErrorMessage verifies that error messages still
// show the original (user-provided) case, not the normalized lowercase version.
func TestTypeFromString_PreserveErrorMessage(t *testing.T) {
	tests := []struct {
		input       string
		wantContain string
	}{
		{"UnknownType", "UnknownType"},
		{"unknowntype", "unknowntype"},
		{"UNKNOWNTYPE", "UNKNOWNTYPE"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := TypeFromString(tt.input)
			if err == nil {
				t.Fatalf("TypeFromString(%q) expected error but got none", tt.input)
			}
			if err.Error() != "unknown type: "+tt.wantContain {
				t.Errorf("TypeFromString(%q) error = %q, want to contain %q",
					tt.input, err.Error(), tt.wantContain)
			}
		})
	}
}

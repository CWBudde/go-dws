package types

import (
	"testing"
)

// TestTypeFromString_CaseInsensitive verifies that TypeFromString correctly handles
// type names in any case (lowercase, UPPERCASE, MiXeD case).
// DWScript is case-insensitive, so all variations should work.
func TestTypeFromString_CaseInsensitive(t *testing.T) {
	tests := []struct {
		expected Type
		name     string
		input    string
		wantErr  bool
	}{
		// PascalCase (original/traditional format)
		{expected: INTEGER, name: "Integer PascalCase", input: "Integer", wantErr: false},
		{expected: FLOAT, name: "Float PascalCase", input: "Float", wantErr: false},
		{expected: STRING, name: "String PascalCase", input: "String", wantErr: false},
		{expected: BOOLEAN, name: "Boolean PascalCase", input: "Boolean", wantErr: false},
		{expected: DATETIME, name: "TDateTime PascalCase", input: "TDateTime", wantErr: false},
		{expected: VOID, name: "Void PascalCase", input: "Void", wantErr: false},
		{expected: VARIANT, name: "Variant PascalCase", input: "Variant", wantErr: false},

		// lowercase (common in some Pascal code)
		{expected: INTEGER, name: "Integer lowercase", input: "integer", wantErr: false},
		{expected: FLOAT, name: "Float lowercase", input: "float", wantErr: false},
		{expected: STRING, name: "String lowercase", input: "string", wantErr: false},
		{expected: BOOLEAN, name: "Boolean lowercase", input: "boolean", wantErr: false},
		{expected: DATETIME, name: "TDateTime lowercase", input: "tdatetime", wantErr: false},
		{expected: VOID, name: "Void lowercase", input: "void", wantErr: false},
		{expected: VARIANT, name: "Variant lowercase", input: "variant", wantErr: false},

		// UPPERCASE (sometimes used for emphasis)
		{expected: INTEGER, name: "Integer UPPERCASE", input: "INTEGER", wantErr: false},
		{expected: FLOAT, name: "Float UPPERCASE", input: "FLOAT", wantErr: false},
		{expected: STRING, name: "String UPPERCASE", input: "STRING", wantErr: false},
		{expected: BOOLEAN, name: "Boolean UPPERCASE", input: "BOOLEAN", wantErr: false},
		{expected: DATETIME, name: "TDateTime UPPERCASE", input: "TDATETIME", wantErr: false},
		{expected: VOID, name: "Void UPPERCASE", input: "VOID", wantErr: false},
		{expected: VARIANT, name: "Variant UPPERCASE", input: "VARIANT", wantErr: false},

		// MiXeD CaSe (edge cases)
		{expected: INTEGER, name: "Integer mixed 1", input: "InTeGeR", wantErr: false},
		{expected: INTEGER, name: "Integer mixed 2", input: "iNtEgEr", wantErr: false},
		{expected: FLOAT, name: "Float mixed 1", input: "FlOaT", wantErr: false},
		{expected: FLOAT, name: "Float mixed 2", input: "fLoAt", wantErr: false},
		{expected: STRING, name: "String mixed 1", input: "StRiNg", wantErr: false},
		{expected: STRING, name: "String mixed 2", input: "sTrInG", wantErr: false},
		{expected: BOOLEAN, name: "Boolean mixed 1", input: "BoOlEaN", wantErr: false},
		{expected: BOOLEAN, name: "Boolean mixed 2", input: "bOoLeAn", wantErr: false},
		{expected: DATETIME, name: "TDateTime mixed 1", input: "TdAtEtImE", wantErr: false},
		{expected: DATETIME, name: "TDateTime mixed 2", input: "tDaTeTiMe", wantErr: false},
		{expected: VOID, name: "Void mixed", input: "VoId", wantErr: false},
		{expected: VARIANT, name: "Variant mixed", input: "VaRiAnT", wantErr: false},

		// Unknown type (should still error regardless of case)
		{expected: nil, name: "Unknown lowercase", input: "unknown", wantErr: true},
		{expected: nil, name: "Unknown UPPERCASE", input: "UNKNOWN", wantErr: true},
		{expected: nil, name: "Unknown mixed", input: "UnKnOwN", wantErr: true},
		{expected: nil, name: "Unknown PascalCase", input: "UnknownType", wantErr: true},
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
		expected  Type
		lowercase string
	}{
		{expected: INTEGER, lowercase: "integer"},
		{expected: FLOAT, lowercase: "float"},
		{expected: STRING, lowercase: "string"},
		{expected: BOOLEAN, lowercase: "boolean"},
		{expected: DATETIME, lowercase: "tdatetime"},
		{expected: VOID, lowercase: "void"},
		{expected: VARIANT, lowercase: "variant"},
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
		{input: "UnknownType", wantContain: "UnknownType"},
		{input: "unknowntype", wantContain: "unknowntype"},
		{input: "UNKNOWNTYPE", wantContain: "UNKNOWNTYPE"},
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

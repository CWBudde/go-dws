package semantic

import (
	"testing"
)

// ============================================================================
// String Helper Method Tests (Task 9.23.6.1)
// ============================================================================

func TestStringHelperMethods_Conversion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"ToUpper", "var s := 'hello'; var u := s.ToUpper;", true},
		{"ToLower", "var s := 'HELLO'; var l := s.ToLower;", true},
		{"ToInteger valid", "var s := '123'; var i := s.ToInteger;", true},
		{"ToFloat valid", "var s := '3.14'; var f := s.ToFloat;", true},
		{"ToString identity", "var s := 'hello'; var t := s.ToString;", true},

		// Test with parentheses (optional for zero-arg methods)
		{"ToUpper with parens", "var s := 'hello'; var u := s.ToUpper();", true},
		{"ToLower with parens", "var s := 'HELLO'; var l := s.ToLower();", true},

		// Test in expressions
		{"ToUpper in concat", "var s := 'hello' + ' ' + 'world'.ToUpper;", true},
		{"ToInteger in arithmetic", "var n := '5'.ToInteger + 10;", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				expectNoErrors(t, tt.input)
			} else {
				_, err := analyzeSource(t, tt.input)
				if err == nil {
					t.Errorf("expected error for: %s", tt.input)
				}
			}
		})
	}
}

func TestStringHelperMethods_SearchCheck(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"StartsWith", "var s := 'hello'; var b := s.StartsWith('he');", true},
		{"EndsWith", "var s := 'hello'; var b := s.EndsWith('lo');", true},
		{"Contains", "var s := 'hello'; var b := s.Contains('ll');", true},
		{"IndexOf", "var s := 'hello'; var i := s.IndexOf('ll');", true},

		// Test with variables
		{"StartsWith with var", "var s := 'hello'; var prefix := 'he'; var b := s.StartsWith(prefix);", true},

		// Test return types (Boolean for StartsWith/EndsWith/Contains, Integer for IndexOf)
		{"StartsWith returns Boolean", "var s := 'hello'; if s.StartsWith('he') then PrintLn('yes');", true},
		{"IndexOf returns Integer", "var s := 'hello'; var i: Integer := s.IndexOf('ll');", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				expectNoErrors(t, tt.input)
			} else {
				_, err := analyzeSource(t, tt.input)
				if err == nil {
					t.Errorf("expected error for: %s", tt.input)
				}
			}
		})
	}
}

func TestStringHelperMethods_Extraction(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"Copy with 2 params", "var s := 'hello'; var c := s.Copy(2, 3);", true},
		{"Copy with 1 param", "var s := 'hello'; var c := s.Copy(3);", true},
		{"Before", "var s := 'hello world'; var b := s.Before(' ');", true},
		{"After", "var s := 'hello world'; var a := s.After(' ');", true},

		// Test that Copy returns String
		{"Copy returns String", "var s := 'hello'; var c: String := s.Copy(2, 3);", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				expectNoErrors(t, tt.input)
			} else {
				_, err := analyzeSource(t, tt.input)
				if err == nil {
					t.Errorf("expected error for: %s", tt.input)
				}
			}
		})
	}
}

func TestStringHelperMethods_Split(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"Split basic", "var s := 'a,b,c'; var parts := s.Split(',');", true},
		{"Split with access", "var s := 'a,b,c'; var parts := s.Split(','); var first := parts[0];", true},

		// Split should return array of strings
		{"Split returns array", "var s := 'a,b,c'; var parts: array of String := s.Split(',');", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				expectNoErrors(t, tt.input)
			} else {
				_, err := analyzeSource(t, tt.input)
				if err == nil {
					t.Errorf("expected error for: %s", tt.input)
				}
			}
		})
	}
}

func TestStringHelperMethods_Chaining(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"Copy then ToUpper", "var s := 'hello'; var r := s.Copy(2, 3).ToUpper;", true},
		{"ToLower then Copy", "var s := 'HELLO'; var r := s.ToLower.Copy(1, 3);", true},
		{"Before then ToUpper", "var s := 'hello world'; var r := s.Before(' ').ToUpper;", true},

		// Complex chaining
		{"Triple chain", "var s := 'hello world'; var r := s.Before(' ').Copy(2, 3).ToUpper;", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				expectNoErrors(t, tt.input)
			} else {
				_, err := analyzeSource(t, tt.input)
				if err == nil {
					t.Errorf("expected error for: %s", tt.input)
				}
			}
		})
	}
}

func TestStringHelperMethods_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"Empty string literal", "var s := ''; var u := s.ToUpper;", true},
		{"String variable", "var s: String; s := 'test'; var u := s.ToUpper;", true},

		// Helper on string literals directly
		{"Literal ToUpper", "var u := 'hello'.ToUpper;", true},
		{"Literal Copy", "var c := 'hello'.Copy(2, 3);", true},
		{"Literal StartsWith", "var b := 'hello'.StartsWith('he');", true},

		// In function calls
		{"Helper in PrintLn", "PrintLn('hello'.ToUpper);", true},
		{"Helper in concat", "var s := 'hello'.ToUpper + ' ' + 'world'.ToLower;", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				expectNoErrors(t, tt.input)
			} else {
				_, err := analyzeSource(t, tt.input)
				if err == nil {
					t.Errorf("expected error for: %s", tt.input)
				}
			}
		})
	}
}

func TestStringHelperMethods_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		// Note: Some of these error checks may not be enforced at semantic analysis time
		// (e.g., parameter counts might be checked at runtime). These tests verify
		// what errors ARE caught during semantic analysis.

		// Unknown methods would be caught
		{"Unknown method", "var s := 'hello'; var x := s.UnknownMethod();", "unknown"},

		// Type mismatches for non-string receivers (if enforced)
		{"ToUpper on Integer", "var i := 42; var s := i.ToUpper;", ""},
		{"StartsWith on Integer", "var i := 42; var b := i.StartsWith('4');", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := analyzeSource(t, tt.input)
			if tt.expectedError == "" {
				// Test expects error but we don't check the message
				if err == nil {
					t.Errorf("expected error for: %s", tt.input)
				}
			} else {
				// Test expects specific error message
				expectError(t, tt.input, tt.expectedError)
			}
		})
	}
}

func TestStringHelperMethods_TypeCompatibility(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		// Test that helper methods work with String type annotations
		{"String variable ToUpper", "var s: String := 'hello'; var u := s.ToUpper;", true},
		{"String param helper", "function Test(s: String): String; begin Result := s.ToUpper; end;", true},

		// Test in various contexts
		{"Helper in if condition", "var s := 'test'; if s.StartsWith('t') then PrintLn('yes');", true},
		{"Helper in while condition", "var s := 'test'; while s.Contains('e') do s := s.Copy(2);", true},
		{"Helper in assignment", "var s: String; s := 'hello'.ToUpper;", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				expectNoErrors(t, tt.input)
			} else {
				_, err := analyzeSource(t, tt.input)
				if err == nil {
					t.Errorf("expected error for: %s", tt.input)
				}
			}
		})
	}
}

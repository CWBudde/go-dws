package bytecode

import (
	"testing"
)

// TestSetStringLength tests the setStringLength function
func TestSetStringLength(t *testing.T) {
	vm := NewVM()

	tests := []struct {
		name      string
		input     string
		newLength int
		expected  string
	}{
		{
			name:      "same length",
			input:     "hello",
			newLength: 5,
			expected:  "hello",
		},
		{
			name:      "truncate string",
			input:     "hello world",
			newLength: 5,
			expected:  "hello",
		},
		{
			name:      "truncate to empty",
			input:     "hello",
			newLength: 0,
			expected:  "",
		},
		{
			name:      "extend string with null bytes",
			input:     "hi",
			newLength: 5,
			expected:  "hi\x00\x00\x00",
		},
		{
			name:      "negative length becomes zero",
			input:     "test",
			newLength: -1,
			expected:  "",
		},
		{
			name:      "extend empty string",
			input:     "",
			newLength: 3,
			expected:  "\x00\x00\x00",
		},
		{
			name:      "unicode string truncate",
			input:     "hello世界",
			newLength: 5,
			expected:  "hello",
		},
		{
			name:      "unicode string extend",
			input:     "世界",
			newLength: 4,
			expected:  "世界\x00\x00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.setStringLength(tt.input, tt.newLength)
			if result != tt.expected {
				t.Errorf("setStringLength(%q, %d) = %q, want %q", tt.input, tt.newLength, result, tt.expected)
			}
		})
	}
}

// TestEqualsCaseInsensitive tests the equalsCaseInsensitive function
func TestEqualsCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "equal same case",
			a:        "hello",
			b:        "hello",
			expected: true,
		},
		{
			name:     "equal different case",
			a:        "Hello",
			b:        "hello",
			expected: true,
		},
		{
			name:     "equal all uppercase",
			a:        "HELLO",
			b:        "hello",
			expected: true,
		},
		{
			name:     "equal mixed case",
			a:        "HeLLo",
			b:        "hEllO",
			expected: true,
		},
		{
			name:     "not equal",
			a:        "hello",
			b:        "world",
			expected: false,
		},
		{
			name:     "different lengths",
			a:        "hello",
			b:        "helloworld",
			expected: false,
		},
		{
			name:     "empty strings",
			a:        "",
			b:        "",
			expected: true,
		},
		{
			name:     "one empty",
			a:        "hello",
			b:        "",
			expected: false,
		},
		{
			name:     "type names",
			a:        "Integer",
			b:        "integer",
			expected: true,
		},
		{
			name:     "type names uppercase",
			a:        "INTEGER",
			b:        "integer",
			expected: true,
		},
		{
			name:     "numbers in strings",
			a:        "Test123",
			b:        "test123",
			expected: true,
		},
		{
			name:     "special characters",
			a:        "Test_Name",
			b:        "test_name",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := equalsCaseInsensitive(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("equalsCaseInsensitive(%q, %q) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestFindHelperMethodCaseInsensitive tests the findHelperMethodCaseInsensitive function
func TestFindHelperMethodCaseInsensitive(t *testing.T) {
	methods := map[string]uint16{
		"ToUpper":   1,
		"ToLower":   2,
		"Substring": 3,
		"Length":    4,
	}

	tests := []struct {
		name          string
		methodName    string
		expectedSlot  uint16
		expectedFound bool
	}{
		{
			name:          "exact match",
			methodName:    "ToUpper",
			expectedSlot:  1,
			expectedFound: true,
		},
		{
			name:          "lowercase match",
			methodName:    "toupper",
			expectedSlot:  1,
			expectedFound: true,
		},
		{
			name:          "uppercase match",
			methodName:    "TOLOWER",
			expectedSlot:  2,
			expectedFound: true,
		},
		{
			name:          "mixed case match",
			methodName:    "sUbStRiNg",
			expectedSlot:  3,
			expectedFound: true,
		},
		{
			name:          "not found",
			methodName:    "NotAMethod",
			expectedSlot:  0,
			expectedFound: false,
		},
		{
			name:          "empty method name",
			methodName:    "",
			expectedSlot:  0,
			expectedFound: false,
		},
		{
			name:          "partial match (should not match)",
			methodName:    "To",
			expectedSlot:  0,
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slot, found := findHelperMethodCaseInsensitive(methods, tt.methodName)
			if found != tt.expectedFound {
				t.Errorf("findHelperMethodCaseInsensitive(%q) found = %v, want %v", tt.methodName, found, tt.expectedFound)
			}
			if found && slot != tt.expectedSlot {
				t.Errorf("findHelperMethodCaseInsensitive(%q) slot = %d, want %d", tt.methodName, slot, tt.expectedSlot)
			}
		})
	}
}

// TestFindHelperForValue tests the findHelperForValue function
func TestFindHelperForValue(t *testing.T) {
	vm := NewVM()

	// Setup helpers for different types
	intHelper := &HelperInfo{
		Name:       "IntHelper",
		TargetType: "Integer",
		Methods:    map[string]uint16{"ToString": 1},
	}
	floatHelper := &HelperInfo{
		Name:       "FloatHelper",
		TargetType: "Float",
		Methods:    map[string]uint16{"Round": 2},
	}
	stringHelper := &HelperInfo{
		Name:       "StringHelper",
		TargetType: "String",
		Methods:    map[string]uint16{"ToUpper": 3},
	}
	boolHelper := &HelperInfo{
		Name:       "BoolHelper",
		TargetType: "Boolean",
		Methods:    map[string]uint16{"Negate": 4},
	}
	arrayHelper := &HelperInfo{
		Name:       "ArrayHelper",
		TargetType: "Array",
		Methods:    map[string]uint16{"Sort": 5},
	}
	// Add helper with different case
	intHelperUpper := &HelperInfo{
		Name:       "IntHelperUpper",
		TargetType: "INTEGER",
		Methods:    map[string]uint16{"ToHex": 6},
	}

	vm.helpers = map[string]*HelperInfo{
		"IntHelper":      intHelper,
		"FloatHelper":    floatHelper,
		"StringHelper":   stringHelper,
		"BoolHelper":     boolHelper,
		"ArrayHelper":    arrayHelper,
		"IntHelperUpper": intHelperUpper,
	}

	tests := []struct {
		name          string
		value         Value
		expectedType  string // Expected target type of helper (empty for no helper expected)
		shouldBeFound bool
	}{
		{
			name:          "integer value",
			value:         IntValue(42),
			expectedType:  "Integer",
			shouldBeFound: true,
		},
		{
			name:          "float value",
			value:         FloatValue(3.14),
			expectedType:  "Float",
			shouldBeFound: true,
		},
		{
			name:          "string value",
			value:         StringValue("hello"),
			expectedType:  "String",
			shouldBeFound: true,
		},
		{
			name:          "boolean value",
			value:         BoolValue(true),
			expectedType:  "Boolean",
			shouldBeFound: true,
		},
		{
			name:          "array value",
			value:         ArrayValue(NewArrayInstance([]Value{IntValue(1)})),
			expectedType:  "Array",
			shouldBeFound: true,
		},
		{
			name:          "nil value - no helper",
			value:         NilValue(),
			expectedType:  "",
			shouldBeFound: false,
		},
		{
			name:          "function value - no helper",
			value:         FunctionValue(&FunctionObject{Name: "test"}),
			expectedType:  "",
			shouldBeFound: false,
		},
		{
			name:          "object value - no helper",
			value:         ObjectValue(&ObjectInstance{}),
			expectedType:  "",
			shouldBeFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.findHelperForValue(tt.value)
			if tt.shouldBeFound {
				if result == nil {
					t.Errorf("findHelperForValue(%v) = nil, want helper for type %s", tt.value.Type, tt.expectedType)
					return
				}
				// Check that the helper's target type matches (case-insensitive)
				if !equalsCaseInsensitive(result.TargetType, tt.expectedType) {
					t.Errorf("findHelperForValue(%v) = helper for %q, want helper for %q", tt.value.Type, result.TargetType, tt.expectedType)
				}
			} else {
				if result != nil {
					t.Errorf("findHelperForValue(%v) = %q, want nil", tt.value.Type, result.Name)
				}
			}
		})
	}
}

// TestFindHelperForValueCaseInsensitive tests case-insensitive helper lookup
func TestFindHelperForValueCaseInsensitive(t *testing.T) {
	vm := NewVM()

	// Setup helpers with various case combinations
	vm.helpers = map[string]*HelperInfo{
		"IntHelper1": {Name: "IntHelper1", TargetType: "integer", Methods: map[string]uint16{"M1": 1}},
		"IntHelper2": {Name: "IntHelper2", TargetType: "INTEGER", Methods: map[string]uint16{"M2": 2}},
		"IntHelper3": {Name: "IntHelper3", TargetType: "InTeGeR", Methods: map[string]uint16{"M3": 3}},
		"StrHelper1": {Name: "StrHelper1", TargetType: "string", Methods: map[string]uint16{"M4": 4}},
		"StrHelper2": {Name: "StrHelper2", TargetType: "STRING", Methods: map[string]uint16{"M5": 5}},
	}

	// All these should find a helper (any one that matches case-insensitively)
	intValue := IntValue(42)
	helper := vm.findHelperForValue(intValue)
	if helper == nil {
		t.Error("findHelperForValue should find helper for integer with case-insensitive match")
	}
	if helper != nil && !equalsCaseInsensitive(helper.TargetType, "Integer") {
		t.Errorf("findHelperForValue should find Integer helper, got helper for %q", helper.TargetType)
	}

	strValue := StringValue("test")
	helper = vm.findHelperForValue(strValue)
	if helper == nil {
		t.Error("findHelperForValue should find helper for string with case-insensitive match")
	}
	if helper != nil && !equalsCaseInsensitive(helper.TargetType, "String") {
		t.Errorf("findHelperForValue should find String helper, got helper for %q", helper.TargetType)
	}
}

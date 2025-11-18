package bytecode

import (
	"testing"
)

// TestResolveValueType tests value type resolution from string names
func TestResolveValueType(t *testing.T) {
	tests := []struct {
		typeName string
		expected ValueType
	}{
		{"Integer", ValueInt},
		{"int", ValueInt},
		{"INT", ValueInt},
		{"Float", ValueFloat},
		{"float", ValueFloat},
		{"FLOAT", ValueFloat},
		{"String", ValueString},
		{"string", ValueString},
		{"STRING", ValueString},
		{"Boolean", ValueBool},
		{"boolean", ValueBool},
		{"bool", ValueBool},
		{"BOOL", ValueBool},
		{"unknown", ValueNil}, // Unknown types default to nil
		{"", ValueNil},        // Empty string defaults to nil
		{"random", ValueNil},  // Random string defaults to nil
		{"Array", ValueNil},   // Complex types default to nil
		{"Object", ValueNil},  // Complex types default to nil
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			result := resolveValueType(tt.typeName)
			if result != tt.expected {
				t.Errorf("resolveValueType(%q) = %v, want %v", tt.typeName, result, tt.expected)
			}
		})
	}
}

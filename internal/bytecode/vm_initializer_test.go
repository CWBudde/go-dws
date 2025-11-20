package bytecode

import (
	"testing"
)

// TestVM_ExecuteInitializer tests the executeInitializer function
func TestVM_ExecuteInitializer(t *testing.T) {
	tests := []struct {
		createChunk   func() *Chunk
		setupParentVM func(*VM)
		expectedValue Value
		name          string
		errorContains string
		expectError   bool
	}{
		{
			name: "simple_integer_initializer",
			createChunk: func() *Chunk {
				chunk := NewChunk("init")
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
				chunk.WriteInstruction(MakeInstruction(OpReturn, 1, 0), 1)
				chunk.Constants = []Value{IntValue(42)}
				return chunk
			},
			setupParentVM: func(vm *VM) {},
			expectedValue: IntValue(42),
			expectError:   false,
		},
		{
			name: "string_initializer",
			createChunk: func() *Chunk {
				chunk := NewChunk("init")
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
				chunk.WriteInstruction(MakeInstruction(OpReturn, 1, 0), 1)
				chunk.Constants = []Value{StringValue("hello")}
				return chunk
			},
			setupParentVM: func(vm *VM) {},
			expectedValue: StringValue("hello"),
			expectError:   false,
		},
		{
			name: "computed_initializer",
			createChunk: func() *Chunk {
				// Compute 10 + 20
				chunk := NewChunk("init")
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 1), 1)
				chunk.WriteInstruction(MakeSimpleInstruction(OpAddInt), 1)
				chunk.WriteInstruction(MakeInstruction(OpReturn, 1, 0), 1)
				chunk.Constants = []Value{IntValue(10), IntValue(20)}
				return chunk
			},
			setupParentVM: func(vm *VM) {},
			expectedValue: IntValue(30),
			expectError:   false,
		},
		{
			name: "initializer_using_global",
			createChunk: func() *Chunk {
				// Load global at index 100 and return it (index 100 is beyond built-in functions)
				chunk := NewChunk("init")
				chunk.WriteInstruction(MakeInstruction(OpLoadGlobal, 0, 100), 1)
				chunk.WriteInstruction(MakeInstruction(OpReturn, 1, 0), 1)
				return chunk
			},
			setupParentVM: func(vm *VM) {
				// Set a global value that the initializer will access
				// Use index 100 which is beyond built-in functions (0-50)
				vm.globals = make([]Value, 150)
				vm.globals[100] = IntValue(100)
			},
			expectedValue: IntValue(100),
			expectError:   false,
		},
		{
			name: "initializer_with_helper_access",
			createChunk: func() *Chunk {
				chunk := NewChunk("init")
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
				chunk.WriteInstruction(MakeInstruction(OpReturn, 1, 0), 1)
				chunk.Constants = []Value{IntValue(5)}
				// Helpers will be inherited from parent VM
				chunk.Helpers = map[string]*HelperInfo{
					"testhelper": {
						Name:       "TestHelper",
						TargetType: "Integer",
						Methods:    map[string]uint16{},
					},
				}
				return chunk
			},
			setupParentVM: func(vm *VM) {
				// Set up helpers that the initializer can access
				vm.helpers = map[string]*HelperInfo{
					"testhelper": {
						Name:       "TestHelper",
						TargetType: "Integer",
						Methods:    map[string]uint16{},
					},
				}
			},
			expectedValue: IntValue(5),
			expectError:   false,
		},
		{
			name:          "nil_chunk",
			createChunk:   func() *Chunk { return nil },
			setupParentVM: func(vm *VM) {},
			expectError:   true,
			errorContains: "nil chunk",
		},
		{
			name: "initializer_with_error",
			createChunk: func() *Chunk {
				// Try to divide by zero
				chunk := NewChunk("init")
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 1), 1)
				chunk.WriteInstruction(MakeSimpleInstruction(OpDivInt), 1)
				chunk.WriteInstruction(MakeInstruction(OpReturn, 1, 0), 1)
				chunk.Constants = []Value{IntValue(10), IntValue(0)}
				return chunk
			},
			setupParentVM: func(vm *VM) {},
			expectError:   true,
			errorContains: "division by zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parentVM := NewVM()
			tt.setupParentVM(parentVM)

			chunk := tt.createChunk()
			result, err := parentVM.executeInitializer(chunk)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errorContains != "" {
					if !contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !valuesEqual(result, tt.expectedValue) {
					t.Errorf("Expected value %v, got %v", tt.expectedValue, result)
				}
			}
		})
	}
}

// TestVM_ExecuteInitializer_InheritGlobals tests that initializers inherit globals from parent VM
func TestVM_ExecuteInitializer_InheritGlobals(t *testing.T) {
	parentVM := NewVM()

	// Set up globals in parent VM (use indices beyond built-in functions)
	parentVM.globals = make([]Value, 150)
	parentVM.globals[100] = IntValue(10)
	parentVM.globals[101] = IntValue(20)
	parentVM.globals[102] = StringValue("parent")

	// Create an initializer that reads and modifies globals
	initChunk := NewChunk("init")
	// Load global[100], load global[101], add them
	initChunk.WriteInstruction(MakeInstruction(OpLoadGlobal, 0, 100), 1)
	initChunk.WriteInstruction(MakeInstruction(OpLoadGlobal, 0, 101), 1)
	initChunk.WriteInstruction(MakeSimpleInstruction(OpAddInt), 1)
	initChunk.WriteInstruction(MakeInstruction(OpReturn, 1, 0), 1)

	result, err := parentVM.executeInitializer(initChunk)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return 30 (10 + 20)
	if !valuesEqual(result, IntValue(30)) {
		t.Errorf("Expected 30, got %v", result)
	}

	// Parent globals should still be intact
	if !valuesEqual(parentVM.globals[100], IntValue(10)) {
		t.Error("Parent global[100] should not be modified")
	}
}

// TestVM_ExecuteInitializer_InheritHelpers tests that initializers inherit helpers from parent VM
func TestVM_ExecuteInitializer_InheritHelpers(t *testing.T) {
	parentVM := NewVM()

	// Set up helpers in parent VM
	parentVM.helpers = map[string]*HelperInfo{
		"tstringhelper": {
			Name:       "TStringHelper",
			TargetType: "String",
			Methods: map[string]uint16{
				"length": 0,
			},
		},
	}

	// Create a simple initializer
	initChunk := NewChunk("init")
	initChunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
	initChunk.WriteInstruction(MakeInstruction(OpReturn, 1, 0), 1)
	initChunk.Constants = []Value{IntValue(42)}

	result, err := parentVM.executeInitializer(initChunk)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !valuesEqual(result, IntValue(42)) {
		t.Errorf("Expected 42, got %v", result)
	}

	// Verify parent helpers are still intact
	if len(parentVM.helpers) != 1 {
		t.Error("Parent helpers should not be modified")
	}
}

// TestVM_ExecuteInitializer_IsolatedExecution tests that initializer execution is isolated
func TestVM_ExecuteInitializer_IsolatedExecution(t *testing.T) {
	parentVM := NewVM()

	// Set initial stack state in parent VM
	parentVM.stack = append(parentVM.stack, IntValue(999))

	// Create an initializer
	initChunk := NewChunk("init")
	initChunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
	initChunk.WriteInstruction(MakeInstruction(OpReturn, 1, 0), 1)
	initChunk.Constants = []Value{IntValue(42)}

	result, err := parentVM.executeInitializer(initChunk)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !valuesEqual(result, IntValue(42)) {
		t.Errorf("Expected 42, got %v", result)
	}

	// Parent VM stack should not be affected by initializer execution
	if len(parentVM.stack) != 1 || !valuesEqual(parentVM.stack[0], IntValue(999)) {
		t.Error("Parent VM stack should not be affected by initializer execution")
	}
}

// TestVM_ExecuteInitializer_ComplexExpression tests complex initializer expressions
func TestVM_ExecuteInitializer_ComplexExpression(t *testing.T) {
	tests := []struct {
		name          string
		setupChunk    func() *Chunk
		expectedValue Value
	}{
		{
			name: "arithmetic_expression",
			setupChunk: func() *Chunk {
				// (10 + 20) * 3
				chunk := NewChunk("init")
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1) // 10
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 1), 1) // 20
				chunk.WriteInstruction(MakeSimpleInstruction(OpAddInt), 1)    // 30
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 2), 1) // 3
				chunk.WriteInstruction(MakeSimpleInstruction(OpMulInt), 1)    // 90
				chunk.WriteInstruction(MakeInstruction(OpReturn, 1, 0), 1)
				chunk.Constants = []Value{IntValue(10), IntValue(20), IntValue(3)}
				return chunk
			},
			expectedValue: IntValue(90),
		},
		{
			name: "string_concatenation",
			setupChunk: func() *Chunk {
				// "hello" + " " + "world"
				chunk := NewChunk("init")
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 1), 1)
				chunk.WriteInstruction(MakeSimpleInstruction(OpStringConcat), 1)
				chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 2), 1)
				chunk.WriteInstruction(MakeSimpleInstruction(OpStringConcat), 1)
				chunk.WriteInstruction(MakeInstruction(OpReturn, 1, 0), 1)
				chunk.Constants = []Value{
					StringValue("hello"),
					StringValue(" "),
					StringValue("world"),
				}
				return chunk
			},
			expectedValue: StringValue("hello world"),
		},
		{
			name: "boolean_expression",
			setupChunk: func() *Chunk {
				// true and false
				chunk := NewChunk("init")
				chunk.WriteInstruction(MakeSimpleInstruction(OpLoadTrue), 1)
				chunk.WriteInstruction(MakeSimpleInstruction(OpLoadFalse), 1)
				chunk.WriteInstruction(MakeSimpleInstruction(OpAnd), 1)
				chunk.WriteInstruction(MakeInstruction(OpReturn, 1, 0), 1)
				return chunk
			},
			expectedValue: BoolValue(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parentVM := NewVM()
			chunk := tt.setupChunk()

			result, err := parentVM.executeInitializer(chunk)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !valuesEqual(result, tt.expectedValue) {
				t.Errorf("Expected %v, got %v", tt.expectedValue, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

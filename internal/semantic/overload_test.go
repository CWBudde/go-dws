package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Unit Tests for Overload Set Storage and Retrieval
// ============================================================================

// ============================================================================
// Part A: Storing Multiple Overloads
// ============================================================================

func TestDefineOverload_TwoFunctionsDifferentParamCount(t *testing.T) {
	st := NewSymbolTable()

	// Define first overload: Max(a, b: Integer): Integer
	func1 := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER, types.INTEGER}, // params
		[]string{"a", "b"},                         // names
		nil,                                        // defaults
		[]bool{false, false},                       // lazy flags
		[]bool{false, false},                       // var flags
		[]bool{false, false},                       // const flags
		types.INTEGER,                              // return type
	)
	err := st.DefineOverload("Max", func1, true, false)
	if err != nil {
		t.Fatalf("DefineOverload failed for first overload: %v", err)
	}

	// Define second overload: Max(a, b, c: Integer): Integer
	func2 := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER, types.INTEGER, types.INTEGER}, // params
		[]string{"a", "b", "c"},                                   // names
		nil,                                                       // defaults
		[]bool{false, false, false},                               // lazy
		[]bool{false, false, false},                               // var
		[]bool{false, false, false},                               // const
		types.INTEGER,                                             // return type
	)
	err = st.DefineOverload("Max", func2, true, false)
	if err != nil {
		t.Fatalf("DefineOverload failed for second overload: %v", err)
	}

	// Verify overload set
	overloads := st.GetOverloadSet("Max")
	if overloads == nil {
		t.Fatal("Expected overload set, got nil")
	}
	if len(overloads) != 2 {
		t.Fatalf("Expected 2 overloads, got %d", len(overloads))
	}
}

func TestDefineOverload_TwoFunctionsSameCountDifferentTypes(t *testing.T) {
	st := NewSymbolTable()

	// Define first overload: Convert(value: Integer): String
	func1 := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"value"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)
	err := st.DefineOverload("Convert", func1, true, false)
	if err != nil {
		t.Fatalf("DefineOverload failed for first overload: %v", err)
	}

	// Define second overload: Convert(value: Float): String
	func2 := types.NewFunctionTypeWithMetadata([]types.Type{types.FLOAT}, // params
		[]string{"value"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)
	err = st.DefineOverload("Convert", func2, true, false)
	if err != nil {
		t.Fatalf("DefineOverload failed for second overload: %v", err)
	}

	// Verify overload set
	overloads := st.GetOverloadSet("Convert")
	if len(overloads) != 2 {
		t.Fatalf("Expected 2 overloads, got %d", len(overloads))
	}
}

func TestDefineOverload_ThreeOverloads(t *testing.T) {
	st := NewSymbolTable()

	// Define three overloads of Print
	func1 := types.NewFunctionTypeWithMetadata([]types.Type{types.STRING}, // params
		[]string{"msg"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.VOID,    // return type
	)
	func2 := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"value"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.VOID,    // return type
	)
	func3 := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.STRING, types.INTEGER}, // params
		[]string{"msg", "count"},                  // names
		nil,                                       // defaults
		[]bool{false, false},                      // lazy
		[]bool{false, false},                      // var
		[]bool{false, false},                      // const
		types.VOID,                                // return type
	)

	st.DefineOverload("Print", func1, true, false)
	st.DefineOverload("Print", func2, true, false)
	err := st.DefineOverload("Print", func3, true, false)
	if err != nil {
		t.Fatalf("DefineOverload failed for third overload: %v", err)
	}

	overloads := st.GetOverloadSet("Print")
	if len(overloads) != 3 {
		t.Fatalf("Expected 3 overloads, got %d", len(overloads))
	}
}

func TestDefineOverload_DifferentReturnTypes(t *testing.T) {
	st := NewSymbolTable()

	// In DWScript, overloads can differ only in return type is allowed
	// Get(key: String): String
	func1 := types.NewFunctionTypeWithMetadata([]types.Type{types.STRING}, // params
		[]string{"key"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)
	// Get(key: String): Integer
	func2 := types.NewFunctionTypeWithMetadata([]types.Type{types.STRING}, // params
		[]string{"key"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.INTEGER, // return type
	)

	st.DefineOverload("Get", func1, true, false)
	err := st.DefineOverload("Get", func2, true, false)
	if err != nil {
		t.Fatalf("DefineOverload failed: %v", err)
	}

	overloads := st.GetOverloadSet("Get")
	if len(overloads) != 2 {
		t.Fatalf("Expected 2 overloads, got %d", len(overloads))
	}
}

func TestDefineOverload_ProceduresAndFunctions(t *testing.T) {
	st := NewSymbolTable()

	// Procedure: DoSomething(x: Integer)
	proc := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.VOID,    // return type
	)
	// Function: DoSomething(x: Integer): String
	func1 := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)

	st.DefineOverload("DoSomething", proc, true, false)
	err := st.DefineOverload("DoSomething", func1, true, false)
	if err != nil {
		t.Fatalf("DefineOverload failed: %v", err)
	}

	overloads := st.GetOverloadSet("DoSomething")
	if len(overloads) != 2 {
		t.Fatalf("Expected 2 overloads, got %d", len(overloads))
	}
}

// ============================================================================
// Part B: Retrieving Overload Sets
// ============================================================================

func TestGetOverloadSet_MultipleOverloads(t *testing.T) {
	st := NewSymbolTable()

	func1 := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)
	func2 := types.NewFunctionTypeWithMetadata([]types.Type{types.STRING}, // params
		[]string{"s"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)

	st.DefineOverload("Format", func1, true, false)
	st.DefineOverload("Format", func2, true, false)

	overloads := st.GetOverloadSet("Format")
	if overloads == nil {
		t.Fatal("Expected overload set, got nil")
	}
	if len(overloads) != 2 {
		t.Fatalf("Expected 2 overloads, got %d", len(overloads))
	}

	// Verify both overloads are present
	if overloads[0].Type == nil || overloads[1].Type == nil {
		t.Fatal("Overload types should not be nil")
	}
}

func TestGetOverloadSet_SingleFunction(t *testing.T) {
	st := NewSymbolTable()

	funcType := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.INTEGER, // return type
	)
	st.DefineFunction("Square", funcType)

	// GetOverloadSet should return single-element slice for non-overloaded functions
	overloads := st.GetOverloadSet("Square")
	if overloads == nil {
		t.Fatal("Expected single-element slice, got nil")
	}
	if len(overloads) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(overloads))
	}
	if overloads[0].Name != "Square" {
		t.Fatalf("Expected name 'Square', got '%s'", overloads[0].Name)
	}
}

func TestGetOverloadSet_NonExistentFunction(t *testing.T) {
	st := NewSymbolTable()

	overloads := st.GetOverloadSet("DoesNotExist")
	if overloads != nil {
		t.Fatalf("Expected nil for non-existent function, got %v", overloads)
	}
}

func TestGetOverloadSet_CaseInsensitive(t *testing.T) {
	st := NewSymbolTable()

	funcType := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.INTEGER, // return type
	)
	st.DefineOverload("MyFunc", funcType, true, false)

	// DWScript is case-insensitive
	overloads1 := st.GetOverloadSet("MyFunc")
	overloads2 := st.GetOverloadSet("myfunc")
	overloads3 := st.GetOverloadSet("MYFUNC")

	if overloads1 == nil || overloads2 == nil || overloads3 == nil {
		t.Fatal("GetOverloadSet should be case-insensitive")
	}
	if len(overloads1) != 1 || len(overloads2) != 1 || len(overloads3) != 1 {
		t.Fatal("All case variations should return same result")
	}
}

// ============================================================================
// Part C: Conflict Detection
// ============================================================================

func TestDefineOverload_DuplicateSignatureError(t *testing.T) {
	st := NewSymbolTable()

	funcType := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER, types.STRING}, // params
		[]string{"x", "s"},                        // names
		nil,                                       // defaults
		[]bool{false, false},                      // lazy
		[]bool{false, false},                      // var
		[]bool{false, false},                      // const
		types.VOID,                                // return type
	)

	// Define first time
	err := st.DefineOverload("Process", funcType, true, false)
	if err != nil {
		t.Fatalf("First DefineOverload should succeed: %v", err)
	}

	// Try to define exact same signature again
	err = st.DefineOverload("Process", funcType, true, false)
	if err == nil {
		t.Fatal("Expected error for duplicate signature, got nil")
	}
	// Task 9.59: Error message matches DWScript format (Task 9.50: Fixed capitalization)
	expectedMsg := "There is already a method with name \"Process\""
	if err.Error() != expectedMsg {
		t.Fatalf("Wrong error message: got %q, expected %q", err.Error(), expectedMsg)
	}
}

func TestDefineOverload_MissingOverloadDirectiveError(t *testing.T) {
	st := NewSymbolTable()

	func1 := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)
	func2 := types.NewFunctionTypeWithMetadata([]types.Type{types.STRING}, // params
		[]string{"s"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)

	// Define first overload with directive
	st.DefineOverload("Convert", func1, true, false)

	// Try to define second without directive (hasOverloadDirective = false)
	err := st.DefineOverload("Convert", func2, false, false)
	if err == nil {
		t.Fatal("Expected error when second declaration lacks 'overload' directive")
	}
}

func TestDefineOverload_NonFunctionSymbolError(t *testing.T) {
	st := NewSymbolTable()

	// Define a variable
	st.Define("MyVar", types.INTEGER)

	// Try to define function with same name
	funcType := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)

	err := st.DefineOverload("MyVar", funcType, true, false)
	if err == nil {
		t.Fatal("Expected error when trying to overload a non-function symbol")
	}
	if err.Error() != "'MyVar' is already declared as a non-function symbol" {
		t.Fatalf("Wrong error message: %v", err)
	}
}

func TestDefineOverload_SuccessWithDifferentSignatures(t *testing.T) {
	st := NewSymbolTable()

	func1 := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)
	func2 := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER, types.INTEGER}, // params
		[]string{"x", "y"},                         // names
		nil,                                        // defaults
		[]bool{false, false},                       // lazy
		[]bool{false, false},                       // var
		[]bool{false, false},                       // const
		types.STRING,                               // return type
	)

	err1 := st.DefineOverload("Build", func1, true, false)
	err2 := st.DefineOverload("Build", func2, true, false)

	if err1 != nil || err2 != nil {
		t.Fatal("Both overload definitions should succeed with different signatures")
	}
}

func TestDefineOverload_BothHaveOverloadDirective(t *testing.T) {
	st := NewSymbolTable()

	func1 := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.VOID,    // return type
	)
	func2 := types.NewFunctionTypeWithMetadata([]types.Type{types.STRING}, // params
		[]string{"s"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.VOID,    // return type
	)

	// Both have overload directive = true
	err1 := st.DefineOverload("Log", func1, true, false)
	err2 := st.DefineOverload("Log", func2, true, false)

	if err1 != nil || err2 != nil {
		t.Fatal("Both definitions should succeed when both have overload directive")
	}
}

func TestDefineOverload_DifferentParameterModifiers(t *testing.T) {
	st := NewSymbolTable()

	// Swap(var a, var b: Integer)
	func1 := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER, types.INTEGER}, // params
		[]string{"a", "b"},                         // names
		nil,                                        // defaults
		[]bool{false, false},                       // lazy
		[]bool{true, true},                         // var
		[]bool{false, false},                       // const
		types.VOID,                                 // return type
	)
	// Swap(const a, const b: Integer) - different modifiers
	func2 := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.INTEGER, types.INTEGER}, // params
		[]string{"a", "b"},                         // names
		nil,                                        // defaults
		[]bool{false, false},                       // lazy
		[]bool{false, false},                       // var
		[]bool{true, true},                         // const
		types.VOID,                                 // return type
	)

	st.DefineOverload("Swap", func1, true, false)
	err := st.DefineOverload("Swap", func2, true, false)

	// Task 9.59: SignaturesEqual now checks parameter modifiers (var/const/lazy),
	// so these two functions with different modifiers should be recognized as
	// having distinct signatures and both should be accepted as valid overloads.
	if err != nil {
		t.Fatalf("DefineOverload should succeed for functions with different parameter modifiers, got error: %v", err)
	}

	// Verify both overloads were stored
	overloads := st.GetOverloadSet("Swap")
	if len(overloads) != 2 {
		t.Fatalf("Expected 2 overloads, got %d", len(overloads))
	}
}

// ============================================================================
// Part D: Nested Scopes
// ============================================================================

func TestDefineOverload_GlobalScope(t *testing.T) {
	st := NewSymbolTable()

	func1 := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)
	func2 := types.NewFunctionTypeWithMetadata([]types.Type{types.STRING}, // params
		[]string{"s"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)

	st.DefineOverload("ToString", func1, true, false)
	st.DefineOverload("ToString", func2, true, false)

	overloads := st.GetOverloadSet("ToString")
	if len(overloads) != 2 {
		t.Fatalf("Expected 2 overloads in global scope, got %d", len(overloads))
	}
}

func TestDefineOverload_NestedScope(t *testing.T) {
	global := NewSymbolTable()
	local := NewEnclosedSymbolTable(global)

	// Define overload in global scope
	globalFunc := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)
	global.DefineOverload("Helper", globalFunc, true, false)

	// Define different overload in local scope
	localFunc := types.NewFunctionTypeWithMetadata([]types.Type{types.STRING}, // params
		[]string{"s"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)
	local.DefineOverload("Helper", localFunc, true, false)

	// Local scope should have its own overload set
	localOverloads := local.GetOverloadSet("Helper")
	if len(localOverloads) != 1 {
		t.Fatalf("Expected 1 overload in local scope, got %d", len(localOverloads))
	}

	// Global scope should still have its overload
	globalOverloads := global.GetOverloadSet("Helper")
	if len(globalOverloads) != 1 {
		t.Fatalf("Expected 1 overload in global scope, got %d", len(globalOverloads))
	}
}

func TestDefineOverload_InnerScopeHidesOuter(t *testing.T) {
	outer := NewSymbolTable()
	inner := NewEnclosedSymbolTable(outer)

	// Define in outer scope
	outerFunc := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)
	outer.DefineOverload("Calculate", outerFunc, true, false)

	// Define in inner scope (shadows outer)
	innerFunc := types.NewFunctionTypeWithMetadata([]types.Type{types.FLOAT}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)
	inner.DefineOverload("Calculate", innerFunc, true, false)

	// GetOverloadSet from inner scope should only see inner definition
	innerOverloads := inner.GetOverloadSet("Calculate")
	if len(innerOverloads) != 1 {
		t.Fatalf("Inner scope should see only 1 overload (local), got %d", len(innerOverloads))
	}

	// Check that it's the inner function (Float parameter)
	funcType := innerOverloads[0].Type.(*types.FunctionType)
	if len(funcType.Parameters) != 1 || funcType.Parameters[0] != types.FLOAT {
		t.Fatal("Inner scope should see the FLOAT version, not INTEGER")
	}
}

func TestDefineOverload_ResolveAcrossScopes(t *testing.T) {
	global := NewSymbolTable()
	local := NewEnclosedSymbolTable(global)

	// Define in global scope
	globalFunc := types.NewFunctionTypeWithMetadata([]types.Type{types.INTEGER}, // params
		[]string{"x"}, nil, // names
		[]bool{false}, // lazy
		[]bool{false}, // var
		[]bool{false}, // const
		types.STRING,  // return type
	)
	global.DefineOverload("Util", globalFunc, true, false)

	// Don't define in local scope - resolve should find global
	sym, ok := local.Resolve("Util")
	if !ok {
		t.Fatal("Should resolve global function from local scope")
	}
	if sym.Name != "Util" {
		t.Fatalf("Expected 'Util', got '%s'", sym.Name)
	}
}

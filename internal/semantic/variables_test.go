package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

func inferVariableType(t *testing.T, input, name string) types.Type {
	t.Helper()

	analyzer, err := analyzeSource(t, input)
	if err != nil {
		t.Fatalf("unexpected semantic error: %v", err)
	}

	sym, ok := analyzer.symbols.Resolve(name)
	if !ok {
		t.Fatalf("variable %q was not registered", name)
	}

	return sym.Type
}

func TestVarTypeInferenceInteger(t *testing.T) {
	typ := inferVariableType(t, "var x = 42;", "x")
	if !typ.Equals(types.INTEGER) {
		t.Fatalf("expected Integer type, got %s", typ.String())
	}
}

func TestVarTypeInferenceFloat(t *testing.T) {
	typ := inferVariableType(t, "var f = 3.14;", "f")
	if !typ.Equals(types.FLOAT) {
		t.Fatalf("expected Float type, got %s", typ.String())
	}
}

func TestVarTypeInferenceString(t *testing.T) {
	typ := inferVariableType(t, `var s = "hello";`, "s")
	if !typ.Equals(types.STRING) {
		t.Fatalf("expected String type, got %s", typ.String())
	}
}

func TestVarTypeInferenceArray(t *testing.T) {
	typ := inferVariableType(t, "var arr = [1, 2, 3];", "arr")
	arrayType, ok := types.GetUnderlyingType(typ).(*types.ArrayType)
	if !ok {
		t.Fatalf("expected array type, got %s", typ.String())
	}

	if !arrayType.ElementType.Equals(types.INTEGER) {
		t.Fatalf("expected array of Integer, got array of %s", arrayType.ElementType.String())
	}
}

func TestVarTypeInferenceEmptyArrayError(t *testing.T) {
	expectError(t, "var x = [];", "cannot infer type for variable 'x'")
}

func TestVarTypeInferenceNilInitializerError(t *testing.T) {
	expectError(t, "var n = nil;", "cannot infer type for variable 'n' from nil initializer")
}

// ============================================================================
// Variant Variable Tests
// ============================================================================

// Task 9.144 & 9.147: Test Variant variable declarations
func TestVariantVarDeclaration(t *testing.T) {
	// Uninitialized Variant
	typ := inferVariableType(t, "var v: Variant;", "v")
	if !typ.Equals(types.VARIANT) {
		t.Fatalf("expected Variant type, got %s", typ.String())
	}
}

// Task 9.144 & 9.147: Test Variant variable with Integer initializer
func TestVariantVarWithIntegerInit(t *testing.T) {
	typ := inferVariableType(t, "var v: Variant := 42;", "v")
	if !typ.Equals(types.VARIANT) {
		t.Fatalf("expected Variant type, got %s", typ.String())
	}
}

// Task 9.144 & 9.147: Test Variant variable with String initializer
func TestVariantVarWithStringInit(t *testing.T) {
	typ := inferVariableType(t, `var v: Variant := "hello";`, "v")
	if !typ.Equals(types.VARIANT) {
		t.Fatalf("expected Variant type, got %s", typ.String())
	}
}

// Task 9.223 & 9.226: Test Variant variable with Float initializer
func TestVariantVarWithFloatInit(t *testing.T) {
	typ := inferVariableType(t, "var v: Variant := 3.14;", "v")
	if !typ.Equals(types.VARIANT) {
		t.Fatalf("expected Variant type, got %s", typ.String())
	}
}

// Task 9.223 & 9.226: Test Variant variable with Boolean initializer
func TestVariantVarWithBooleanInit(t *testing.T) {
	typ := inferVariableType(t, "var v: Variant := true;", "v")
	if !typ.Equals(types.VARIANT) {
		t.Fatalf("expected Variant type, got %s", typ.String())
	}
}

// Task 9.224 & 9.226: Test assignment of different types to Variant
func TestVariantAssignmentFromInteger(t *testing.T) {
	input := `
	var v: Variant;
	var i: Integer := 10;
	begin
		v := i;
	end.`
	expectNoErrors(t, input)
}

func TestVariantAssignmentFromString(t *testing.T) {
	input := `
	var v: Variant;
	var s: String := "test";
	begin
		v := s;
	end.`
	expectNoErrors(t, input)
}

func TestVariantAssignmentFromFloat(t *testing.T) {
	input := `
	var v: Variant;
	var f: Float := 2.5;
	begin
		v := f;
	end.`
	expectNoErrors(t, input)
}

func TestVariantAssignmentFromBoolean(t *testing.T) {
	input := `
	var v: Variant;
	var b: Boolean := false;
	begin
		v := b;
	end.`
	expectNoErrors(t, input)
}

// Task 9.224 & 9.226: Test Variant-to-Variant assignment
func TestVariantToVariantAssignment(t *testing.T) {
	input := `
	var v1: Variant := 42;
	var v2: Variant;
	begin
		v2 := v1;
	end.`
	expectNoErrors(t, input)
}

// Task 9.226: Test type inference does NOT infer Variant from Variant initializer
// (should use actual wrapped type semantics in later tasks)
func TestVariantInferencePreservesVariantType(t *testing.T) {
	typ := inferVariableType(t, "var v1: Variant := 42; var v2 := v1;", "v2")
	// v2 should infer Variant type from v1
	if !typ.Equals(types.VARIANT) {
		t.Fatalf("expected to infer Variant type, got %s", typ.String())
	}
}

// Task 9.225 & 9.226: Test heterogeneous array with Variant element type
func TestHeterogeneousArrayWithVariantType(t *testing.T) {
	input := `
	var arr: array of Variant := [1, "hello", 3.14, true];
	`
	expectNoErrors(t, input)
}

// Task 9.225 & 9.226: Test empty array with Variant element type
func TestEmptyVariantArray(t *testing.T) {
	input := `
	var arr: array of Variant := [];
	`
	expectNoErrors(t, input)
}

// Task 9.226: Test multiple Variant variables
func TestMultipleVariantDeclarations(t *testing.T) {
	input := `
	var v1: Variant := 10;
	var v2: Variant := "test";
	var v3: Variant := 3.14;
	`
	analyzer, err := analyzeSource(t, input)
	if err != nil {
		t.Fatalf("unexpected semantic error: %v", err)
	}

	// Check all three are Variant type
	for _, name := range []string{"v1", "v2", "v3"} {
		sym, ok := analyzer.symbols.Resolve(name)
		if !ok {
			t.Fatalf("variable %q was not registered", name)
		}
		if !sym.Type.Equals(types.VARIANT) {
			t.Fatalf("expected Variant type for %s, got %s", name, sym.Type.String())
		}
	}
}

// Task 9.226: Test Variant with type alias
func TestVariantWithTypeAlias(t *testing.T) {
	input := `
	type MyVariant = Variant;
	var v: MyVariant := 42;
	`
	analyzer, err := analyzeSource(t, input)
	if err != nil {
		t.Fatalf("unexpected semantic error: %v", err)
	}

	sym, ok := analyzer.symbols.Resolve("v")
	if !ok {
		t.Fatal("variable 'v' was not registered")
	}

	// Underlying type should be Variant
	underlying := types.GetUnderlyingType(sym.Type)
	if !underlying.Equals(types.VARIANT) {
		t.Fatalf("expected underlying Variant type, got %s", underlying.String())
	}
}

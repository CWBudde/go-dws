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

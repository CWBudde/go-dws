package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// TestGetLowBound_NilAdapter tests that GetLowBound returns an error when adapter is nil.
func TestGetLowBound_NilAdapter(t *testing.T) {
	e := createTestEvaluator()
	// adapter is nil by default

	arrayVal := &runtime.ArrayValue{
		ArrayType: types.NewStaticArrayType(types.INTEGER, 1, 10),
		Elements:  make([]runtime.Value, 10),
	}

	_, err := e.GetLowBound(arrayVal)
	if err == nil {
		t.Error("GetLowBound with nil adapter should return error")
	}
}

// TestGetHighBound_NilAdapter tests that GetHighBound returns an error when adapter is nil.
func TestGetHighBound_NilAdapter(t *testing.T) {
	e := createTestEvaluator()
	// adapter is nil by default

	arrayVal := &runtime.ArrayValue{
		ArrayType: types.NewStaticArrayType(types.INTEGER, 1, 10),
		Elements:  make([]runtime.Value, 10),
	}

	_, err := e.GetHighBound(arrayVal)
	if err == nil {
		t.Error("GetHighBound with nil adapter should return error")
	}
}

// Note: Full functional tests for GetLowBound/GetHighBound are in the Interpreter's
// builtins_context_test.go file, since these methods use the delegation pattern and
// delegate to the Interpreter's implementation via the adapter interface.
//
// The Interpreter implements both InterpreterAdapter and builtins.Context interfaces,
// so when used in production (via the Interpreter), these methods will correctly
// delegate to the full Interpreter implementation which handles:
// - Static and dynamic arrays
// - Type meta-values (Integer, Float, Boolean, enum types)
// - Enum values (requires enum registry lookup)
//
// These tests verify only that the delegation mechanism works correctly (i.e., that
// an error is returned when no adapter is present). The actual bounds logic is tested
// comprehensively in the Interpreter's test suite.

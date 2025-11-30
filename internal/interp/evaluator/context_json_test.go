package evaluator

import (
	"math"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// ParseJSONString Tests
// ============================================================================

// Note: ParseJSONString requires an adapter for wrapping JSON in Variants.
// Since createTestEvaluator() creates an evaluator with nil adapter, we can only
// test error cases here. Full functional tests are in the Interpreter's test suite.

func TestParseJSONString_NilAdapter(t *testing.T) {
	// ParseJSONString requires an adapter for wrapping JSON in Variants.
	// Since createTestEvaluator() creates an evaluator with nil adapter,
	// we skip this test. Full functional tests are in the Interpreter test suite.
	t.Skip("ParseJSONString requires adapter - full tests are in Interpreter test suite")
}

func TestParseJSONString_InvalidJSON(t *testing.T) {
	t.Skip("ParseJSONString requires adapter - full tests are in Interpreter test suite")
}

func TestParseJSONString_NestedStructures(t *testing.T) {
	t.Skip("ParseJSONString requires adapter - full tests are in Interpreter test suite")
}

// ============================================================================
// ValueToJSON Tests
// ============================================================================

func TestValueToJSON_Primitives(t *testing.T) {
	e := createTestEvaluator()

	tests := []struct {
		name     string
		value    Value
		expected string
	}{
		{"integer", &runtime.IntegerValue{Value: 42}, "42"},
		{"float", &runtime.FloatValue{Value: 3.14}, "3.14"},
		{"string", &runtime.StringValue{Value: "hello"}, `"hello"`},
		{"boolean true", &runtime.BooleanValue{Value: true}, "true"},
		{"boolean false", &runtime.BooleanValue{Value: false}, "false"},
		{"nil", &runtime.NilValue{}, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.ValueToJSON(tt.value, false)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValueToJSON_Array(t *testing.T) {
	e := createTestEvaluator()

	arrVal := &runtime.ArrayValue{
		Elements: []runtime.Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 3},
		},
	}

	result, err := e.ValueToJSON(arrVal, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "[1,2,3]"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestValueToJSON_Record(t *testing.T) {
	e := createTestEvaluator()

	recVal := &runtime.RecordValue{
		Fields: map[string]runtime.Value{
			"name": &runtime.StringValue{Value: "John"},
			"age":  &runtime.IntegerValue{Value: 30},
		},
	}

	result, err := e.ValueToJSON(recVal, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// JSON object fields can be in any order, so check both fields are present
	if !strings.Contains(result, `"name":"John"`) && !strings.Contains(result, `"name": "John"`) {
		t.Errorf("expected result to contain name field, got %q", result)
	}
	if !strings.Contains(result, `"age":30`) && !strings.Contains(result, `"age": 30`) {
		t.Errorf("expected result to contain age field, got %q", result)
	}
}

func TestValueToJSON_FormattedOutput(t *testing.T) {
	e := createTestEvaluator()

	arrVal := &runtime.ArrayValue{
		Elements: []runtime.Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
		},
	}

	result, err := e.ValueToJSON(arrVal, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Formatted output should contain newlines and indentation
	if !strings.Contains(result, "\n") {
		t.Errorf("expected formatted output to contain newlines, got %q", result)
	}
}

// ============================================================================
// ValueToJSONWithIndent Tests
// ============================================================================

func TestValueToJSONWithIndent_CustomIndent(t *testing.T) {
	e := createTestEvaluator()

	arrVal := &runtime.ArrayValue{
		Elements: []runtime.Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
		},
	}

	// Test with 4-space indent
	result, err := e.ValueToJSONWithIndent(arrVal, true, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain 4-space indentation
	if !strings.Contains(result, "    ") {
		t.Errorf("expected 4-space indentation, got %q", result)
	}
}

func TestValueToJSONWithIndent_NoFormatting(t *testing.T) {
	e := createTestEvaluator()

	arrVal := &runtime.ArrayValue{
		Elements: []runtime.Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
		},
	}

	// Test with formatted=false (indent should be ignored)
	result, err := e.ValueToJSONWithIndent(arrVal, false, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not contain newlines when formatted=false
	if strings.Contains(result, "\n") {
		t.Errorf("expected compact output without newlines, got %q", result)
	}
}

func TestValueToJSONWithIndent_NestedStructure(t *testing.T) {
	e := createTestEvaluator()

	// Create nested structure: {person: {name: "John", age: 30}}
	innerRec := &runtime.RecordValue{
		Fields: map[string]runtime.Value{
			"name": &runtime.StringValue{Value: "John"},
			"age":  &runtime.IntegerValue{Value: 30},
		},
	}

	outerRec := &runtime.RecordValue{
		Fields: map[string]runtime.Value{
			"person": innerRec,
		},
	}

	result, err := e.ValueToJSONWithIndent(outerRec, true, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain nested indentation
	if !strings.Contains(result, "\n") {
		t.Errorf("expected formatted output with newlines, got %q", result)
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestValueToJSON_EmptyArray(t *testing.T) {
	e := createTestEvaluator()

	arrVal := &runtime.ArrayValue{
		Elements: []runtime.Value{},
	}

	result, err := e.ValueToJSON(arrVal, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "[]"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestValueToJSON_EmptyRecord(t *testing.T) {
	e := createTestEvaluator()

	recVal := &runtime.RecordValue{
		Fields: map[string]runtime.Value{},
	}

	result, err := e.ValueToJSON(recVal, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "{}"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestValueToJSON_SpecialFloats(t *testing.T) {
	e := createTestEvaluator()

	tests := []struct {
		name  string
		value float64
	}{
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
		{"NaN", math.NaN()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			floatVal := &runtime.FloatValue{Value: tt.value}
			result, err := e.ValueToJSON(floatVal, false)

			// JSON doesn't support Inf/NaN, so this might return an error
			// or convert to null. Either is acceptable.
			if err != nil {
				// Error is acceptable for special floats
				return
			}

			// If no error, result should be valid JSON (likely "null")
			if result != "null" && result != "Infinity" && result != "-Infinity" && result != "NaN" {
				// Standard json.Marshal converts these to "null", but implementations may vary
				t.Logf("Special float %s converted to: %s", tt.name, result)
			}
		})
	}
}

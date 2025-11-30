package evaluator

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

func TestConcatStrings_TwoStrings(t *testing.T) {
	e := createTestEvaluator()

	args := []Value{
		&runtime.StringValue{Value: "Hello"},
		&runtime.StringValue{Value: "World"},
	}

	result := e.ConcatStrings(args)

	strVal, ok := result.(*runtime.StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %T", result)
	}

	expected := "HelloWorld"
	if strVal.Value != expected {
		t.Errorf("expected %q, got %q", expected, strVal.Value)
	}
}

func TestConcatStrings_MultipleStrings(t *testing.T) {
	e := createTestEvaluator()

	args := []Value{
		&runtime.StringValue{Value: "Hello"},
		&runtime.StringValue{Value: " "},
		&runtime.StringValue{Value: "World"},
		&runtime.StringValue{Value: "!"},
	}

	result := e.ConcatStrings(args)

	strVal, ok := result.(*runtime.StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %T", result)
	}

	expected := "Hello World!"
	if strVal.Value != expected {
		t.Errorf("expected %q, got %q", expected, strVal.Value)
	}
}

func TestConcatStrings_EmptyStrings(t *testing.T) {
	e := createTestEvaluator()

	args := []Value{
		&runtime.StringValue{Value: ""},
		&runtime.StringValue{Value: "test"},
		&runtime.StringValue{Value: ""},
	}

	result := e.ConcatStrings(args)

	strVal, ok := result.(*runtime.StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %T", result)
	}

	expected := "test"
	if strVal.Value != expected {
		t.Errorf("expected %q, got %q", expected, strVal.Value)
	}
}

func TestConcatStrings_NonStringArgument(t *testing.T) {
	e := createTestEvaluator()

	args := []Value{
		&runtime.StringValue{Value: "Hello"},
		&runtime.IntegerValue{Value: 42},
	}

	result := e.ConcatStrings(args)

	if result.Type() != "ERROR" {
		t.Errorf("expected ERROR, got %s", result.Type())
	}

	errVal, ok := result.(*ErrorValue)
	if !ok {
		t.Fatalf("expected ErrorValue, got %T", result)
	}

	// Error message should mention argument 2 (1-based indexing)
	if !strings.Contains(errVal.Message, "argument 2") {
		t.Errorf("error message should mention 'argument 2', got: %s", errVal.Message)
	}

	if !strings.Contains(errVal.Message, "INTEGER") {
		t.Errorf("error message should mention 'INTEGER' type, got: %s", errVal.Message)
	}
}

func TestConcatStrings_FirstArgumentNonString(t *testing.T) {
	e := createTestEvaluator()

	args := []Value{
		&runtime.BooleanValue{Value: true},
		&runtime.StringValue{Value: "test"},
	}

	result := e.ConcatStrings(args)

	if result.Type() != "ERROR" {
		t.Errorf("expected ERROR, got %s", result.Type())
	}

	errVal, ok := result.(*ErrorValue)
	if !ok {
		t.Fatalf("expected ErrorValue, got %T", result)
	}

	// Error message should mention argument 1 (1-based indexing)
	if !strings.Contains(errVal.Message, "argument 1") {
		t.Errorf("error message should mention 'argument 1', got: %s", errVal.Message)
	}
}

func TestConcatStrings_SingleString(t *testing.T) {
	e := createTestEvaluator()

	args := []Value{
		&runtime.StringValue{Value: "OnlyOne"},
	}

	result := e.ConcatStrings(args)

	strVal, ok := result.(*runtime.StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %T", result)
	}

	expected := "OnlyOne"
	if strVal.Value != expected {
		t.Errorf("expected %q, got %q", expected, strVal.Value)
	}
}

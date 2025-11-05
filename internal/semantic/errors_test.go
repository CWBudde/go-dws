package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
)

func TestNewTypeMismatch(t *testing.T) {
	pos := lexer.Position{Line: 10, Column: 5}
	varName := "count"
	expectedType := types.INTEGER
	gotType := types.STRING

	err := NewTypeMismatch(pos, varName, expectedType, gotType)

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorTypeMismatch {
		t.Errorf("Expected error type %s, got %s", ErrorTypeMismatch, err.Type)
	}

	if err.VariableName != varName {
		t.Errorf("Expected variable name %s, got %s", varName, err.VariableName)
	}

	if err.Expected == nil {
		t.Error("Expected type should not be nil")
	}

	if err.Got == nil {
		t.Error("Got type should not be nil")
	}

	if err.Expected != expectedType {
		t.Errorf("Expected type %v, got %v", expectedType, err.Expected)
	}

	if err.Got != gotType {
		t.Errorf("Got type %v, got %v", gotType, err.Got)
	}

	if !strings.Contains(err.Message, varName) {
		t.Errorf("Error message should contain variable name %s: %s", varName, err.Message)
	}
}

func TestToCompilerError(t *testing.T) {
	pos := lexer.Position{Line: 2, Column: 9}
	varName := "count"
	expectedType := types.INTEGER
	gotType := types.STRING

	err := NewTypeMismatch(pos, varName, expectedType, gotType)
	source := "var count: Integer;\ncount := \"hello\";\n"
	filename := "test.dws"

	compilerErr := err.ToCompilerError(source, filename)

	if compilerErr == nil {
		t.Fatal("Expected CompilerError to be created")
	}

	formatted := compilerErr.Error()

	// Print for debugging
	t.Logf("Formatted error:\n%s", formatted)

	// Check that the formatted error contains expected information
	if !strings.Contains(formatted, "Expected:") {
		t.Errorf("Expected formatted error to contain 'Expected:', got: %s", formatted)
	}

	if !strings.Contains(formatted, "Got:") {
		t.Errorf("Expected formatted error to contain 'Got:', got: %s", formatted)
	}

	if !strings.Contains(formatted, "Integer") {
		t.Errorf("Expected formatted error to contain 'Integer', got: %s", formatted)
	}

	if !strings.Contains(formatted, "String") {
		t.Errorf("Expected formatted error to contain 'String', got: %s", formatted)
	}
}

func TestNewUndefinedVariable(t *testing.T) {
	pos := lexer.Position{Line: 15, Column: 3}
	varName := "xyz"

	err := NewUndefinedVariable(pos, varName)

	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Type != ErrorUndefinedVariable {
		t.Errorf("Expected error type %s, got %s", ErrorUndefinedVariable, err.Type)
	}

	if err.VariableName != varName {
		t.Errorf("Expected variable name %s, got %s", varName, err.VariableName)
	}

	if !strings.Contains(err.Message, varName) {
		t.Errorf("Error message should contain variable name %s: %s", varName, err.Message)
	}
}

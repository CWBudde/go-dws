package semantic

import (
	"testing"
)

// TestClassVariableAccessViaClassName tests accessing class variables through class name
func TestClassVariableAccessViaClassName(t *testing.T) {
	input := `
type TBase = class
  class var Test : Integer;
end;

var x : Integer;
x := TBase.Test;
`
	_, err := analyzeSource(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestClassVariableAccessViaInstance tests accessing class variables through instance
func TestClassVariableAccessViaInstance(t *testing.T) {
	input := `
type TBase = class
  class var Test : Integer;
end;

var b : TBase;
var x : Integer;
x := b.Test;
`
	_, err := analyzeSource(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestClassVariableInheritance tests that child classes inherit parent class variables
func TestClassVariableInheritance(t *testing.T) {
	input := `
type TBase = class
  class var Test : Integer;
end;

type TChild = class(TBase)
end;

var x : Integer;
x := TChild.Test;  // Should access inherited class variable
`
	_, err := analyzeSource(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestClassVariableShadowing tests that child can shadow parent class variables
func TestClassVariableShadowing(t *testing.T) {
	input := `
type TBase = class
  class var Test : Integer;
end;

type TChild = class(TBase)
  class var Test : String;  // Shadows parent class variable
end;

var x : String;
x := TChild.Test;  // Should access child's class variable (String type)
`
	_, err := analyzeSource(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestClassVariableAssignment tests assignment to class variables
func TestClassVariableAssignment(t *testing.T) {
	input := `
type TBase = class
  class var Test : Integer;
end;

TBase.Test := 123;
`
	_, err := analyzeSource(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestClassVariableWithInitialization tests class variables with initialization
func TestClassVariableWithInitialization(t *testing.T) {
	input := `
type TBase = class
  class var Test : Integer := 42;
end;

var x : Integer;
x := TBase.Test;
`
	_, err := analyzeSource(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestClassVariableUndefinedMember tests error when accessing non-existent class variable
func TestClassVariableUndefinedMember(t *testing.T) {
	input := `
type TBase = class
  class var Test : Integer;
end;

var x : Integer;
x := TBase.NonExistent;  // Should error
`
	analyzer, err := analyzeSource(t, input)
	if err == nil {
		t.Fatal("expected error for undefined member, got none")
	}

	foundError := false
	for _, errMsg := range analyzer.Errors() {
		if contains(errMsg, "has no member") {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Errorf("expected 'has no member' error, got %v", analyzer.Errors())
	}
}

// TestClassVariableTypeCompatibility tests type checking for class variable assignments
func TestClassVariableTypeCompatibility(t *testing.T) {
	input := `
type TBase = class
  class var Test : Integer;
end;

TBase.Test := 'string';  // Should error - type mismatch
`
	_, err := analyzeSource(t, input)
	if err == nil {
		t.Fatal("expected type mismatch error, got none")
	}
}

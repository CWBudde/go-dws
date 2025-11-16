package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestPrivateClassVariableAccessBlocked verifies that private class variables
// cannot be accessed from outside the class.
// This test was added to address PR #126 feedback about visibility enforcement.
func TestPrivateClassVariableAccessBlocked(t *testing.T) {
	code := `
type
  TBase = class
  private
    class var PrivateVar: Integer;
  public
    class var PublicVar: Integer;
  end;

var x: Integer;
begin
  x := TBase.PublicVar;   // Should work
  x := TBase.PrivateVar;  // Should fail - private
end.
`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.Analyze(program)

	// We expect exactly one error about accessing private class variable
	if len(analyzer.Errors()) == 0 {
		t.Fatal("Expected semantic error for accessing private class variable, but got none")
	}

	// Check that we got the expected error
	foundPrivateError := false
	for _, err := range analyzer.Errors() {
		if strings.Contains(err, "cannot access private class variable") &&
			(strings.Contains(err, "PrivateVar") || strings.Contains(err, "privatevar")) {
			foundPrivateError = true
			break
		}
	}

	if !foundPrivateError {
		t.Errorf("Expected error about private class variable access, got: %v", analyzer.Errors())
	}
}

// TestProtectedClassVariableAccessFromChild verifies that protected class variables
// can be accessed from derived classes.
func TestProtectedClassVariableAccessFromChild(t *testing.T) {
	code := `
type
  TBase = class
  protected
    class var ProtectedVar: Integer;
  end;

  TDerived = class(TBase)
  public
    class function GetProtectedVar: Integer;
  end;

class function TDerived.GetProtectedVar: Integer;
begin
  Result := TBase.ProtectedVar;  // Should work - accessing from derived class
end;
`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.Analyze(program)

	// Should not have errors about accessing protected class variable from derived class
	for _, err := range analyzer.Errors() {
		if strings.Contains(err, "cannot access") && strings.Contains(err, "ProtectedVar") {
			t.Errorf("Unexpected error accessing protected class variable from derived class: %s", err)
		}
	}
}

// TestPublicClassVariableAccessFromAnywhere verifies that public class variables
// can be accessed from anywhere.
func TestPublicClassVariableAccessFromAnywhere(t *testing.T) {
	code := `
type
  TBase = class
  public
    class var PublicVar: Integer;
  end;

var x: Integer;
begin
  x := TBase.PublicVar;  // Should work
end.
`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.Analyze(program)

	// Should not have any errors
	if len(analyzer.Errors()) > 0 {
		t.Errorf("Unexpected errors: %v", analyzer.Errors())
	}
}

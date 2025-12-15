package semantic

import (
	"strings"
	"testing"
)

// Test helpers to reduce cyclomatic complexity

// expectAnalysisError runs analysis expecting an error containing all specified keywords
func expectAnalysisError(t *testing.T, input string, keywords ...string) {
	t.Helper()
	analyzer, err := analyzeSource(t, input)
	if err == nil {
		t.Errorf("Expected error containing keywords %v, but got no error", keywords)
		return
	}

	if !hasErrorWithKeywords(analyzer.Errors(), keywords...) {
		t.Errorf("Expected error containing keywords %v, got: %v", keywords, analyzer.Errors())
	}
}

// expectNoError runs analysis expecting success
func expectNoError(t *testing.T, input string) {
	t.Helper()
	_, err := analyzeSource(t, input)
	if err != nil {
		t.Errorf("Expected no errors, got: %v", err)
	}
}

// hasErrorWithKeywords checks if any error message contains all specified keywords
func hasErrorWithKeywords(errors []string, keywords ...string) bool {
	for _, errMsg := range errors {
		if containsAllKeywords(errMsg, keywords...) {
			return true
		}
	}
	return false
}

// containsAllKeywords checks if a string contains all specified keywords (case-insensitive)
func containsAllKeywords(s string, keywords ...string) bool {
	lowerMsg := strings.ToLower(s)
	for _, keyword := range keywords {
		if !strings.Contains(lowerMsg, strings.ToLower(keyword)) {
			return false
		}
	}
	return true
}

// ============================================================================
// Inherited Expression Semantic Tests
// ============================================================================

func TestInheritedExpression(t *testing.T) {
	t.Run("inherited in method with parent class", func(t *testing.T) {
		input := `
type TBase = class
	function GetValue: Integer; virtual;
end;

type TChild = class(TBase)
	function GetValue: Integer; override;
end;

function TBase.GetValue: Integer;
begin
	result := 10;
end;

function TChild.GetValue: Integer;
begin
	result := inherited GetValue() + 5;
end;
`
		expectNoError(t, input)
	})

	t.Run("bare inherited in method", func(t *testing.T) {
		input := `
type TBase = class
	procedure DoSomething; virtual;
end;

type TChild = class(TBase)
	procedure DoSomething; override;
end;

procedure TBase.DoSomething;
begin
	PrintLn('Base');
end;

procedure TChild.DoSomething;
begin
	inherited;  // Bare inherited - calls parent's DoSomething
	PrintLn('Child');
end;
`
		expectNoError(t, input)
	})

	t.Run("inherited with arguments", func(t *testing.T) {
		input := `
type TBase = class
	function Add(a, b: Integer): Integer; virtual;
end;

type TChild = class(TBase)
	function Add(a, b: Integer): Integer; override;
end;

function TBase.Add(a, b: Integer): Integer;
begin
	result := a + b;
end;

function TChild.Add(a, b: Integer): Integer;
begin
	result := inherited Add(a, b) * 2;
end;
`
		expectNoError(t, input)
	})

	t.Run("inherited property access", func(t *testing.T) {
		input := `
type TBase = class
	FValue: Integer;
	property Value: Integer read FValue write FValue;
end;

type TChild = class(TBase)
	FChildValue: Integer;
	property Value: Integer read FChildValue write FChildValue;
	function GetParentValue: Integer;
end;

function TChild.GetParentValue: Integer;
begin
	result := inherited Value;
end;
`
		expectNoError(t, input)
	})
}

func TestInheritedExpression_Errors(t *testing.T) {
	t.Run("inherited outside class method", func(t *testing.T) {
		input := `
begin
	inherited DoSomething;
end.
`
		expectAnalysisError(t, input, "inherited", "class method")
	})

	t.Run("inherited in class with no parent", func(t *testing.T) {
		input := `
type TBase = class
	procedure DoSomething;
end;

procedure TBase.DoSomething;
begin
	inherited DoSomething;  // Error: TBase has no parent (except TObject)
end;
`
		expectAnalysisError(t, input, "parent")
	})

	t.Run("bare inherited outside method context", func(t *testing.T) {
		input := `
type TBase = class
	procedure DoSomething; virtual;
end;

type TChild = class(TBase)
	procedure DoSomething; override;
end;

procedure TBase.DoSomething;
begin
	PrintLn('Base');
end;

begin
	inherited;  // Error: bare inherited needs method context
end.
`
		expectAnalysisError(t, input, "inherited", "class method")
	})

	t.Run("inherited method not found in parent", func(t *testing.T) {
		input := `
type TBase = class
	procedure DoSomething;
end;

type TChild = class(TBase)
	procedure DoSomething; override;
	procedure OtherMethod;
end;

procedure TBase.DoSomething;
begin
	PrintLn('Base');
end;

procedure TChild.DoSomething;
begin
	inherited DoSomething;
end;

procedure TChild.OtherMethod;
begin
	inherited NonExistent;  // Error: NonExistent not in parent
end;
`
		expectAnalysisError(t, input, "not found")
	})

	t.Run("inherited with wrong number of arguments", func(t *testing.T) {
		input := `
type TBase = class
	function Add(a, b: Integer): Integer;
end;

type TChild = class(TBase)
	function Add(a, b: Integer): Integer; override;
end;

function TBase.Add(a, b: Integer): Integer;
begin
	result := a + b;
end;

function TChild.Add(a, b: Integer): Integer;
begin
	result := inherited Add(a);  // Error: wrong number of arguments
end;
`
		expectAnalysisError(t, input, "expected")
	})

	t.Run("inherited with wrong argument types", func(t *testing.T) {
		input := `
type TBase = class
	function Process(value: Integer): Integer;
end;

type TChild = class(TBase)
	function Process(value: Integer): Integer; override;
end;

function TBase.Process(value: Integer): Integer;
begin
	result := value * 2;
end;

function TChild.Process(value: Integer): Integer;
begin
	result := inherited Process('string');  // Error: wrong type
end;
`
		expectAnalysisError(t, input, "type")
	})

	t.Run("cannot call property as method", func(t *testing.T) {
		input := `
type TBase = class
	FValue: Integer;
	property Value: Integer read FValue write FValue;
end;

type TChild = class(TBase)
	function GetValue: Integer;
end;

function TChild.GetValue: Integer;
begin
	result := inherited Value(10);  // Error: cannot call property
end;
`
		expectAnalysisError(t, input, "property", "call")
	})

	t.Run("cannot call field as method", func(t *testing.T) {
		input := `
type TBase = class
	Value: Integer;
end;

type TChild = class(TBase)
	function GetValue: Integer;
end;

function TChild.GetValue: Integer;
begin
	result := inherited Value(10);  // Error: cannot call field
end;
`
		expectAnalysisError(t, input, "field", "call")
	})
}

func TestInheritedExpression_ComplexCases(t *testing.T) {
	t.Run("inherited in constructor", func(t *testing.T) {
		input := `
type TBase = class
	constructor Create;
end;

type TChild = class(TBase)
	constructor Create;
end;

constructor TBase.Create;
begin
	PrintLn('Base Create');
end;

constructor TChild.Create;
begin
	inherited Create;
	PrintLn('Child Create');
end;
`
		expectNoError(t, input)
	})

	t.Run("inherited field access", func(t *testing.T) {
		input := `
type TBase = class
	Value: Integer;
end;

type TChild = class(TBase)
	Value: Integer;  // Shadows parent field
	function GetParentValue: Integer;
end;

function TChild.GetParentValue: Integer;
begin
	result := inherited Value;
end;
`
		expectNoError(t, input)
	})

	t.Run("multiple inheritance levels", func(t *testing.T) {
		input := `
type TBase = class
	function GetValue: Integer; virtual;
end;

type TMiddle = class(TBase)
	function GetValue: Integer; override;
end;

type TChild = class(TMiddle)
	function GetValue: Integer; override;
end;

function TBase.GetValue: Integer;
begin
	result := 1;
end;

function TMiddle.GetValue: Integer;
begin
	result := inherited GetValue() + 1;
end;

function TChild.GetValue: Integer;
begin
	result := inherited GetValue() + 1;  // Calls TMiddle.GetValue
end;
`
		expectNoError(t, input)
	})
}

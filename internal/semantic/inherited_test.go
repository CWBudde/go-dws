package semantic

import (
	"strings"
	"testing"
)

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
		_, err := analyzeSource(t, input)
		if err != nil {
			t.Errorf("Expected no errors for valid inherited usage, got: %v", err)
		}
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
		_, err := analyzeSource(t, input)
		if err != nil {
			t.Errorf("Expected no errors for bare inherited, got: %v", err)
		}
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
		_, err := analyzeSource(t, input)
		if err != nil {
			t.Errorf("Expected no errors for inherited with arguments, got: %v", err)
		}
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
		_, err := analyzeSource(t, input)
		if err != nil {
			t.Errorf("Expected no errors for inherited property access, got: %v", err)
		}
	})
}

func TestInheritedExpression_Errors(t *testing.T) {
	t.Run("inherited outside class method", func(t *testing.T) {
		input := `
begin
	inherited DoSomething;
end.
`
		analyzer, err := analyzeSource(t, input)
		if err == nil {
			t.Error("Expected error for inherited outside class method")
			return
		}

		errors := analyzer.Errors()
		found := false
		for _, errMsg := range errors {
			if strings.Contains(errMsg, "inherited") && strings.Contains(errMsg, "class method") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error about inherited outside class method, got: %v", errors)
		}
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
		analyzer, err := analyzeSource(t, input)
		if err == nil {
			t.Error("Expected error for inherited in class with no parent")
			return
		}

		errors := analyzer.Errors()
		found := false
		for _, errMsg := range errors {
			if strings.Contains(errMsg, "no parent") || strings.Contains(errMsg, "has no parent class") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error about no parent class, got: %v", errors)
		}
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
		analyzer, err := analyzeSource(t, input)
		if err == nil {
			t.Error("Expected error for bare inherited outside method")
			return
		}

		errors := analyzer.Errors()
		found := false
		for _, errMsg := range errors {
			if strings.Contains(errMsg, "inherited") && strings.Contains(errMsg, "class method") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error about inherited outside class method, got: %v", errors)
		}
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
		analyzer, err := analyzeSource(t, input)
		if err == nil {
			t.Error("Expected error for inherited method not found")
			return
		}

		errors := analyzer.Errors()
		found := false
		for _, errMsg := range errors {
			if strings.Contains(errMsg, "not found") && (strings.Contains(errMsg, "NonExistent") || strings.Contains(errMsg, "parent")) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error about method not found in parent, got: %v", errors)
		}
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
		analyzer, err := analyzeSource(t, input)
		if err == nil {
			t.Error("Expected error for wrong number of arguments")
			return
		}

		errors := analyzer.Errors()
		found := false
		for _, errMsg := range errors {
			if strings.Contains(errMsg, "wrong number") || strings.Contains(errMsg, "expected") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error about wrong number of arguments, got: %v", errors)
		}
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
		analyzer, err := analyzeSource(t, input)
		if err == nil {
			t.Error("Expected error for wrong argument type")
			return
		}

		errors := analyzer.Errors()
		found := false
		for _, errMsg := range errors {
			if strings.Contains(errMsg, "type") && (strings.Contains(errMsg, "expected") || strings.Contains(errMsg, "argument")) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error about wrong argument type, got: %v", errors)
		}
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
		analyzer, err := analyzeSource(t, input)
		if err == nil {
			t.Error("Expected error for calling property as method")
			return
		}

		errors := analyzer.Errors()
		found := false
		for _, errMsg := range errors {
			if strings.Contains(errMsg, "property") && strings.Contains(errMsg, "call") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error about calling property as method, got: %v", errors)
		}
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
		analyzer, err := analyzeSource(t, input)
		if err == nil {
			t.Error("Expected error for calling field as method")
			return
		}

		errors := analyzer.Errors()
		found := false
		for _, errMsg := range errors {
			if strings.Contains(errMsg, "field") && strings.Contains(errMsg, "call") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error about calling field as method, got: %v", errors)
		}
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
		_, err := analyzeSource(t, input)
		if err != nil {
			t.Errorf("Expected no errors for inherited in constructor, got: %v", err)
		}
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
		_, err := analyzeSource(t, input)
		if err != nil {
			t.Errorf("Expected no errors for inherited field access, got: %v", err)
		}
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
		_, err := analyzeSource(t, input)
		if err != nil {
			t.Errorf("Expected no errors for multiple inheritance levels, got: %v", err)
		}
	})
}

package semantic

import (
	"testing"
)

// ============================================================================
// Constructor/Destructor Semantic Analysis Tests
// ============================================================================

// TestConstructorBasic tests basic constructor semantic analysis
func TestConstructorBasic(t *testing.T) {
	input := `
type TExample = class
public
	FValue: Integer;
	constructor Create(AValue: Integer);
end;

constructor TExample.Create(AValue: Integer);
begin
	FValue := AValue;
end;

var obj: TExample;
begin
	obj := TExample.Create(42);
end;
`
	expectNoErrors(t, input)
}

// TestConstructorParameterValidation tests constructor parameter type checking
func TestConstructorParameterValidation(t *testing.T) {
	input := `
type TExample = class
	constructor Create(AValue: Integer);
end;

constructor TExample.Create(AValue: Integer);
begin
end;

var obj: TExample;
begin
	obj := TExample.Create('wrong type');
end;
`
	expectError(t, input, "type mismatch")
}

// TestConstructorWrongArgumentCount tests constructor with wrong number of arguments
func TestConstructorWrongArgumentCount(t *testing.T) {
	input := `
type TExample = class
	constructor Create(AValue: Integer);
end;

constructor TExample.Create(AValue: Integer);
begin
end;

var obj: TExample;
begin
	obj := TExample.Create();
end;
`
	expectError(t, input, "wrong number of arguments")
}

// TestMultipleConstructorsSemantics tests semantic analysis with multiple constructors
func TestMultipleConstructorsSemantics(t *testing.T) {
	input := `
type TExample = class
	constructor Create;
	constructor CreateWithValue(AValue: Integer);
end;

constructor TExample.Create;
begin
end;

constructor TExample.CreateWithValue(AValue: Integer);
begin
end;

var obj1: TExample;
var obj2: TExample;
begin
	obj1 := TExample.Create();
	obj2 := TExample.CreateWithValue(42);
end;
`
	expectNoErrors(t, input)
}

// TestDestructorBasic tests basic destructor semantic analysis
func TestDestructorBasic(t *testing.T) {
	input := `
type TExample = class
	FValue: Integer;
	destructor Destroy;
end;

destructor TExample.Destroy;
begin
	FValue := 0;
end;
`
	expectNoErrors(t, input)
}

// TestConstructorInInheritance tests constructors in inheritance hierarchy
func TestConstructorInInheritance(t *testing.T) {
	input := `
type TBase = class
	constructor Create;
end;

type TDerived = class(TBase)
	constructor CreateDerived(AValue: Integer);
end;

constructor TBase.Create;
begin
end;

constructor TDerived.CreateDerived(AValue: Integer);
begin
end;

var base: TBase;
var derived: TDerived;
begin
	base := TBase.Create();
	derived := TDerived.CreateDerived(42);
end;
`
	expectNoErrors(t, input)
}

// TestPrivateConstructor tests that private constructors are accessible within the class
func TestPrivateConstructor(t *testing.T) {
	input := `
type TExample = class
private
	constructor Create;
public
	class function GetInstance: TExample;
end;

constructor TExample.Create;
begin
end;

class function TExample.GetInstance: TExample;
begin
	Result := TExample.Create();
end;
`
	expectNoErrors(t, input)
}

// TestPrivateConstructorFromOutside tests that private constructors cannot be called from outside
func TestPrivateConstructorFromOutside(t *testing.T) {
	input := `
type TExample = class
private
	constructor Create;
end;

constructor TExample.Create;
begin
end;

var obj: TExample;
begin
	obj := TExample.Create();
end;
`
	expectError(t, input, "cannot access private")
}

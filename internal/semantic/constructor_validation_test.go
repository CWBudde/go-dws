package semantic

import (
	"testing"
)

// ============================================================================
// Task 9.19: Constructor Validation Comprehensive Test Suite
// ============================================================================

// TestConstructorOverloadResolution tests constructor overload selection
func TestConstructorOverloadResolution(t *testing.T) {
	input := `
type TExample = class
	constructor Create;
	constructor Create(AValue: Integer);
	constructor Create(AName: String);
end;

constructor TExample.Create;
begin
end;

constructor TExample.Create(AValue: Integer);
begin
end;

constructor TExample.Create(AName: String);
begin
end;

var obj1, obj2, obj3: TExample;
begin
	obj1 := TExample.Create();
	obj2 := TExample.Create(42);
	obj3 := TExample.Create('test');
end;
`
	expectNoErrors(t, input)
}

// TestConstructorOverloadTypeMatching tests that overloads match by type
func TestConstructorOverloadTypeMatching(t *testing.T) {
	input := `
type TExample = class
	constructor Create(AValue: Integer);
	constructor Create(AValue: Float);
end;

constructor TExample.Create(AValue: Integer);
begin
end;

constructor TExample.Create(AValue: Float);
begin
end;

var obj1, obj2: TExample;
begin
	obj1 := TExample.Create(42);
	obj2 := TExample.Create(3.14);
end;
`
	expectNoErrors(t, input)
}

// TestConstructorVisibilityPublic tests public constructor access
func TestConstructorVisibilityPublic(t *testing.T) {
	input := `
type TExample = class
public
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
	expectNoErrors(t, input)
}

// TestConstructorVisibilityPrivateFromOutside tests private constructor cannot be called from outside
func TestConstructorVisibilityPrivateFromOutside(t *testing.T) {
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
	expectError(t, input, "cannot access private constructor")
}

// TestConstructorVisibilityPrivateFromInside tests private constructor can be called from inside class
func TestConstructorVisibilityPrivateFromInside(t *testing.T) {
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

var obj: TExample;
begin
	obj := TExample.GetInstance();
end;
`
	expectNoErrors(t, input)
}

// TestConstructorVisibilityProtected tests protected constructor access
func TestConstructorVisibilityProtected(t *testing.T) {
	input := `
type TBase = class
protected
	constructor Create;
end;

type TDerived = class(TBase)
public
	class function MakeInstance: TDerived;
end;

constructor TBase.Create;
begin
end;

class function TDerived.MakeInstance: TDerived;
begin
	Result := TDerived.Create();
end;
`
	expectNoErrors(t, input)
}

// TestConstructorWrongType tests type mismatch in constructor arguments
func TestConstructorWrongType(t *testing.T) {
	input := `
type TExample = class
	constructor Create(AValue: Integer);
end;

constructor TExample.Create(AValue: Integer);
begin
end;

var obj: TExample;
begin
	obj := TExample.Create('not an integer');
end;
`
	expectError(t, input, "has type String, expected Integer")
}

// TestConstructorWrongCount tests wrong number of arguments
func TestConstructorWrongCount(t *testing.T) {
	input := `
type TExample = class
	constructor Create(AValue: Integer; AName: String);
end;

constructor TExample.Create(AValue: Integer; AName: String);
begin
end;

var obj: TExample;
begin
	obj := TExample.Create(42);
end;
`
	expectError(t, input, "expects 2 arguments, got 1")
}

// TestConstructorCaseInsensitiveClassName tests case-insensitive class name
func TestConstructorCaseInsensitiveClassName(t *testing.T) {
	input := `
type TExample = class
	constructor Create;
end;

constructor TExample.Create;
begin
end;

var obj: TExample;
begin
	obj := texample.Create();
	obj := TEXAMPLE.Create();
	obj := TExample.Create();
end;
`
	expectNoErrors(t, input)
}

// TestConstructorCaseInsensitiveConstructorName tests case-insensitive constructor name
func TestConstructorCaseInsensitiveConstructorName(t *testing.T) {
	input := `
type TExample = class
	constructor Create;
end;

constructor TExample.Create;
begin
end;

var obj: TExample;
begin
	obj := TExample.create();
	obj := TExample.CREATE();
	obj := TExample.Create();
end;
`
	expectNoErrors(t, input)
}

// TestConstructorImplicitParameterless tests implicit parameterless constructor behavior
func TestConstructorImplicitParameterless(t *testing.T) {
	input := `
type TExample = class
end;

var obj: TExample;
begin
	obj := TExample.Create();
end;
`
	expectNoErrors(t, input)
}

// TestConstructorExplicitReturnTypeFails tests that constructors cannot have explicit return types
func TestConstructorExplicitReturnTypeFails(t *testing.T) {
	input := `
type TExample = class
	constructor Create: TExample;
end;

constructor TExample.Create: TExample;
begin
end;
`
	expectError(t, input, "constructor 'Create' cannot have an explicit return type")
}

// TestConstructorMultipleOverloadsWithWrongType tests overload resolution with type mismatch
func TestConstructorMultipleOverloadsWithWrongType(t *testing.T) {
	input := `
type TExample = class
	constructor Create(AValue: Integer);
	constructor Create(AName: String);
end;

constructor TExample.Create(AValue: Integer);
begin
end;

constructor TExample.Create(AName: String);
begin
end;

var obj: TExample;
begin
	obj := TExample.Create(3.14);
end;
`
	expectError(t, input, "no constructor")
}

// TestConstructorInheritanceCallParent tests calling parent constructor
func TestConstructorInheritanceCallParent(t *testing.T) {
	input := `
type TBase = class
public
	constructor Create(AValue: Integer);
end;

type TDerived = class(TBase)
public
	constructor Create(AValue: Integer; AName: String);
end;

constructor TBase.Create(AValue: Integer);
begin
end;

constructor TDerived.Create(AValue: Integer; AName: String);
begin
end;

var base: TBase;
var derived: TDerived;
begin
	base := TBase.Create(42);
	derived := TDerived.Create(42, 'test');
end;
`
	expectNoErrors(t, input)
}

// TestConstructorNewSyntax tests 'new TClass(args)' syntax
func TestConstructorNewSyntax(t *testing.T) {
	input := `
type TExample = class
public
	constructor Create(AValue: Integer);
end;

constructor TExample.Create(AValue: Integer);
begin
end;

var obj: TExample;
begin
	obj := new TExample(42);
end;
`
	expectNoErrors(t, input)
}

// TestConstructorNewSyntaxWrongType tests type validation with 'new' syntax
func TestConstructorNewSyntaxWrongType(t *testing.T) {
	input := `
type TExample = class
public
	constructor Create(AValue: Integer);
end;

constructor TExample.Create(AValue: Integer);
begin
end;

var obj: TExample;
begin
	obj := new TExample('wrong');
end;
`
	expectError(t, input, "has type String, expected Integer")
}

// TestConstructorAbstractClass tests that abstract classes cannot be instantiated
func TestConstructorAbstractClass(t *testing.T) {
	input := `
type TAbstract = class abstract
public
	constructor Create;
end;

constructor TAbstract.Create;
begin
end;

var obj: TAbstract;
begin
	obj := TAbstract.Create();
end;
`
	expectError(t, input, "cannot instantiate abstract class")
}

// TestConstructorOverloadWithImplicitConversion tests implicit Integer to Float conversion
func TestConstructorOverloadWithImplicitConversion(t *testing.T) {
	input := `
type TExample = class
public
	constructor Create(AValue: Float);
end;

constructor TExample.Create(AValue: Float);
begin
end;

var obj: TExample;
begin
	obj := TExample.Create(42);
end;
`
	expectNoErrors(t, input)
}

// TestConstructorNoArgumentsWithParameterized tests error when calling with no args but constructor needs args
func TestConstructorNoArgumentsWithParameterized(t *testing.T) {
	input := `
type TExample = class
public
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
	expectError(t, input, "expects 1 arguments, got 0")
}

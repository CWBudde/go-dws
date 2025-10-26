package semantic

import (
	"testing"
)

// ============================================================================
// Visibility Tests (Task 7.63m-n)
// ============================================================================

// TestPrivateFieldAccessFromSameClass tests that private fields can be accessed
// from within the same class (Task 7.63g, 7.63l)
func TestPrivateFieldAccessFromSameClass(t *testing.T) {
	input := `
type TExample = class
private
	FValue: Integer;
public
	function GetValue: Integer;
end;

function TExample.GetValue: Integer;
begin
	Result := FValue; // Should work - accessing private field from same class
end;
`
	expectNoErrors(t, input)
}

// TestPrivateFieldAccessFromOutsideClass tests that private fields cannot be
// accessed from outside the class (Task 7.63g)
func TestPrivateFieldAccessFromOutsideClass(t *testing.T) {
	input := `
type TExample = class
private
	FValue: Integer;
public
	constructor Create(AValue: Integer);
end;

constructor TExample.Create(AValue: Integer);
begin
	FValue := AValue;
end;

var obj: TExample;
begin
	obj := TExample.Create(42);
	PrintLn(obj.FValue); // Should error - cannot access private field from outside
end;
`
	expectError(t, input, "cannot access private field 'FValue'")
}

// TestProtectedFieldAccessFromChild tests that protected fields can be accessed
// from derived classes (Task 7.63h)
func TestProtectedFieldAccessFromChild(t *testing.T) {
	input := `
type TBase = class
protected
	FValue: Integer;
end;

type TDerived = class(TBase)
public
	function GetValue: Integer;
end;

function TDerived.GetValue: Integer;
begin
	Result := FValue; // Should work - accessing protected field from child class
end;
`
	expectNoErrors(t, input)
}

// TestProtectedFieldAccessFromOutside tests that protected fields cannot be
// accessed from unrelated code (Task 7.63h)
func TestProtectedFieldAccessFromOutside(t *testing.T) {
	input := `
type TExample = class
protected
	FValue: Integer;
public
	constructor Create(AValue: Integer);
end;

constructor TExample.Create(AValue: Integer);
begin
	FValue := AValue;
end;

var obj: TExample;
begin
	obj := TExample.Create(42);
	PrintLn(obj.FValue); // Should error - cannot access protected field from outside
end;
`
	expectError(t, input, "cannot access protected field 'FValue'")
}

// TestPublicFieldAccessFromAnywhere tests that public fields can be accessed
// from anywhere (Task 7.63i)
func TestPublicFieldAccessFromAnywhere(t *testing.T) {
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
	PrintLn(obj.FValue); // Should work - accessing public field from outside
end;
`
	expectNoErrors(t, input)
}

// TestPrivateMethodAccessFromSameClass tests that private methods can be called
// from within the same class (Task 7.63k, 7.63l)
func TestPrivateMethodAccessFromSameClass(t *testing.T) {
	input := `
type TExample = class
private
	function Helper: Integer;
public
	function GetValue: Integer;
end;

function TExample.Helper: Integer;
begin
	Result := 42;
end;

function TExample.GetValue: Integer;
begin
	Result := Helper(); // Should work - calling private method from same class
end;
`
	expectNoErrors(t, input)
}

// TestPrivateMethodAccessFromOutside tests that private methods cannot be called
// from outside the class (Task 7.63k)
func TestPrivateMethodAccessFromOutside(t *testing.T) {
	input := `
type TExample = class
private
	function Helper: Integer;
public
	constructor Create;
end;

function TExample.Helper: Integer;
begin
	Result := 42;
end;

constructor TExample.Create;
begin
end;

var obj: TExample;
begin
	obj := TExample.Create();
	PrintLn(obj.Helper()); // Should error - cannot call private method from outside
end;
`
	expectError(t, input, "cannot call private method 'Helper'")
}

// TestProtectedMethodAccessFromChild tests that protected methods can be called
// from derived classes (Task 7.63k)
func TestProtectedMethodAccessFromChild(t *testing.T) {
	input := `
type TBase = class
protected
	function Helper: Integer;
end;

function TBase.Helper: Integer;
begin
	Result := 42;
end;

type TDerived = class(TBase)
public
	function GetValue: Integer;
end;

function TDerived.GetValue: Integer;
begin
	Result := Helper(); // Should work - calling protected method from child class
end;
`
	expectNoErrors(t, input)
}

// TestProtectedMethodAccessFromOutside tests that protected methods cannot be
// called from unrelated code (Task 7.63k)
func TestProtectedMethodAccessFromOutside(t *testing.T) {
	input := `
type TExample = class
protected
	function Helper: Integer;
public
	constructor Create;
end;

function TExample.Helper: Integer;
begin
	Result := 42;
end;

constructor TExample.Create;
begin
end;

var obj: TExample;
begin
	obj := TExample.Create();
	PrintLn(obj.Helper()); // Should error - cannot call protected method from outside
end;
`
	expectError(t, input, "cannot call protected method 'Helper'")
}

// TestPublicMethodAccessFromAnywhere tests that public methods can be called
// from anywhere (Task 7.63k)
func TestPublicMethodAccessFromAnywhere(t *testing.T) {
	input := `
type TExample = class
public
	function GetValue: Integer;
	constructor Create;
end;

function TExample.GetValue: Integer;
begin
	Result := 42;
end;

constructor TExample.Create;
begin
end;

var obj: TExample;
begin
	obj := TExample.Create();
	PrintLn(obj.GetValue()); // Should work - calling public method from outside
end;
`
	expectNoErrors(t, input)
}

// TestMixedVisibility tests a class with mixed visibility levels (Task 7.63n)
func TestMixedVisibility(t *testing.T) {
	input := `
type TExample = class
private
	FPrivate: Integer;
protected
	FProtected: String;
public
	FPublic: Float;
	constructor Create;
end;

constructor TExample.Create;
begin
	FPrivate := 1;    // OK - same class
	FProtected := 'a'; // OK - same class
	FPublic := 3.14;   // OK - same class
end;

var obj: TExample;
begin
	obj := TExample.Create();
	PrintLn(obj.FPublic);  // Should work - public field
end;
`
	expectNoErrors(t, input)
}

// TestInheritedVisibility tests visibility across inheritance hierarchy (Task 7.63n)
func TestInheritedVisibility(t *testing.T) {
	input := `
type TBase = class
private
	FPrivateBase: Integer;
protected
	FProtectedBase: String;
public
	FPublicBase: Float;
end;

type TDerived = class(TBase)
public
	function TestAccess: String;
end;

function TDerived.TestAccess: String;
begin
	FPublicBase := 1.0;      // OK - public in base
	FProtectedBase := 'test'; // OK - protected in base, accessible from child
	Result := FProtectedBase;
end;
`
	expectNoErrors(t, input)
}

// TestPrivateFieldNotInheritedAccess tests that child cannot access parent's
// private fields (Task 7.63n)
func TestPrivateFieldNotInheritedAccess(t *testing.T) {
	input := `
type TBase = class
private
	FPrivate: Integer;
end;

type TDerived = class(TBase)
public
	function TestAccess: Integer;
end;

function TDerived.TestAccess: Integer;
begin
	Result := FPrivate; // Should error - cannot access private field from parent
end;
`
	expectError(t, input, "cannot access private field 'FPrivate'")
}

// TestDefaultVisibilityIsPublic tests that members without explicit visibility
// are public by default (Task 7.63e)
func TestDefaultVisibilityIsPublic(t *testing.T) {
	input := `
type TExample = class
	FValue: Integer;  // No visibility keyword - should be public by default
	constructor Create(AValue: Integer);
end;

constructor TExample.Create(AValue: Integer);
begin
	FValue := AValue;
end;

var obj: TExample;
begin
	obj := TExample.Create(42);
	PrintLn(obj.FValue); // Should work - default is public
end;
`
	expectNoErrors(t, input)
}

// TestMultipleVisibilitySections tests a class with multiple visibility sections
func TestMultipleVisibilitySections(t *testing.T) {
	input := `
type TExample = class
private
	FPrivate1: Integer;
	FPrivate2: Integer;
protected
	FProtected1: String;
public
	FPublic1: Float;
private
	FPrivate3: Integer;  // Back to private section
public
	FPublic2: Boolean;
end;
`
	// Just test that it parses correctly without semantic errors
	expectNoErrors(t, input)
}

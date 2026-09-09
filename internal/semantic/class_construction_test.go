package semantic

import (
	"strings"
	"testing"
)

// TestClassConstruction_InheritanceResolvedBeforeMembers covers L-S1a: parents are
// linked and class-level shape flags are known before member/body checking, so
// class analysis no longer depends on declaration order.
func TestClassConstruction_InheritanceResolvedBeforeMembers(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		errorContains string
		wantErr       bool
	}{
		{
			name: "method body of a child declared before its parent sees inherited fields",
			input: `type TChild = class(TParent)
	procedure Show;
end;
type TParent = class
	F: Integer;
end;
procedure TChild.Show;
begin
	PrintLn(F);
end;
begin
end;`,
			wantErr: false,
		},
		{
			name: "child declared before parent inherits parent field",
			input: `type TChild = class(TParent) end;
type TParent = class
	F: Integer;
end;
var c: TChild;
begin
	c := TChild.Create;
	c.F := 7;
end;`,
			wantErr: false,
		},
		{
			name: "child declared before parent inherits parent method",
			input: `type TChild = class(TParent) end;
type TParent = class
	procedure Hello;
	begin
	end;
end;
var c: TChild;
begin
	c := TChild.Create;
	c.Hello;
end;`,
			wantErr: false,
		},
		{
			name: "grandchild declared before its ancestors",
			input: `type TC = class(TB) end;
type TB = class(TA) end;
type TA = class
	F: Integer;
end;
var x: TC;
begin
	x := TC.Create;
	x.F := 1;
end;`,
			wantErr: false,
		},
		{
			name:          "unknown parent reports a diagnostic",
			input:         `type TChild = class(TNope) end;` + "\nbegin\nend;",
			wantErr:       true,
			errorContains: "parent class 'TNope' not found",
		},
		{
			name: "two class inheritance cycle reports a diagnostic",
			input: `type TA = class(TB) end;
type TB = class(TA) end;
begin
end;`,
			wantErr:       true,
			errorContains: "circular inheritance",
		},
		{
			name: "three class inheritance cycle reports a diagnostic",
			input: `type TA = class(TB) end;
type TB = class(TC) end;
type TC = class(TA) end;
begin
end;`,
			wantErr:       true,
			errorContains: "circular inheritance",
		},
		{
			name: "self inheritance reports a diagnostic",
			input: `type TA = class(TA) end;
begin
end;`,
			wantErr:       true,
			errorContains: "circular inheritance",
		},
		{
			name: "external parent declared after the child is still rejected",
			input: `type TChild = class(TExt) end;
type TExt = class external end;
begin
end;`,
			wantErr:       true,
			errorContains: "cannot inherit from external class",
		},
		{
			name: "external parent declared before the child is rejected",
			input: `type TExt = class external end;
type TChild = class(TExt) end;
begin
end;`,
			wantErr:       true,
			errorContains: "cannot inherit from external class",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := analyzeSource(t, tt.input)

			if tt.wantErr && err == nil {
				t.Fatalf("expected an error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errorContains) {
				t.Fatalf("expected error containing %q, got: %v", tt.errorContains, err)
			}
		})
	}
}

// TestClassConstruction_PreservesForwardAndPartial guards the behavior that the
// inheritance phase must not disturb: explicit `forward` declarations and
// partial classes.
func TestClassConstruction_PreservesForwardAndPartial(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		errorContains string
		wantErr       bool
	}{
		{
			name: "explicit forward declaration resolved by later implementation",
			input: `type TFoo = class;
type TBar = class
	Foo: TFoo;
end;
type TFoo = class
	Bar: TBar;
end;
begin
end;`,
			wantErr: false,
		},
		{
			// `class(TParent);` is a complete empty subclass, not a forward
			// declaration (see parser.parseClassParentAndInterfaces), so a second
			// declaration of the same name is a redeclaration.
			name: "empty subclass shorthand is not a forward declaration",
			input: `type TA = class end;
type TFoo = class(TA);
type TFoo = class(TA) end;
begin
end;`,
			wantErr:       true,
			errorContains: "already defined",
		},
		{
			name: "empty subclass shorthand may name a parent declared later",
			input: `type TFoo = class(TBase);
type TBase = class
	F: Integer;
end;
var f: TFoo;
begin
	f := TFoo.Create;
	f.F := 1;
end;`,
			wantErr: false,
		},
		{
			name: "unimplemented forward declaration is reported",
			input: `type TFoo = class;
begin
end;`,
			wantErr:       true,
			errorContains: "TFoo",
		},
		{
			name: "partial class parts are merged",
			input: `type TPart = partial class
	A: Integer;
end;
type TPart = partial class
	B: Integer;
end;
var p: TPart;
begin
	p := TPart.Create;
	p.A := 1;
	p.B := 2;
end;`,
			wantErr: false,
		},
		{
			name: "partial class parts with conflicting parents are rejected",
			input: `type TA = class end;
type TB = class end;
type TPart = partial class(TA)
	X: Integer;
end;
type TPart = partial class(TB)
	Y: Integer;
end;
begin
end;`,
			wantErr:       true,
			errorContains: "conflicting parent",
		},
		{
			name: "partial class declared before its parent",
			input: `type TPart = partial class(TBase)
	A: Integer;
end;
type TPart = partial class(TBase)
	B: Integer;
end;
type TBase = class
	F: Integer;
end;
var p: TPart;
begin
	p := TPart.Create;
	p.F := 1;
	p.A := 2;
	p.B := 3;
end;`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := analyzeSource(t, tt.input)

			if tt.wantErr && err == nil {
				t.Fatalf("expected an error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errorContains) {
				t.Fatalf("expected error containing %q, got: %v", tt.errorContains, err)
			}
		})
	}
}

// TestClassConstruction_MemberSignaturesBeforeBodies covers L-S1b: every class
// member signature (field, class var, constant, property, method) is registered
// before any inline class method body is checked, so a body may name a class or
// a member declared later in the file.
func TestClassConstruction_MemberSignaturesBeforeBodies(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		errorContains string
		wantErr       bool
	}{
		{
			name: "field typed by a later-declared class",
			input: `type TA = class
	B: TB;
end;
type TB = class
	N: Integer;
end;
var a := TA.Create;
a.B := TB.Create;
a.B.N := 7;`,
		},
		{
			name: "mutually referring class fields without forward",
			input: `type TA = class
	Other: TB;
	N: Integer;
end;
type TB = class
	Back: TA;
	M: Integer;
end;
var a := TA.Create;
var b := TB.Create;
a.Other := b;
b.Back := a;`,
		},
		{
			name: "property and method signatures typed by a later-declared class",
			input: `type TA = class
private
	FP: TB;
public
	property P: TB read FP write FP;
	function Make(x: TB): TB;
	begin
		Result := x;
	end;
end;
type TB = class
	V: Integer;
end;
var a := TA.Create;
var b := TB.Create;
a.P := b;
PrintLn(a.Make(b).V);`,
		},
		{
			name: "inline body calls a method of a later-declared class",
			input: `type TA = class
	B: TB;
	procedure Go;
	begin
		B.Hello;
		PrintLn(B.V);
	end;
end;
type TB = class
	V: Integer;
	procedure Hello;
	begin
		PrintLn('hi');
	end;
end;`,
		},
		{
			name: "inline body constructs and uses a later-declared class",
			input: `type TA = class
	function Make: TB;
	begin
		Result := TB.Create;
		Result.V := 9;
	end;
end;
type TB = class
	V: Integer;
end;`,
		},
		{
			name: "inline body calls a method declared later in the same class",
			input: `type TA = class
	procedure A;
	begin
		B;
	end;
	procedure B;
	begin
		PrintLn('b');
	end;
end;`,
		},
		{
			name: "inline body reads a property declared later in the same class",
			input: `type TA = class
private
	FV: Integer;
	procedure Show;
	begin
		PrintLn(V);
	end;
public
	property V: Integer read FV write FV;
	procedure Run;
	begin
		Show;
	end;
end;`,
		},
		{
			name: "inline body reads a constant declared later in the same class",
			input: `type TA = class
	function Greet: String;
	begin
		Result := cHello;
	end;
	const cHello = 'Hello';
end;`,
		},
		{
			name: "identity: value reached through a forward-referencing field is the later-declared class",
			input: `type TA = class
	B: TB;
end;
type TB = class
	procedure Hello;
	begin
		PrintLn('hello');
	end;
end;
var a := TA.Create;
var b: TB := TB.Create;
a.B := b;
b := a.B;
a.B.Hello;
b.Hello;`,
		},
		{
			name: "a body still reports genuinely unknown members of a later-declared class",
			input: `type TA = class
	B: TB;
	procedure Go;
	begin
		B.Missing;
	end;
end;
type TB = class
	V: Integer;
end;`,
			wantErr:       true,
			errorContains: "Missing",
		},
		{
			name: "a global declared after the type section does not shadow a class constant",
			input: `type TChild = class
	const B = 4.5;
	procedure P;
	begin
		PrintLn(B);
	end;
end;
var b := TChild.Create;
b.P;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := analyzeSource(t, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error but got none")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Fatalf("expected error containing %q, got: %v", tt.errorContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestClassConstruction_ValidationAfterSignaturesComplete covers L-S1c: the
// ancestor-dependent validations (override, hiding, interface implementation,
// abstract) run only once every ancestor's member surface is registered, so an
// `override` against a parent declared later in the file is accepted — while the
// genuinely invalid cases stay diagnosed with the same messages.
func TestClassConstruction_ValidationAfterSignaturesComplete(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		errorContains string
		wantErr       bool
	}{
		{
			name: "override against a parent declared later is accepted",
			input: `type TC = class(TA)
	procedure Go; override;
	begin
		PrintLn(2);
	end;
end;
type TA = class
	procedure Go; virtual;
	begin
		PrintLn(1);
	end;
end;
var c := TC.Create;
c.Go;`,
		},
		{
			name: "inherited resolves when the parent is declared later",
			input: `type TC = class(TA)
	procedure Go; override;
	begin
		inherited Go;
		PrintLn(2);
	end;
end;
type TA = class
	procedure Go; virtual;
	begin
		PrintLn(1);
	end;
end;
var c := TC.Create;
c.Go;`,
		},
		{
			name: "override of a constructor declared in a later parent is accepted",
			input: `type TC = class(TA)
	constructor Create; override;
	begin
		inherited Create;
	end;
end;
type TA = class
	constructor Create; virtual;
	begin
	end;
end;
var c := TC.Create;`,
		},
		{
			name: "out-of-line implementation of a method overriding a later parent",
			input: `type TC = class(TA)
	procedure Go; override;
end;
type TA = class
	procedure Go; virtual;
	begin
	end;
end;
procedure TC.Go;
begin
	inherited Go;
	PrintLn(TB.Value);
end;
type TB = class
	const Value = 42;
end;
var c := TC.Create;
c.Go;`,
		},
		{
			name: "override through a grandparent declared last",
			input: `type TC = class(TB)
	procedure Go; override;
	begin
	end;
end;
type TB = class(TA)
end;
type TA = class
	procedure Go; virtual;
	begin
	end;
end;
var c := TC.Create;
c.Go;`,
		},
		{
			name: "override with no such parent method stays diagnosed (parent first)",
			input: `type TA = class
	procedure Go; virtual;
	begin
	end;
end;
type TC = class(TA)
	procedure Nope; override;
	begin
	end;
end;`,
			wantErr:       true,
			errorContains: "no such method exists in parent class",
		},
		{
			name: "override with no such parent method stays diagnosed (parent last)",
			input: `type TC = class(TA)
	procedure Nope; override;
	begin
	end;
end;
type TA = class
	procedure Go; virtual;
	begin
	end;
end;`,
			wantErr:       true,
			errorContains: "no such method exists in parent class",
		},
		{
			name: "override with a mismatched signature stays diagnosed (parent last)",
			input: `type TC = class(TA)
	procedure Go(x: String); override;
	begin
	end;
end;
type TA = class
	procedure Go(x: Integer); virtual;
	begin
	end;
end;`,
			wantErr:       true,
			errorContains: "no matching signature exists in parent class",
		},
		{
			name: "override of a non-virtual parent method stays diagnosed (parent last)",
			input: `type TC = class(TA)
	procedure Go; override;
	begin
	end;
end;
type TA = class
	procedure Go;
	begin
	end;
end;`,
			wantErr:       true,
			errorContains: "parent method is not virtual",
		},
		{
			name: "hiding a virtual parent method without override stays diagnosed (parent last)",
			input: `type TC = class(TA)
	procedure Go;
	begin
	end;
end;
type TA = class
	procedure Go; virtual;
	begin
	end;
end;`,
			wantErr:       true,
			errorContains: "hides virtual parent method",
		},
		{
			name: "duplicate member stays diagnosed",
			input: `type TC = class(TA)
	procedure Go; virtual;
	procedure Go; virtual;
end;
type TA = class
end;`,
			wantErr:       true,
			errorContains: "duplicate method signature",
		},
		{
			name: "declared but unimplemented method stays diagnosed (parent last)",
			input: `type TC = class(TA)
	procedure Go;
end;
type TA = class
end;
var c := TC.Create;`,
			wantErr:       true,
			errorContains: "not implemented",
		},
		{
			name: "instantiating a class with an unimplemented abstract method stays diagnosed",
			input: `type TA = class abstract
	procedure Go; virtual; abstract;
end;
type TC = class(TA)
end;
var c := TC.Create;`,
			wantErr:       true,
			errorContains: "abstract",
		},
		{
			name: "missing interface implementation stays diagnosed (interface declared later)",
			input: `type TC = class(TObject, IFoo)
end;
type IFoo = interface
	procedure Go;
end;`,
			wantErr:       true,
			errorContains: "IFoo",
		},
		{
			name: "executable statements keep source order around a type section",
			input: `PrintLn(1);
type TA = class
	const C = 5;
end;
PrintLn(TA.C);
PrintLn(3);`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := analyzeSource(t, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error but got none")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Fatalf("expected error containing %q, got: %v", tt.errorContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

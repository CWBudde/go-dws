package interp

import (
	"strings"
	"testing"
)

// These tests run through testEvalWithOutputAndSemantic rather than
// testEvalWithOutput: binding a call to the receiver's declared type needs the
// analyzer's type annotations, and without them dispatch necessarily falls back
// to the runtime class.
//
// TestDispatch_NonVirtualMethodBindsToDeclaredType pins Delphi's static binding
// rule: a method that is not virtual is chosen by the receiver's *declared*
// type, so redeclaring it in a descendant hides it rather than overriding it.
func TestDispatch_NonVirtualMethodBindsToDeclaredType(t *testing.T) {
	input := `
type TA = class
	procedure Q;
	class procedure P;
end;
type TB = class (TA)
	procedure Q;
	class procedure P;
end;

procedure TA.Q; begin PrintLn('TA.Q'); end;
procedure TB.Q; begin PrintLn('TB.Q'); end;
class procedure TA.P; begin PrintLn('TA.P'); end;
class procedure TB.P; begin PrintLn('TB.P'); end;

var a : TA;
begin
	a := TB.Create;
	a.Q;
	a.P;
end
`

	_, output := testEvalWithOutputAndSemantic(t, input)
	if want := "TA.Q\nTA.P\n"; output != want {
		t.Errorf("non-virtual dispatch: got %q, want %q", output, want)
	}
}

// TestDispatch_VirtualMethodStillDispatchesDynamically guards the other side of
// the rule: static binding must not swallow genuine virtual dispatch.
func TestDispatch_VirtualMethodStillDispatchesDynamically(t *testing.T) {
	input := `
type TA = class
	procedure Q; virtual;
end;
type TB = class (TA)
	procedure Q; override;
end;

procedure TA.Q; begin PrintLn('TA.Q'); end;
procedure TB.Q; begin PrintLn('TB.Q'); end;

var a : TA;
begin
	a := TB.Create;
	a.Q;
end
`

	_, output := testEvalWithOutputAndSemantic(t, input)
	if want := "TB.Q\n"; output != want {
		t.Errorf("virtual dispatch: got %q, want %q", output, want)
	}
}

// TestDispatch_ReintroduceBreaksTheVirtualChain covers the case the virtual
// method table models but name lookup does not: a reintroduced method does not
// take over its ancestor's slot, so a call typed at the ancestor keeps running
// the ancestor's body — and so does a call on a class below the reintroduction.
func TestDispatch_ReintroduceBreaksTheVirtualChain(t *testing.T) {
	input := `
type TBase = class
	procedure Func; virtual;
end;
type TChild = class (TBase)
	procedure Func; reintroduce;
end;
type TSubChild = class (TChild)
	procedure Func;
end;

procedure TBase.Func; begin PrintLn(ClassName + ' base'); end;
procedure TChild.Func; begin PrintLn(ClassName + ' child'); end;
procedure TSubChild.Func; begin PrintLn(ClassName + ' sub'); end;

var o : TBase;
begin
	o := TChild.Create;
	o.Func;
	o := TSubChild.Create;
	o.Func;
end
`

	_, output := testEvalWithOutputAndSemantic(t, input)
	if want := "TChild base\nTSubChild base\n"; output != want {
		t.Errorf("reintroduce: got %q, want %q", output, want)
	}
}

// TestDispatch_ReintroduceVirtualStartsANewChain covers `reintroduce; virtual`:
// it begins a second virtual chain that shares a signature with the one it
// hides. A call typed at the original base must stay on the original chain.
func TestDispatch_ReintroduceVirtualStartsANewChain(t *testing.T) {
	input := `
type TBase = class
	procedure Func; virtual;
end;
type TChild = class (TBase)
	procedure Func; reintroduce; virtual;
end;
type TSubChild = class (TChild)
	procedure Func; override;
end;

procedure TBase.Func; begin PrintLn(ClassName + ' base'); end;
procedure TChild.Func; begin PrintLn(ClassName + ' child'); end;
procedure TSubChild.Func; begin PrintLn(ClassName + ' sub'); end;

var o : TBase;
var c : TChild;
begin
	o := TSubChild.Create;
	o.Func;
	c := TSubChild.Create;
	c.Func;
end
`

	_, output := testEvalWithOutputAndSemantic(t, input)
	// Typed as TBase the call stays on TBase's chain; typed as TChild it
	// dispatches down the new chain to the override.
	if want := "TSubChild base\nTSubChild sub\n"; output != want {
		t.Errorf("reintroduce virtual: got %q, want %q", output, want)
	}
}

// TestDispatch_OverloadsAreLeftToTheOverloadResolver guards the guard: a real
// overload set must not be short-circuited by static binding, which cannot see
// argument types and would pick by arity alone.
func TestDispatch_OverloadsAreLeftToTheOverloadResolver(t *testing.T) {
	input := `
type TBase = class
	procedure Test; overload; virtual;
	procedure Test(a : Integer); overload; virtual;
end;
type TSub = class(TBase)
	procedure Test; overload; override;
	procedure Test(a : Integer); overload; override;
end;

procedure TBase.Test; begin PrintLn('base none'); end;
procedure TBase.Test(a : Integer); begin PrintLn('base ' + IntToStr(a)); end;
procedure TSub.Test; begin PrintLn('sub none'); end;
procedure TSub.Test(a : Integer); begin PrintLn('sub ' + IntToStr(a)); end;

var b : TBase;
begin
	b := TSub.Create;
	b.Test;
	b.Test(1);
end
`

	_, output := testEvalWithOutputAndSemantic(t, input)
	if want := "sub none\nsub 1\n"; output != want {
		t.Errorf("overloaded virtual dispatch: got %q, want %q", output, want)
	}
}

// TestDispatch_NilMetaclassReportsClassTypeIsNil pins the message difference
// between an unassigned `class of X` and an unassigned object reference.
func TestDispatch_NilMetaclassReportsClassTypeIsNil(t *testing.T) {
	input := `
type TObjClass = class of TObject;
var o : TObjClass;
begin
	try
		PrintLn(o.ClassName);
	except
		on E : Exception do
			PrintLn(E.Message);
	end;
end
`

	_, output := testEvalWithOutputAndSemantic(t, input)
	// The raised message carries the position of the member access.
	if !strings.HasPrefix(output, "ClassType is nil [line:") {
		t.Errorf("nil metaclass: got %q, want a \"ClassType is nil\" message", output)
	}
}

// TestClassAlias_UsableWhereverAClassNameIs covers a type alias for a class
// being used as a static receiver, as a metaclass operand and as a parent.
func TestClassAlias_UsableWhereverAClassNameIs(t *testing.T) {
	input := `
type TMyControl = TObject;
type TMyControlClass = class of TMyControl;
type TMySubControl = class (TMyControl);

begin
	PrintLn(TMyControl.ClassName);
	PrintLn(TMyControlClass.ClassName);
	PrintLn(TMySubControl.ClassName);
end
`

	_, output := testEvalWithOutputAndSemantic(t, input)
	if want := "TObject\nTObject\nTMySubControl\n"; output != want {
		t.Errorf("class alias: got %q, want %q", output, want)
	}
}

package interp

import (
	"bytes"
	"testing"
)

func TestLifetime_DeclarationVsAssignmentParityForMethodPointers(t *testing.T) {
	input := `
		type TIntFunc = function: Integer;

		type TBox = class
			Value: Integer;
			function GetValue: Integer;
			begin
				Result := Value;
			end;
		end;

		function CallTwice(fn: TIntFunc): Integer;
		begin
			Result := fn() + fn();
		end;

		var box := TBox.Create;
		box.Value := 21;

		var declFn: TIntFunc := @box.GetValue;
		var assignFn: TIntFunc;
		assignFn := @box.GetValue;

		PrintLn(CallTwice(declFn));
		PrintLn(CallTwice(assignFn));
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "42\n42\n"
	if out.String() != expected {
		t.Fatalf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestLifetime_InterfaceAcrossNestedScopes(t *testing.T) {
	input := `
		type
			IMyInterface = interface
				procedure A;
			end;

		type
			TMyImplementation = class(TObject, IMyInterface)
				procedure A; begin PrintLn('A'); end;
				destructor Destroy; override; begin PrintLn('Destroy'); end;
			end;

		procedure UseIntf(intf: IMyInterface);
		begin
			intf.A;
		end;

		function MakeIntf: IMyInterface;
		begin
			var local: IMyInterface;
			local := TMyImplementation.Create;
			UseIntf(local);
			Result := local;
		end;

		var ref: IMyInterface;
		ref := MakeIntf;
		UseIntf(ref);
		PrintLn('---');
		ref := nil;
		PrintLn('end');
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "A\nA\n---\nDestroy\nend\n"
	if out.String() != expected {
		t.Fatalf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

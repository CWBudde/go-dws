package dwscript

import (
	"bytes"
	"testing"
)

func TestEngine_StructuredTypes(t *testing.T) {
	cases := []struct{ name, source, output string }{
		{"nested bounds", `var a: array[-2..0] of array of Integer; a[-2] := [4, 5]; PrintLn(a[-2][1]);`, "5\n"},
		{"function result", `function Values: array of Integer; begin Result := [2, 3]; end; PrintLn(Values()[1]);`, "3\n"},
		{"inline record assignment", `var value: record X: Integer; end; value := (X: 7); PrintLn(value.X);`, "7\n"},
		{"record reference context", `type TPoint = record X: Integer; end; procedure Update(var point: TPoint); begin point := (X: 9); end; var point: TPoint; Update(point); PrintLn(point.X);`, "9\n"},
		{"record context", `type TInner = record X: Integer; end; type TOuter = record A: TInner; B: TInner; end; var value: TOuter := (A: (X: 4); B: (X: 5)); PrintLn(value.A.X); PrintLn(value.B.X);`, "4\n5\n"},
	}
	for _, test := range cases {
		for _, checked := range []bool{true, false} {
			mode := "unchecked"
			if checked {
				mode = "checked"
			}
			t.Run(test.name+"/"+mode, func(t *testing.T) {
				var output bytes.Buffer
				engine, err := New(WithTypeCheck(checked), WithOutput(&output))
				if err != nil {
					t.Fatal(err)
				}
				program, err := engine.Compile(test.source)
				if err != nil {
					t.Fatal(err)
				}
				for range 2 {
					output.Reset()
					if _, err := engine.Run(program); err != nil {
						t.Fatal(err)
					}
					if output.String() != test.output {
						t.Fatalf("output %q, want %q", output.String(), test.output)
					}
				}
			})
		}
	}
}

func TestEngine_BuiltinFunctionPointerIdentity(t *testing.T) {
	var output bytes.Buffer
	engine, err := New(WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(`var a: array of Integer := [1, 2]; PrintLn(a.Map(IntToStr).Join(','));`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Run(program); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "1,2\n" {
		t.Fatalf("output %q", got)
	}
}

func TestEngine_ExternalNestedArrayContext(t *testing.T) {
	engine, err := New(WithTypeCheck(false))
	if err != nil {
		t.Fatal(err)
	}
	called := false
	if err := engine.RegisterFunction("Inspect", func(values [][]int64) {
		called = true
		if len(values) != 2 || len(values[0]) != 0 || len(values[1]) != 1 || values[1][0] != 7 {
			t.Errorf("unexpected values: %v", values)
		}
	}); err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(`Inspect([[], [7]]);`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Run(program); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("external function was not called")
	}
}

func TestEngine_AliasedCallableMembers(t *testing.T) {
	var output bytes.Buffer
	engine, err := New(WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(`
		type TFunc = function(s: String): String;
		type TRecord = record Fn: TFunc; end;
		function Echo(s: String): String; begin Result := s; end;
		var r: TRecord;
		r.Fn := Echo;
		PrintLn(r.Fn('record'));
		type TEnum = (First, Second);
		type TAlias = TEnum;
		PrintLn(Ord(TAlias.Second));
		PrintLn(Ord(High(TAlias)));
	`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Run(program); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "record\n1\n1\n" {
		t.Fatalf("output %q", got)
	}
}

func TestEngine_ImplicitCallAndPointerContexts(t *testing.T) {
	cases := []struct{ name, source, output string }{
		{"implicit no arguments", `procedure Emit; begin PrintLn('called'); end; Emit;`, "called\n"},
		{"implicit default arguments", `procedure Emit(values: array of Integer = []); begin PrintLn(values.Length); end; Emit; Emit([1]);`, "0\n1\n"},
		{"explicit pointer context", `type TProc = procedure; procedure Emit; begin PrintLn('called'); end; var callback: TProc := Emit; PrintLn('bound'); callback();`, "bound\ncalled\n"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(test.source)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := engine.Run(program); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != test.output {
				t.Fatalf("output %q, want %q", got, test.output)
			}
		})
	}
}

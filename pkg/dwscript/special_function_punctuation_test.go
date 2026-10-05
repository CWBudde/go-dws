package dwscript

import (
	"bytes"
	"testing"
)

func TestSpecialFunctionPunctuation_ExecutionControls(t *testing.T) {
	tests := []struct {
		name, source, want string
	}{
		{"explicit calls", "var I := 2; Inc(I); Dec(I); Assert(I = 2); PrintLn(Length('abc')); PrintLn(Ord('a')); PrintLn(Succ(I)); PrintLn(Pred(I));", "3\n97\n3\n1\n"},
		{"lexical variable", "var Low := 3; PrintLn(Low);", "3\n"},
		{"lexical routine", "function Low: Integer; begin Result := 3; end; PrintLn(Low);", "3\n"},
		{"shadowed assignment", "var Length := 3; Length := 4; Length += 1; PrintLn(Length);", "5\n"},
		{"parameter", "procedure P(Ord: Integer); begin PrintLn(Ord); end; P(5);", "5\n"},
		{"shadowed routine address", "type F = function(X: Integer): Integer; function Length(X: Integer): Integer; begin Result := X+1; end; var P: F := @Length; PrintLn(P(7));", "8\n"},
		{"implicit class member", "type T = class function Length: Integer; begin Result := 7; end; function Test: Integer; begin Result := Length; end; end; var O := T.Create; PrintLn(O.Test()); PrintLn(O.Length());", "7\n7\n"},
		{"implicit method address", "type F = function: Integer of object; type T = class function Length: Integer; begin Result := 7; end; function Test: Integer; begin var P: F := @Length; var Q := @Length; Result := P() + Q(); end; end; PrintLn(T.Create.Test());", "14\n"},
		{"implicit class method address", "type T = class class function Ord: Integer; begin Result := 5; end; class function Test: Integer; begin var P := @Ord; Result := P(); end; end; PrintLn(T.Test());", "5\n"},
		{"qualified method address", "type F = function(X: Integer): Integer of object; type T = class function Length(X: Integer): Integer; begin Result := X+1; end; end; var O := T.Create; var P: F := @O.Length; PrintLn(P(7));", "8\n"},
		{"helper body", "type H = helper for Integer function Succ: Integer; begin Result := Self+1; end; function Test: Integer; begin Result := Succ; end; end; PrintLn(7.Test());", "8\n"},
		{"ordinary references", "var P := @Abs; PrintLn(P(-3)); var F := IntToStr; PrintLn(F(4));", "3\n4\n"},
		{"shadowed case range", "function Length: Integer; begin Result := 1; end; case 1 of Length..2: PrintLn(1); end;", "1\n"},
		{"shadowed array range", "function Low: Integer; begin Result := 1; end; PrintLn(1 in [Low..2]);", "True\n"},
		{"Default lookup", "function Default: Integer; begin Result := 7; end; PrintLn(Default);", "7\n"},
		{"Default intrinsic and namespace", "PrintLn(Default(Integer)); Default.PrintLn('after');", "0\nafter\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := engine.Eval(tt.source); err != nil {
				t.Fatalf("execution failed: %v", err)
			}
			if got := output.String(); got != tt.want {
				t.Fatalf("output = %q, want %q", got, tt.want)
			}
		})
	}
}

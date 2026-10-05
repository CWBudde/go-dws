package frontend

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// These expected anchors come from the pinned compiler's ReadSpecialFunction,
// ReadName and ReadAt, not from an executed Pascal oracle.
func TestCompile_SpecialFunctionPunctuation(t *testing.T) {
	names := []struct {
		canonical, lower string
		semicolon        int
	}{
		{"Assert", "assert", 14},
		{"Assigned", "assigned", 16},
		{"High", "high", 12},
		{"Length", "length", 14},
		{"Low", "low", 11},
		{"Ord", "ord", 11},
		{"SizeOf", "sizeof", 14},
		{"Defined", "defined", 15},
		{"Declared", "declared", 16},
		{"Inc", "inc", 11},
		{"Dec", "dec", 11},
		{"Succ", "succ", 12},
		{"Pred", "pred", 12},
		{"Include", "include", 15},
		{"Exclude", "exclude", 15},
		{"Swap", "swap", 12},
		{"ConditionalDefined", "conditionaldefined", 26},
	}
	contexts := []struct {
		name, prefix       string
		nameColumn, offset int
	}{
		{"value", "var i:=", 8, 0},
		{"address", "var i:=@", 9, 1},
		{"statement", "", 1, -7},
	}
	for _, name := range names {
		for _, context := range contexts {
			for _, spelling := range []string{name.canonical, name.lower} {
				t.Run(name.canonical+"/"+context.name+"/"+spelling, func(t *testing.T) {
					result := Compile(context.prefix+spelling+";", "<test>", semantic.HintsLevelPedantic)
					want := fmt.Sprintf(`Syntax Error: "(" expected [line: 1, column: %d]`, name.semicolon+context.offset)
					if spelling != name.canonical {
						want = fmt.Sprintf("Hint: %q does not match case of declaration (%q) [line: 1, column: %d]\n", spelling, name.canonical, context.nameColumn) + want
					}
					if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
						t.Fatalf("got %q; want %q", got, want)
					}
				})
			}
		}
	}
}

func TestCompile_SpecialFunctionPunctuationRecovery(t *testing.T) {
	tests := []struct {
		name, source, want string
	}{
		{"comment lookahead", "var i:=Length {note};", `Syntax Error: "(" expected [line: 1, column: 21]`},
		{"newline lookahead", "var i:=Length\n;", `Syntax Error: "(" expected [line: 2, column: 1]`},
		{"EOF uses last real token", "var i:=Length", `Syntax Error: "(" expected [line: 1, column: 8]`},
		{"comment at EOF", "var i:=Length {note}", `Syntax Error: "(" expected [line: 1, column: 8]`},
		{"address comment lookahead", "var P := @Length {note};", `Syntax Error: "(" expected [line: 1, column: 24]`},
		{"address newline lookahead", "var P := @Length\n;", `Syntax Error: "(" expected [line: 2, column: 1]`},
		{"address EOF", "var P := @Length", `Syntax Error: "(" expected [line: 1, column: 11]`},
		{"typed initializer", "var X: Integer := Length;", `Syntax Error: "(" expected [line: 1, column: 25]`},
		{"constant initializer", "const X = Length; PrintLn(X);", `Syntax Error: "(" expected [line: 1, column: 17]`},
		{"grouped statement", "(Length);", `Syntax Error: "(" expected [line: 1, column: 8]`},
		{"nested grouping", "((Length));", `Syntax Error: "(" expected [line: 1, column: 9]`},
		{"grouped callee stops before arguments", "(Length)(Unknown);", `Syntax Error: "(" expected [line: 1, column: 8]`},
		{"grouped address child stops", "var P := @(Length);", `Syntax Error: "(" expected [line: 1, column: 18]`},
		{"assignment target stops before RHS", "Length := Unknown;", `Syntax Error: "(" expected [line: 1, column: 8]`},
		{"compound assignment target", "Length += Unknown;", `Syntax Error: "(" expected [line: 1, column: 8]`},
		{"indexed address stops before index", "var X := @Length[Unknown];", `Syntax Error: "(" expected [line: 1, column: 17]`},
		{"grouped indexed address", "var X := @(Length)[Unknown];", `Syntax Error: "(" expected [line: 1, column: 18]`},
		{"scalar expected type precedes child stop", "var X: Integer := @Length;", "Syntax Error: unexpected \"@\" [line: 1, column: 19]\nSyntax Error: \"(\" expected [line: 1, column: 26]"},
		{"binary operand stops before right child", "Length + Unknown;", `Syntax Error: "(" expected [line: 1, column: 8]`},
		{"enclosing call", "PrintLn(Length, Unknown);", `Syntax Error: "(" expected [line: 1, column: 15]`},
		{"array element", "var X := [Length, Unknown];", `Syntax Error: "(" expected [line: 1, column: 17]`},
		{"case range cannot rewrite a bare intrinsic", "case 1 of Length..2: PrintLn(1); end;", `Syntax Error: "(" expected [line: 1, column: 17]`},
		{"callback context still needs parentheses", "type F = function(V: Variant): Integer;\nvar P: F := Length;", `Syntax Error: "(" expected [line: 2, column: 19]`},
		{"later errors suppressed", "var X := Length;\nUnknown;", `Syntax Error: "(" expected [line: 1, column: 16]`},
		{"address stop suppresses later errors", "var X := @Low;\nUnknown;", `Syntax Error: "(" expected [line: 1, column: 14]`},
		{"earlier error retained", "var I: Integer; I := 'x';\nvar X := Length;", "Syntax Error: Incompatible types: Cannot assign \"String\" to \"Integer\" [line: 1, column: 22]\nSyntax Error: \"(\" expected [line: 2, column: 16]"},
		{"Default has no intrinsic hint or parenthesis requirement", "var X := default;", `Syntax Error: Unknown name "default" [line: 1, column: 10]`},
		{"instance method address from class method", "type T = class function Length: Integer; begin Result := 7; end; class procedure Test; begin var P := @Length; end; end;", `Syntax Error: Class method or constructor expected [line: 1, column: 104]`},
		{"field address from class method", "type T = class High: Integer; class procedure Test; begin var P := @High; end; end;", `Syntax Error: Object reference needed to read/write an object field [line: 1, column: 69]`},
		{"class constant address", "type T = class const Length = 3; class procedure Test; begin var P := @Length; end; end;", `Syntax Error: unexpected "@" [line: 1, column: 71]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelDisabled)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCompile_SpecialFunctionPunctuationControls(t *testing.T) {
	for _, source := range []string{
		"var X := Length('abc'); var Y := Ord('a'); PrintLn(X, Y);",
		"var I := 1; Inc(I); Dec(I); Assert(I = 1); PrintLn(Succ(I), Pred(I));",
		"var X := Declared('Integer'); var Y := ConditionalDefined('X'); PrintLn(X, Y);",
		"var Low: Integer := 3; PrintLn(Low);",
		"function Low: Integer; begin Result := 3; end; PrintLn(Low);",
		"procedure P(Ord: Integer); begin PrintLn(Ord); end; P(5);",
		"type R = record function Ord: Integer; begin Result := 8; end; function Test: Integer; begin Result := Ord; end; end; var V: R; PrintLn(V.Test());",
		"type T = class function Length: Integer; begin Result := 7; end; function Test: Integer; begin Result := Length; end; end; var O := T.Create; PrintLn(O.Test());",
		"type H = helper for Integer function Succ: Integer; begin Result := Self+1; end; function Test: Integer; begin Result := Succ; end; end; PrintLn(7.Test());",
		"type T = class function Length: Integer; begin Result := 7; end; procedure Test; begin var P := @Length; end; end;",
		"type F = function: Integer of object; type T = class function Length: Integer; begin Result := 7; end; procedure Test; begin var P: F := @Length; end; end;",
		"type F = function(X: Integer): Integer; function Length(X: Integer): Integer; begin Result := X+1; end; var P: F := @Length; PrintLn(P(7));",
		"var P := @Abs; PrintLn(P(-3)); var F := IntToStr; PrintLn(F(4));",
		"function Default: Integer; begin Result := 7; end; PrintLn(Default);",
		"var X := Default(Integer); Default.PrintLn(X);",
		"function Length: Integer; begin Result := 1; end; case 1 of Length..2: PrintLn(1); end;",
		"function Low: Integer; begin Result := 1; end; PrintLn(1 in [Low..2]);",
		"var Length := 3; Length := 4; Length += 1; PrintLn(Length);",
		"DebugBreak; DebugBreak();",
	} {
		t.Run(source, func(t *testing.T) {
			result := Compile(source, "<test>", semantic.HintsLevelDisabled)
			if got := result.DiagnosticStrings(); len(got) != 0 {
				t.Fatalf("unexpected diagnostics: %q", got)
			}
		})
	}
}

func TestCompile_SpecialFunctionPunctuationPedanticControls(t *testing.T) {
	tests := []struct{ source, want string }{
		{"length := Unknown;", "Hint: \"length\" does not match case of declaration (\"Length\") [line: 1, column: 1]\nSyntax Error: \"(\" expected [line: 1, column: 8]"},
		{"var X := default;", `Syntax Error: Unknown name "default" [line: 1, column: 10]`},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

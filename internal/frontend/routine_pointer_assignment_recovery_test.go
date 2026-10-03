package frontend

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_RoutinePointerAssignmentRecoveryFixture(t *testing.T) {
	result := Compile(fixtureSource(t, "func_ptr1.pas"), "func_ptr1.pas", semantic.HintsLevelPedantic)
	if got, want := result.DiagnosticStrings(), fixtureExpectation(t, "func_ptr1.txt"); !reflect.DeepEqual(got, want) {
		t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestCompile_RoutinePointerAssignmentRecovery(t *testing.T) {
	tests := []struct {
		name, declarations, rhs string
		messages                []string
	}{
		{"bare procedure integer", "procedure Work(x: Integer); begin end;", "Work", []string{"More arguments expected", "Assignment's right-side-argument has no return type"}},
		{"bare procedure float", "procedure Work(x: Float); begin end;", "Work", []string{"More arguments expected", "Assignment's right-side-argument has no return type"}},
		{"explicit procedure missing argument", "procedure Work(x: Integer); begin end;", "Work()", []string{"More arguments expected", "Assignment's right-side-argument has no return type"}},
		{"bare string function", "function Work: String; begin Result := ''; end;", "Work", []string{"Incompatible operands", `Incompatible types: Cannot assign "String" to "procedure TProc"`}},
		{"explicit string function missing argument", "function Work(x: Integer): String; begin Result := ''; end;", "Work()", []string{"More arguments expected", "Incompatible operands", `Incompatible types: Cannot assign "String" to "procedure TProc"`}},
		{"bare procedure pointer missing argument", "var Work: procedure(x: Integer);", "Work", []string{"More arguments expected", "Assignment's right-side-argument has no return type"}},
		{"bare string pointer", "var Work: function: String;", "Work", []string{"Incompatible operands", `Incompatible types: Cannot assign "String" to "procedure TProc"`}},
		{"explicit procedure pointer", "var Work: TProc;", "Work()", []string{"Assignment's right-side-argument has no return type"}},
		{"unknown callee", "", "Missing()", []string{`Unknown name "Missing"`}},
		{"nested bad arity", "function Work(x: Integer; y: Integer): String; begin Result := ''; end; function Nested(x: Integer): Integer; begin Result := x; end;", "Work(Nested())", []string{"More arguments expected"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := "type TProc = procedure;\n" + tt.declarations + "\nvar p: TProc;\np := " + tt.rhs + ";"
			got := Compile(source, "assignment.pas", semantic.HintsLevelDisabled).DiagnosticStrings()
			want := make([]string, len(tt.messages))
			for i, message := range tt.messages {
				column := 6
				if message == "Assignment's right-side-argument has no return type" || message == "Incompatible operands" {
					column = 3
				}
				if tt.name == "unknown callee" {
					column = 14
				}
				if tt.name == "nested bad arity" {
					column = 11
				}
				want[i] = fmt.Sprintf("Syntax Error: %s [line: 4, column: %d]", message, column)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
			}
		})
	}
}

func TestCompile_ValuelessAssignmentTargets(t *testing.T) {
	tests := []struct{ name, declarations, statement string }{
		{"variant", "var v: Variant;", "v := Work();"},
		{"bare variant", "var v: Variant;", "v := Work;"},
		{"compound variant", "var v: Variant;", "v += Work();"},
		{"array index", "var values: array[0..0] of Variant;", "values[0] := Work();"},
		{"record member", "type TRec = record Value: Variant; end; var r: TRec;", "r.Value := Work();"},
		{"class field", "type TObj = class Value: Variant; end; var obj: TObj;", "obj.Value := Work();"},
		{"field-backed property", "type TObj = class FValue: Variant; property Value: Variant read FValue write FValue; end; var obj: TObj;", "obj.Value := Work();"},
		{"method-backed property", "type TObj = class procedure SetValue(v: Variant); begin end; property Value: Variant write SetValue; end; var obj: TObj;", "obj.Value := Work();"},
		{"function result", "", "function Test: Variant; begin Result := Work(); end;"},
		{"function name result", "", "function Test: Variant; begin Test := Work(); end;"},
		{"implicit class field", "", "type TObj = class Value: Variant; procedure Test; begin Value := Work(); end; end;"},
		{"implicit field property", "", "type TObj = class FValue: Variant; property Value: Variant read FValue write FValue; procedure Test; begin Value := Work(); end; end;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := "procedure Work; begin end;\n" + tt.declarations + "\n" + tt.statement
			column := strings.Index(tt.statement, ":=") + 1
			if column == 0 {
				column = strings.Index(tt.statement, "+=") + 1
			}
			want := []string{fmt.Sprintf("Syntax Error: Assignment's right-side-argument has no return type [line: 3, column: %d]", column)}
			got := Compile(source, "valueless.pas", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
			}
		})
	}
}

// Upstream reads simple dynamic-array stores through its element conversion
// path, while static-array stores use the ordinary assignment reader.
func TestCompile_DynamicSlotValuelessAssignment(t *testing.T) {
	for _, tt := range []struct{ name, declaration, element, rhs, sourceType string }{
		{"declared procedure integer", "procedure Work; begin end;", "Integer", "Work()", "void"},
		{"declared procedure variant", "procedure Work; begin end;", "Variant", "Work()", "void"},
		{"pointer procedure variant", "var Work: procedure;", "Variant", "Work()", "void"},
		{"string function pointer slot", "type TProc = procedure; function Work: String; begin Result := ''; end;", "TProc", "Work()", "String"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := tt.declaration + "\nvar values: array of " + tt.element + ";\nvalues[0] := " + tt.rhs + ";"
			destination := tt.element
			if destination == "TProc" {
				destination = "procedure TProc"
			}
			want := []string{fmt.Sprintf("Syntax Error: Incompatible types: Cannot assign %q to %q [line: 3, column: 11]", tt.sourceType, destination)}
			got := Compile(source, "dynamic_slot.pas", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
			}
		})
	}
}

func TestCompile_CompoundValuelessCallRecovery(t *testing.T) {
	for _, tt := range []struct {
		name, declaration, rhs string
		missing                bool
	}{
		{"bare procedure", "procedure Work; begin end;", "Work", false},
		{"bare missing arguments", "procedure Work(x: Integer); begin end;", "Work", true},
		{"explicit missing arguments", "procedure Work(x: Integer); begin end;", "Work()", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := tt.declaration + "\nvar v: Variant;\nv += " + tt.rhs + ";"
			want := []string{}
			if tt.missing {
				want = append(want, "Syntax Error: More arguments expected [line: 3, column: 6]")
			}
			want = append(want, "Syntax Error: Assignment's right-side-argument has no return type [line: 3, column: 3]")
			got := Compile(source, "compound_void.pas", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
			}
		})
	}
}

func TestCompile_CallableDynamicSlotValuelessAssignment(t *testing.T) {
	for _, tt := range []struct{ name, declaration, work, destination string }{
		{"valueless", "type TArray = array of Integer;", "procedure Work; begin end;", "Integer"},
		{"pointer mismatch", "type TProc = procedure; type TArray = array of TProc;", "function Work: String; begin Result := ''; end;", "procedure TProc"},
	} {
		for _, base := range []string{"Values", "Values()"} {
			t.Run(tt.name+"/"+base, func(t *testing.T) {
				source := tt.declaration + " function Values: TArray; begin end;\n" + tt.work + "\n" + base + "[0] := Work();"
				column := len(base) + 5
				sourceType := "void"
				if tt.name == "pointer mismatch" {
					sourceType = "String"
				}
				want := []string{fmt.Sprintf("Syntax Error: Incompatible types: Cannot assign %q to %q [line: 3, column: %d]", sourceType, tt.destination, column)}
				got := Compile(source, "callable_dynamic_slot.pas", semantic.HintsLevelDisabled).DiagnosticStrings()
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
				}
			})
		}
	}
}

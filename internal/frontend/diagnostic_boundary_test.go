package frontend

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestRestoreStatementWarningOrder(t *testing.T) {
	parse := func(message string, line, col int) Diagnostic {
		return Diagnostic{Message: message, Phase: PhaseParsing, Line: line, Column: col}
	}
	semantic := func(message string, line, col int, severity Severity) Diagnostic {
		return Diagnostic{Rendered: message, Phase: PhaseSemantic, Line: line, Column: col, Severity: severity}
	}
	prior := parse("earlier parser error", 1, 1)
	later := parse("later parser error", 3, 12)
	typeError := semantic("Boolean expected", 3, 4, SeverityError)
	warning := semantic("Warning: Constant condition [line: 3, column: 4]", 3, 4, SeverityWarning)
	messageError := semantic("String expected", 3, 4, SeverityError)
	after := semantic("later semantic error", 4, 1, SeverityError)
	got := []Diagnostic{prior, later, typeError, warning, messageError, after}
	restoreStatementWarningOrder(got)
	want := []Diagnostic{prior, typeError, warning, messageError, later, after}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
	// Repeated normalization must preserve both streams, including equal anchors.
	restoreStatementWarningOrder(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("second pass changed order: %+v", got)
	}
}

func TestRestoreStatementWarningOrder_UnreachableDoesNotMoveTargetErrors(t *testing.T) {
	parser := Diagnostic{Message: "parser error", Phase: PhaseParsing, Line: 3, Column: 4}
	warning := Diagnostic{Rendered: "Warning: Unreachable code [line: 3, column: 4]", Phase: PhaseSemantic, Severity: SeverityWarning, Line: 3, Column: 4}
	targetError := Diagnostic{Message: "target error", Phase: PhaseSemantic, Line: 3, Column: 4}
	got := []Diagnostic{parser, warning, targetError}
	restoreStatementWarningOrder(got)
	if want := []Diagnostic{warning, parser, targetError}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
}

func TestCompile_ContractMessageBeforeParserDelimiter(t *testing.T) {
	source := "procedure P; require 1:2); begin end;"
	want := []string{
		"Syntax Error: Boolean expected [line: 1, column: 22]",
		"Warning: Constant condition [line: 1, column: 22]",
		"Syntax Error: String expected [line: 1, column: 22]",
		"Syntax Error: \";\" expected [line: 1, column: 25]",
	}
	if got := Compile(source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings(); !reflect.DeepEqual(got, want) {
		t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
	}
}

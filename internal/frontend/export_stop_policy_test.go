package frontend

import (
	"reflect"
	"testing"
)

func TestDropDiagnosticsAfterStop_ExportSemanticBoundary(t *testing.T) {
	parserStop := Diagnostic{Phase: PhaseParsing, Line: 2, Column: 14, Stop: true}
	semanticStop := Diagnostic{Phase: PhaseSemantic, Line: 2, Column: 14, Stop: true}
	parserError := Diagnostic{Phase: PhaseParsing, Line: 2, Column: 14}
	semanticError := Diagnostic{Phase: PhaseSemantic, Line: 2, Column: 14}
	earlierParser := Diagnostic{Phase: PhaseParsing, Line: 1, Column: 14, Stop: true}
	earlierSemantic := Diagnostic{Phase: PhaseSemantic, Line: 1, Column: 14, Stop: true}
	for _, tt := range []struct {
		name        string
		input, want []Diagnostic
	}{
		{"parser first tie", []Diagnostic{parserError, parserStop, semanticStop}, []Diagnostic{semanticStop}},
		{"semantic first tie", []Diagnostic{semanticStop, parserError, parserStop}, []Diagnostic{semanticStop}},
		{"earlier parser wins", []Diagnostic{semanticStop, earlierParser, parserStop}, []Diagnostic{earlierParser, parserStop}},
		{"earlier semantic wins", []Diagnostic{parserStop, earlierSemantic, semanticStop}, []Diagnostic{earlierSemantic, semanticStop}},
		{"semantic error at boundary retained", []Diagnostic{semanticError, parserError, parserStop, semanticStop}, []Diagnostic{semanticError, semanticStop}},
		{"parser only unchanged", []Diagnostic{parserError, parserStop}, []Diagnostic{parserError, parserStop}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := dropDiagnosticsAfterStop(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v; want %+v", got, tt.want)
			}
		})
	}
}

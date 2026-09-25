package frontend

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// These fixtures exercise routine metadata through the same compile pipeline
// used by the CLI, including signature compatibility and diagnostic positions.
func TestCompile_RoutinePointerDiagnostics(t *testing.T) {
	for _, name := range []string{"func_ptr3", "func_ptr4", "func_ptr5", "func_ptr_mismatch", "func_ptr_var_param"} {
		t.Run(name, func(t *testing.T) {
			result := Compile(fixtureSource(t, name+".pas"), name+".pas", semantic.HintsLevelPedantic)
			want := fixtureExpectation(t, name+".txt")
			if got := result.DiagnosticStrings(); !reflect.DeepEqual(got, want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, want)
			}
		})
	}
}

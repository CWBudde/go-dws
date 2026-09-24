package frontend

import "testing"

// Pin complete fixture output, including already passing F2 diagnostics, so
// recovery cannot trade one missing diagnostic for an unrelated cascade.
func TestCompile_F2ArrayDiagnostics(t *testing.T) {
	for _, name := range []string{
		"array_not_array", "array_index_extra", "array_bounds", "array_index_int",
		"array_index_bool", "array_index_enum", "array_bounds_not_ordinal",
		"array_error1", "array_error2", "array_error3", "array_error4", "array_error5", "array_error7",
		"array_index_bracket_missing", "array_index_bracket_missing1", "array_index_bracket_missing2",
		"array_method1", "array_range2", "case_range_mismatch", "case_error5", "in_operator8",
	} {
		t.Run(name, func(t *testing.T) {
			assertDiagnostics(t, fixtureSource(t, name+".pas"), name+".pas", fixtureExpectation(t, name+".txt"))
		})
	}
}

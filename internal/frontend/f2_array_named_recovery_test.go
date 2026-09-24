package frontend

import "testing"

func TestCompile_NamedArrayBoundRecovery(t *testing.T) {
	assertDiagnostics(t, "type TProc = procedure;\ntype TArray = array[1..   TProc;\nvar b: Integer;", "named-array-recovery.pas", []string{
		`Syntax Error: Function expected [line: 2, column: 27]`,
		`Syntax Error: Bound isn't of an ordinal type [line: 2, column: 23]`,
		`Syntax Error: "]" expected [line: 2, column: 32]`,
	})
}

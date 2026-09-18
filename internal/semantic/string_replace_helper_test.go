package semantic

import "testing"

func TestStringReplaceHelper_Signature(t *testing.T) {
	for _, source := range []string{
		`var result: String := 'aba'.Replace('a', 'x');`,
		`var s := 'aba'; s.Replace('a', 'x');`,
	} {
		t.Run(source, func(t *testing.T) { expectNoErrors(t, source) })
	}
	for _, source := range []string{
		`'aba'.Replace;`,
		`'aba'.Replace();`,
		`'aba'.Replace('a');`,
		`'aba'.Replace('a', 'x', 1);`,
		`'aba'.Replace(1, 'x');`,
		`'aba'.Replace('a', True);`,
		`var result: Integer := 'aba'.Replace('a', 'x');`,
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := analyzeSource(t, source); err == nil {
				t.Fatal("expected signature error")
			}
		})
	}
}

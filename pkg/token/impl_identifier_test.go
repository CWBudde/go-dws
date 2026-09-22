package token

import "testing"

func TestImplIdentifier(t *testing.T) {
	for _, name := range []string{"impl", "Impl", "IMPL", "iMpL"} {
		t.Run(name, func(t *testing.T) {
			if got := LookupIdent(name); got != IDENT {
				t.Errorf("LookupIdent(%q) = %v, want IDENT", name, got)
			}
			if IsKeyword(name) {
				t.Errorf("IsKeyword(%q) = true, want false", name)
			}
		})
	}
}

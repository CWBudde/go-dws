package semantic

import "testing"

// The indexed compatibility probe must not add a receiver analysis when the
// member turns out to be an ordinary method. The interface-property probe also
// hands its analyzed receiver to the fallback, retaining one diagnostic per read.
func TestIndexedCompatibilityReceiverAnalyzedOnce(t *testing.T) {
	const prefix = "type TInner = class function Items: array of Integer; begin Result := [1]; end; end;\ntype TOuter = class FInner: TInner; property Inner: TInner read FInner; deprecated 'old'; end;\nvar Obj := new TOuter;\n"
	tests := []struct{ name, body, want string }{
		{"deprecated receiver", "var A := Obj.Inner.Items()[0];", `Warning: "Inner" has been deprecated: old [line: 4, column: 14]`},
		{"unknown receiver", "var A := Obj.Missing.Items()[0];", `There is no accessible member with name "Missing" for type TOuter at 4:14`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, _ := analyzeSource(t, prefix+tt.body)
			errs := analyzer.Errors()
			if len(errs) != 1 || errs[0] != tt.want {
				t.Fatalf("receiver diagnostics = %q, want [%q]", errs, tt.want)
			}
		})
	}
}

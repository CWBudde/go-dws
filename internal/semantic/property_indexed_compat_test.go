package semantic

import "testing"

// The indexed compatibility probe must not add a receiver analysis when the
// member turns out to be an ordinary method. The remaining duplicate is the
// interface-property probe re-analyzing the indexed call, which predates it.
// An unknown receiver member is a compiler stop, so it is reported once.
func TestIndexedCompatibilityReceiverAnalyzedOnce(t *testing.T) {
	const prefix = "type TInner = class function Items: array of Integer; begin Result := [1]; end; end;\ntype TOuter = class FInner: TInner; property Inner: TInner read FInner; deprecated 'old'; end;\nvar Obj := new TOuter;\n"
	tests := []struct {
		name, body string
		want       int
	}{
		{"deprecated receiver", "var A := Obj.Inner.Items()[0];", 2},
		{"unknown receiver", "var A := Obj.Missing.Items()[0];", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, _ := analyzeSource(t, prefix+tt.body)
			errs := analyzer.Errors()
			if len(errs) != tt.want {
				t.Fatalf("want the receiver diagnostic %d time(s), got %q", tt.want, errs)
			}
		})
	}
}

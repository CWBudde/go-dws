package interp

import "testing"

// A marked singleton still accepts matching arguments and numeric widening.
// Runtime dispatch among multiple declarations is a separately measured task.
func TestHelperOverloadCandidates_CompiledSingleton(t *testing.T) {
	for _, target := range []struct{ name, declaration, setup string }{
		{"primitive", "Integer", "var item: T := 5;"},
		{"record", "record x: Integer; end", "var item: T;"},
		{"class", "class end", "var item := T.Create;"},
		{"interface", "interface end", "type C = class(TObject, T) end; var item: T := C.Create;"},
	} {
		t.Run(target.name, func(t *testing.T) {
			helperTarget := "T"
			if target.name == "primitive" {
				helperTarget = "Integer"
			}
			source := "type T = " + target.declaration + ";\n" +
				"type H = helper for " + helperTarget + "\n" +
				"function Take(v: Float): Float; overload; begin Result := v + 1; end;\n" +
				"function Run: Float; begin Result := Take(2); end;\nend;\n" +
				target.setup + "\nPrintLn(item.Take(2)); PrintLn(H.Take(item, 2)); PrintLn(item.Run());"
			compileAndRunWithHelperTransfer(t, source, "helper_overload_singleton.dws", "3\n3\n3\n")
		})
	}
}

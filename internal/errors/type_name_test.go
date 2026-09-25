package errors

import "testing"

func TestSimplifyTypeName(t *testing.T) {
	for _, tt := range []struct{ input, want string }{
		{"TChild(TParent)", "TChild"},
		{"array of TChild(TParent)", "array of TChild"},
		{"array[0..2] of TChild(TParent)", "array [0..2] of TChild"},
		{"Void", "void"},
		{"function IntToHex(Integer, Integer): String", "function IntToHex(Integer, Integer): String"},
		{"class function Convert(Integer): String", "class function Convert(Integer): String"},
		{"procedure Test(const String)", "procedure Test(const String)"},
		{"procedure (String)", "procedure (String)"},
		{"procedure ", "procedure "},
		{"constructor Create(Integer)", "constructor Create(Integer)"},
		{"destructor Destroy", "destructor Destroy"},
		{"function(Integer): String", "function(Integer): String"},
		{"procedure(Integer) of object", "procedure(Integer) of object"},
		{"array of procedure (String)", "array of procedure (String)"},
		{"array[0..2] of procedure (String)", "array [0..2] of procedure (String)"},
		{"Functionality(TObject)", "Functionality"},
		{"ProcedureList(TObject)", "ProcedureList"},
	} {
		t.Run(tt.input, func(t *testing.T) {
			if got := SimplifyTypeName(tt.input); got != tt.want {
				t.Errorf("SimplifyTypeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

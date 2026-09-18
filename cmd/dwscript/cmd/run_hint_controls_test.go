package cmd

import "testing"

func TestRun_SourceHintControls(t *testing.T) {
	const source = "var Value := 7;\n{$HINTS OFF}\nPrintLn(value);\n{$HINTS ON}\nPrintLn(value);\n{$HINTS NORMAL}\nPrintLn(value);\n{$HINTS PEDANTIC}\nPrintLn(value);\n"
	const restoredHint = "Hint: \"value\" does not match case of declaration (\"Value\") [line: 5, column: 9]\n"
	const explicitHint = "Hint: \"value\" does not match case of declaration (\"Value\") [line: 9, column: 9]\n"
	for _, tt := range []struct {
		name    string
		level   string
		want    string
		compile bool
	}{
		{"pedantic envelope", "pedantic", "Errors >>>>\n" + restoredHint + explicitHint + "Result >>>>\n7\n7\n7\n7\n", false},
		{"normal envelope", "normal", "Errors >>>>\n" + explicitHint + "Result >>>>\n7\n7\n7\n7\n", false},
		{"pedantic compile only", "pedantic", restoredHint + explicitHint, true},
		{"normal compile only", "normal", explicitHint, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := captureRun(t, source, nil, func() {
				hintsLevel = tt.level
				compileOnly = tt.compile
				testEnvelope = true
			})
			if err != nil {
				t.Fatalf("run failed: %v\n%s", err, out)
			}
			if out != tt.want {
				t.Fatalf("got %q\nwant %q", out, tt.want)
			}
		})
	}
}

package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Inactive conditionals with skipped tokens stop inside ReadUntilEndOrElseSwitch;
// immediate EOF instead stops in its caller. Their source anchors differ upstream.
func TestRun_ConditionalEOFAnchors(t *testing.T) {
	tests := []struct {
		name, source, want string
	}{
		{"immediate EOF", "{$ifdef test}", "Syntax Error: Unbalanced conditional directive [line: 1, column: 3]"},
		{"skipped statement", "{$ifdef test}\nPrintLn('hello');", "Syntax Error: Unbalanced conditional directive [line: 1, column: 9]"},
		{"whitespace only", "{$ifdef test}\n \t\n", "Syntax Error: Unbalanced conditional directive [line: 1, column: 3]"},
		{"comments only", "{$ifdef test}\n{ comment } (* another *) // last\n", "Syntax Error: Unbalanced conditional directive [line: 1, column: 3]"},
		{"directive text in skipped string", "{$ifdef test}\nPrintLn('{$endif}');", "Syntax Error: Unbalanced conditional directive [line: 1, column: 9]"},
		{"unknown skipped switch then body", "{$ifdef test}\n{$NOP}\nPrintLn('hello');", "Syntax Error: Unbalanced conditional directive [line: 1, column: 9]"},
		{"active branch", "{$define test}\n{$ifdef test}\nPrintLn('hello');", "Syntax Error: Unbalanced conditional directive [line: 2, column: 3]"},
		{"active ELSE retains opening IF", "{$if False}\n{$else}\nPrintLn('hello');", "Syntax Error: Unbalanced conditional directive [line: 1, column: 3]"},
		{"nested skipped directive", "{$ifdef test}\n{$ifdef inner}\nPrintLn('hello');", "Syntax Error: Unbalanced conditional directive [line: 2, column: 3]"},
		{"closed nested directive then tokens", "{$ifdef test}\n{$ifdef inner}\n{$endif}\nPrintLn('hello');", "Syntax Error: Unbalanced conditional directive [line: 3, column: 3]"},
		{"nested directive followed by immediate EOF", "{$ifdef test}\n{$ifdef inner}", "Syntax Error: Unbalanced conditional directive [line: 1, column: 3]"},
		{"inactive ELSE", "{$if True}\n{$else}\nPrintLn('hello');", "Syntax Error: Unbalanced conditional directive [line: 2, column: 3]"},
		{"skipped EOF truncates enclosing block", "begin\n{$ifdef test}\nPrintLn('hello');", "Syntax Error: Unbalanced conditional directive [line: 2, column: 9]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := captureRun(t, tt.source, nil, func() { compileOnly = true; hintsLevel = "pedantic" })
			if err == nil || strings.TrimSpace(out) != tt.want {
				t.Fatalf("compile err=%v, diagnostics=%q; want %q", err, out, tt.want)
			}
		})
	}
}

func TestRun_ConditionalAnchorFixtures(t *testing.T) {
	for _, name := range []string{"conditionals2", "conditionals2.1", "conditionals_else3", "invalid_switch"} {
		t.Run(name, func(t *testing.T) {
			base := filepath.Join("../../../testdata/fixtures/FailureScripts", name)
			want, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			out, err := captureRun(t, "", []string{base + ".pas"}, func() { compileOnly = true; hintsLevel = "pedantic" })
			if err == nil || strings.TrimSpace(out) != strings.TrimSpace(string(want)) {
				t.Fatalf("compile err=%v, diagnostics=%q; want %q", err, out, want)
			}
		})
	}
}

func TestRun_IncludeMacroDiagnosticFixtures(t *testing.T) {
	for _, name := range []string{"include_incorrect", "include_expr"} {
		t.Run(name, func(t *testing.T) {
			base := filepath.Join("../../../testdata/fixtures/FailureScripts", name)
			want, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			out, err := captureRun(t, "", []string{base + ".pas"}, func() { compileOnly = true; hintsLevel = "pedantic" })
			if err == nil || strings.TrimSpace(out) != strings.TrimSpace(string(want)) {
				t.Fatalf("compile err=%v, diagnostics=%q; want %q", err, out, want)
			}
		})
	}
}

func TestRun_IncludeMacroAnchors(t *testing.T) {
	tests := []struct {
		name, source, want string
	}{
		{"unknown item", "PrintLn({$I %DUMMY%});", "Syntax Error: Include item \"DUMMY\" unknown [line: 1, column: 13]"},
		{"missing final percent", "PrintLn({$I %LINE});", "Syntax Error: Include item expected [line: 1, column: 13]"},
		{"missing item name", "PrintLn({$I %});", "Syntax Error: Include item expected [line: 1, column: 13]"},
		{"wrong final percent token remains current", "PrintLn({$I %LINE extra});", "Syntax Error: Include item expected [line: 1, column: 13]\nSyntax Error: \"}\" expected [line: 1, column: 19]"},
		{"reserved macro name remains current", "PrintLn({$I %END%});", "Syntax Error: Include item expected [line: 1, column: 13]\nSyntax Error: \"}\" expected [line: 1, column: 14]"},
		{"numeric macro name remains current", "PrintLn({$I %42%});", "Syntax Error: Include item expected [line: 1, column: 13]\nSyntax Error: \"}\" expected [line: 1, column: 14]"},
		{"missing brace at EOF", "PrintLn({$I %LINE%", "Syntax Error: \"}\" expected [line: 1, column: 18]"},
		{"missing brace after spaced argument", "PrintLn({$I   %LINE% \n", "Syntax Error: \"}\" expected [line: 1, column: 20]"},
		{"wrong token instead of brace", "PrintLn({$I %LINE% extra});", "Syntax Error: \"}\" expected [line: 1, column: 20]"},
		{"following token instead of brace", "PrintLn({$I %LINE%\nPrintLn('late');", "Syntax Error: \"}\" expected [line: 2, column: 1]"},
		{"truncates enclosing block", "begin\nPrintLn({$I %LINE%", "Syntax Error: \"}\" expected [line: 2, column: 18]"},
		{"keeps earlier semantic error", "PrintLn(Undeclared);\nPrintLn({$I %LINE%", "Syntax Error: Unknown name \"Undeclared\" [line: 1, column: 9]\nSyntax Error: \"}\" expected [line: 2, column: 18]"},
		{"earlier parser stop wins", "var x := ;\nPrintLn({$I %LINE%", "Syntax Error: Expression expected [line: 1, column: 10]"},
		{"missing filename in expression", "PrintLn({$I });", "Syntax Error: Name of include file expected [line: 1, column: 13]\nSyntax Error: Expression expected [line: 1, column: 13]"},
		{"missing standalone filename", "{$I }", "Syntax Error: Name of include file expected [line: 1, column: 5]"},
		{"inactive malformed macro", "{$ifdef absent}\nPrintLn({$I %DUMMY%});\n{$endif}", ""},
		{"inactive unterminated macro checks brace", "{$ifdef absent}\n{$I %LINE%", "Syntax Error: \"}\" expected [line: 2, column: 10]"},
		{"inactive unknown unterminated macro only checks brace", "{$ifdef absent}\n{$I %DUMMY%", "Syntax Error: \"}\" expected [line: 2, column: 11]"},
		{"adjacent switch name missing brace", "PrintLn({$I%LINE%", "Syntax Error: \"}\" expected [line: 1, column: 17]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := captureRun(t, tt.source, nil, func() { compileOnly = true; hintsLevel = "pedantic" })
			if (err != nil) != (tt.want != "") || strings.TrimSpace(out) != tt.want {
				t.Fatalf("compile err=%v, diagnostics=%q; want %q", err, out, tt.want)
			}
		})
	}
}

func TestRun_IncludeLineSubstitutions(t *testing.T) {
	const source = "PrintLn({$I %LINE%});\nPrintLn(100 + {$INCLUDE %LINENUM%});\nPrintLn({$i %line%} + 'x');\nPrintLn({$I%LINE%});\nPrintLn(100 + {$INCLUDE%LINENUM%});"
	out, err := captureRun(t, source, nil, func() { diagnosticsMode = "plain" })
	if err != nil || out != "1\n102\n3x\n4\n105\n" {
		t.Fatalf("run err=%v, output=%q; want line string and integer substitutions", err, out)
	}
}

func TestRun_IncludeTimeSubstitution(t *testing.T) {
	out, err := captureRun(t, "PrintLn({$I %TIME%});", nil, func() { diagnosticsMode = "plain" })
	if err != nil || !regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\n$`).MatchString(out) {
		t.Fatalf("run err=%v, output=%q; want a substituted hh:mm:ss string", err, out)
	}
}

func TestRun_IncludeDateAndTimestampSubstitutions(t *testing.T) {
	before := time.Now()
	out, err := captureRun(t, "PrintLn('date:' + {$I %DATE%});\nPrintLn({$I %TIMESTAMP%} + 1);", nil, func() { diagnosticsMode = "plain" })
	after := time.Now()
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if err != nil || len(lines) != 2 {
		t.Fatalf("run err=%v, output=%q; want calendar date string and Unix timestamp integer", err, out)
	}
	if lines[0] != "date:"+before.Format("2006-01-02") && lines[0] != "date:"+after.Format("2006-01-02") {
		t.Fatalf("date output=%q; want current local calendar date", lines[0])
	}
	stamp, err := strconv.ParseInt(lines[1], 10, 64)
	if err != nil || stamp < before.Unix()+1 || stamp > after.Unix()+1 {
		t.Fatalf("timestamp output=%q, err=%v; want current Unix seconds plus one", lines[1], err)
	}
}
